package client

import (
	"github.com/eteu-technologies/near-api-go/pkg/types"
)

type ExperimentResult struct {
	Block_Height uint64
	TxnHash      string
	GasFee       types.Balance
	GasBurnt     uint64
	SentAt       int64
	FinalisedAt  uint64
}
