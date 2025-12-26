// go run main.go --account=ether_fifa.node-1 --method="" --contract_id=transfer --args="" --network=local --workers=80 --workload=transfer

// near --keyPath ~/.near/node-3/node_key.json view fifa.node-1 getCount
package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	// "github.com/eteu-technologies/near-api-go/pkg/client"
	client "main/client"

	"github.com/eteu-technologies/near-api-go/pkg/types"
	"github.com/eteu-technologies/near-api-go/pkg/types/action"
	"github.com/eteu-technologies/near-api-go/pkg/types/hash"
	"github.com/eteu-technologies/near-api-go/pkg/types/key"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
)

type ExperimentResult struct {
	TxnHash     string
	SentAt      int64
	FinalisedAt uint64
}

type FunctionCallError struct {
	Err         error
	GoroutineID int
}

// Define a struct to hold the transaction rates configuration.
type TxsConfig struct {
	Txs map[int]int `yaml:"txs"`
}

var resultsMux sync.Mutex

func main() {
	app := &cli.App{
		Name:  "interact_contract",
		Usage: "Interacts with the smart contract",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "account",
				Required: true,
				Usage:    "Account id to interact with",
			},
			&cli.StringFlag{
				Name:     "method",
				Required: true,
				Usage:    "Contract method to call",
			},
			&cli.StringFlag{
				Name:     "contract_id",
				Required: true,
				Usage:    "Contract id to interact with",
			},
			&cli.StringFlag{
				Name:     "network",
				Required: true,
				Usage:    "Network to send transactions to",
			},
			&cli.IntFlag{
				Name:     "workers",
				Required: true,
				Usage:    "Number of workers in the pool",
			},
			&cli.StringFlag{
				Name:     "args",
				Required: true,
				Usage:    "Arguments for the smart contract",
			},
			&cli.StringFlag{
				Name:     "workload",
				Required: true,
				Usage:    "Workload configuration file",
			},
		},
		Action: entrypoint,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func entrypoint(cctx *cli.Context) error {
	network, ok := client.Networks[cctx.String("network")]
	if !ok {
		log.Fatalf("unknown network '%s'", cctx.String("network"))
		return nil
	}

	rpc, err := client.NewClient(network.NodeURL)
	if err != nil {
		log.Fatalf("failed to create rpc client: %v", err)
		return nil
	}

	accountID := cctx.String("account")
	if accountID == "" {
		log.Fatalf("failed to parse account id: %v", err)
		return nil
	}
	contractID := cctx.String("contract_id")
	if contractID == "" {
		log.Fatalf("failed to parse contract id: %v", err)
		return nil
	}

	methodName := cctx.String("method")

	keyPair, err := resolveCredentials(network.NetworkID, accountID)
	if err != nil {
		log.Fatalf("failed to load private key: %v", err)
		return nil
	}

	workers := cctx.Int("workers")
	if workers == 0 {
		log.Fatalf("failed to parse workers: %v", err)
		return nil
	}

	workloadFile := cctx.String("workload")
	if workloadFile == "" {
		log.Fatalf("failed to parse workload file: %v", err)
		return nil
	}
	// Workloads are defined in the workloads directory
	workloadFile = "workloads/" + workloadFile + ".yml"

	config, err := loadConfig(workloadFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	var args []byte
	argsStr := ""
	fileName := cctx.String("args")
	if len(fileName) > 1 {
		content, err := ioutil.ReadFile(fileName)
		if err != nil {
			fmt.Println("Error reading file:", err)
			os.Exit(1)
		}
		argsStr = string(content)

	}
	args = []byte(argsStr)

	// Total number of transactions to be sent
	iterations := calculateTotalTransactions(config)

	// Define the channel to collect results from all goroutines
	resultsChan := make(chan *client.ExperimentResult, iterations)
	errorChan := make(chan *FunctionCallError, iterations)
	var experimentResults []*client.ExperimentResult
	var updateWg sync.WaitGroup

	// Initialize CSV writer
	csvWriter, file, err := client.InitializeCSVWriter(contractID)
	if err != nil {
		log.Fatalf("Error initializing CSV writer: %v", err)
	}

	// Start error logging goroutine
	go func() {
		for err := range errorChan {
			log.Printf("Error from goroutine %d: %v", err.GoroutineID, err.Err.Error())
		}
	}()

	// Create a channel to control worker start
	startSignal := make(chan struct{})

	for i := 0; i < workers; i++ {
		updateWg.Add(1)
		go func(id int) {
			<-startSignal // Wait for the signal to close
			updateWorker(id, accountID, &updateWg, resultsChan, cctx, rpc, &experimentResults, errorChan, csvWriter)
		}(i)
	}

	// Prepare a queue to hold blobs
	var txnQueue []string

	// If no method name is provided, default to transfer
	var actionType action.Action
	if methodName == "" {
		actionType = action.NewTransfer(types.BalanceFromFloat(0.00001))
	} else {
		actionType = action.NewFunctionCall(methodName, args, 300000000000000, types.NEARToYocto(0))
	}

	println("Preparing transactions.")
	// Batch Prepare Transactions
	for i := 0; i < iterations; i++ {
		_, blob, err := rpc.PrepareTransaction(
			cctx.Context,
			accountID,
			contractID,
			[]action.Action{
				actionType,
			},
			client.WithLatestBlock(),
			client.WithKeyPair(keyPair),
		)
		if err != nil {
			fmt.Println("Error preparing transaction:", err)
			continue
		}
		txnQueue = append(txnQueue, blob)
	}

	// Send Transactions in Loop
	go func() {
		println("Sending transactions to the NEAR network.")
		spacingMicros := 0
		startTime := time.Now()
		curElapsed := -1
		for i, blob := range txnQueue {
			elapsed := int(time.Since(startTime).Seconds())
			tps, exists := config.Txs[elapsed]
			if elapsed != curElapsed && exists && tps > 0 {
				spacingMicros = int(1e6 / tps)
				curElapsed = elapsed
			}
			var sentAt int64 = 0
			res, err := rpc.RPCTransactionSend(cctx.Context, blob, &sentAt, spacingMicros)
			if err != nil {
				errorChan <- &FunctionCallError{Err: err, GoroutineID: i}
				continue
			}
			resultsChan <- &client.ExperimentResult{
				TxnHash: res.String(),
				SentAt:  sentAt,
			}
		}
		println("All transactions sent. Waiting for 2 mins to start processing.")
		close(resultsChan)
	}()

	// Timer to close the start signal after 120 seconds
	go func() {
		time.Sleep(120 * time.Second)
		println("Processing transactions. Updating results csv.")
		close(startSignal)
	}()

	// Wait for update workers to complete
	updateWg.Wait()

	close(errorChan)
	file.Close()
	csvWriter.Flush()

	return nil
}

// Load the transactions configuration from the file.
func loadConfig(filename string) (TxsConfig, error) {
	var config TxsConfig
	data, err := os.ReadFile(filename)
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(data, &config)
	return config, err
}

// Calculate the microseconds between transactions based on the transactions per second.
func calculateSpacingMicros(tps int) int64 {
	if tps == 0 {
		return 0 // Return 0 to indicate no transactions should be sent.
	}
	return int64(1e6 / tps)
}

// Calculate the total number of transactions to be sent based on the transactions configuration.
func calculateTotalTransactions(config TxsConfig) int {
	totalTransactions := 0

	// Create a sorted slice of keys from the map to process times in order
	var times []int
	for time := range config.Txs {
		times = append(times, time)
	}
	sort.Ints(times)

	// Calculate transactions for each segment
	for i := 0; i < len(times)-1; i++ {
		startTime := times[i]
		nextTime := times[i+1]
		tps := config.Txs[startTime]
		duration := nextTime - startTime
		totalTransactions += tps * duration
	}

	fmt.Println("Total Transactions:", totalTransactions)
	return totalTransactions
}

func updateWorker(id int, accountID string, wg *sync.WaitGroup, resultsChan <-chan *client.ExperimentResult, cctx *cli.Context, rpc client.Client, experimentResults *[]*client.ExperimentResult, errorChan chan *FunctionCallError, csvWriter *csv.Writer) {
	defer wg.Done()

	for result := range resultsChan {
		// log.Printf("Update worker %d updating result with txn hash %s", id, result.TxnHash)
		err := updateExperimentResult(id, accountID, result, cctx, rpc, experimentResults, errorChan, csvWriter)
		if err != nil {
			errorChan <- &FunctionCallError{Err: err, GoroutineID: id}
		}
	}
}

func updateExperimentResult(id int, accountID string, result *client.ExperimentResult, cctx *cli.Context, rpc client.Client, experimentResults *[]*client.ExperimentResult, errorChan chan *FunctionCallError, csvWriter *csv.Writer) error {

	cryptoHash, err := hash.NewCryptoHashFromBase58(result.TxnHash)
	if err != nil {
		fmt.Println("Error converting TxnHash to CryptoHash:", err)
		return err
	}

	var txnStatus client.FinalExecutionOutcomeView
	const maxRetries = 2
	const waitTime = 2 * time.Second
	// println("Current timestamp", time.Now().UnixMicro())

	for i := 0; i < maxRetries; i++ {
		txnStatus, err = rpc.TransactionStatus(cctx.Context, cryptoHash, accountID)
		if err == nil {
			break
		}
		time.Sleep(waitTime)
	}

	if err != nil {
		return err
	}

	var blockInfo client.BlockView
	for i := 0; i < maxRetries; i++ {
		time.Sleep(waitTime)
		blockInfo, err = rpc.BlockDetails(cctx.Context, func(params map[string]interface{}) {
			params["block_id"] = txnStatus.TransactionOutcome.BlockHash
		})
		if err == nil {
			break
		}
	}
	if err != nil {
		return err
	}

	result.Block_Height = blockInfo.Header.Height
	result.GasFee = blockInfo.Header.GasPrice
	result.GasBurnt = txnStatus.TransactionOutcome.Outcome.GasBurnt
	result.FinalisedAt = blockInfo.Header.Timestamp

	// Add the result to the slice of results
	resultsMux.Lock()
	// *experimentResults = append(*experimentResults, result)
	client.WriteToCSV(csvWriter, result)
	resultsMux.Unlock()

	return nil
}

func worker(id int,
	wg *sync.WaitGroup,
	tasks <-chan int,
	results chan<- *client.ExperimentResult,
	errors chan<- *FunctionCallError,
	cctx *cli.Context,
	rpc client.Client,
	accountID string,
	contractID string,
	methodName string,
	keyPair key.KeyPair,
	resultsChan chan *client.ExperimentResult,
	errorChan chan *FunctionCallError,
) {
	defer wg.Done()

	for task := range tasks {
		log.Printf("Worker %d processing task %d", id, task)

		// makeFunctionCall(cctx, rpc, accountID, contractID, methodName, keyPair, spacingMicros, resultsChan, errorChan)
	}
}

func makeFunctionCall(
	cctx *cli.Context,
	rpc client.Client,
	accountID string,
	contractID string,
	methodName string,
	keyPair key.KeyPair,
	spacingMicros int,
	id int,
	resultsChan chan *client.ExperimentResult,
	errorChan chan *FunctionCallError,

) {
	startTime := time.Now()
	var sentAt int64 = 0

	// Send the transaction
	res, err := rpc.TransactionSend(
		cctx.Context,
		accountID,
		contractID,
		[]action.Action{
			action.NewFunctionCall(methodName, nil, types.DefaultFunctionCallGas, types.NEARToYocto(0)),
		},
		&sentAt,
		spacingMicros,
		client.WithLatestBlock(),
		client.WithKeyPair(keyPair),
	)
	if err != nil {
		resultsChan <- nil
		errorChan <- &FunctionCallError{Err: err, GoroutineID: id}
		return
	}

	if err != nil {
		resultsChan <- nil
		errorChan <- &FunctionCallError{Err: err, GoroutineID: id}
		return
	}

	result := &client.ExperimentResult{
		TxnHash: res.String(),
		SentAt:  sentAt,
	}
	resultsChan <- result

	endTime := time.Now()
	elapsed := endTime.Sub(startTime).Microseconds()
	fmt.Printf("Total time taken by makeFunctionCall: %d microseconds\n", elapsed)
}

func resolveCredentials(networkName string, id types.AccountID) (kp key.KeyPair, err error) {
	var creds struct {
		AccountID  types.AccountID     `json:"account_id"`
		PublicKey  key.Base58PublicKey `json:"public_key"`
		PrivateKey key.KeyPair         `json:"private_key"`
	}

	var home string
	home, err = os.UserHomeDir()
	if err != nil {
		return
	}

	credsFile := filepath.Join(home, ".near-credentials", networkName, fmt.Sprintf("%s.json", id))
	// credsFile := filepath.Join(home, ".near", id, "node_key.json")

	var cf *os.File
	if cf, err = os.Open(credsFile); err != nil {
		return
	}
	defer cf.Close()

	if err = json.NewDecoder(cf).Decode(&creds); err != nil {
		return
	}

	if creds.PublicKey.String() != creds.PrivateKey.PublicKey.String() {
		err = fmt.Errorf("inconsistent public key, %s != %s", creds.PublicKey.String(), creds.PrivateKey.PublicKey.String())
		return
	}
	kp = creds.PrivateKey
	return
}
