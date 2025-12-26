// DecentralizedYoutube.ts
import {
  NearBindgen,
  near,
  call,
  view,
  LookupMap,
  UnorderedMap,
  initialize,
} from 'near-sdk-js';

@NearBindgen({})
export class DecentralizedYoutube {
  modifications: UnorderedMap<string> = new UnorderedMap<string>(
    'unique-key-2',
  );

  @initialize({ privateFunction: true })
  init() {}

  @view({})
  getModification({ accountId }: { accountId: string }): string | null {
    let val: string | null = this.modifications.get(accountId);
    return val ? val : null;
  }

  @call({})
  upload({ data }: { data: string }): void {
    const sender = near.signerAccountId();
    this.modifications.set(sender, data);
    near.log(`NewUpload: ${sender}, ${data}`);
  }
}
