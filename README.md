# mlp-go

Go implementation of the Media Lane Protocol (MLP) wire format.

The wire is frozen by the [specification](https://github.com/medialaneprotocol/mlp):
this package's PDU codec — fragment header, TLV header extension, update
metadata, carriers — matches `medialaneprotocol/mlp-core` (Rust) byte for
byte, and the golden-vector test re-encodes the reference's own vectors
without a byte of drift.

## What it implements

- **The codec**: classes `0..=4`, the flag byte (FIN/ANCHOR/RETRANSMIT/DUP/
  ECHO_REQUESTED/PROBE/URGENT, bit 7 reserved), the full TLV registry with
  the reserved and experimental skip bands, the five carriers with their
  payload and message caps.
- **Object assembly**: FIN-gated completion, the four-object and one-megabyte
  budgets, the windowed object-id dedup.
- **The sender's byte-level half**: enqueue an update, fragment it for a
  carrier, stamp every part with its metadata, FIN the last part, poll the
  encoded fragments.
- **The receiver's accept path**: parse, dedup, assemble, and the 64-update
  ack window (`MediaAckWindow`) the way `mlp-ts` carries it.
- **The lane vocabulary**: the ten wire states, the close codes, the
  unsupported reasons; the park table with the TTL floor/ceiling; the
  pressed-set input sequencing; flow control, loss detection, the token
  bucket, QoS tiers, 0-RTT tickets, path migration and probing helpers.

## What it does not implement (yet)

The pacing machines, the bandwidth estimator, the repair plans and the
scheduler's priority bands are the Rust machine's; a Go peer that needs
them consumes `mlp-core` or waits for the port. The package exposes no
placeholder for any of them.

## Usage

```go
import "github.com/medialaneprotocol/mlp-go/pkg/mlp"

sender := mlp.NewMediaSenderMachine(mlp.CellGrid{WidthPx: 320, HeightPx: 240, CellSize: 64})
outcome, err := sender.EnqueueUpdate(
    mlp.UpdateRequest{Payload: frame, Cells: cells, Anchor: true},
    mlp.CarrierDataChannel, sendEpochMs)

receiver := mlp.NewMediaReceiverMachine()
for _, raw := range datagrams {
    obj, err := receiver.AcceptFragment(raw)
    if err == nil && obj != nil {
        receiver.NoteApplied(mlp.UpdateMetaFromExt(obj.Ext).UpdateID)
    }
}
```

## Test

```sh
go test ./...
```

The conformance test parses and re-encodes every vector in
`pkg/mlp/conformance/vectors.json` — the same corpus `mlp-core`,
`mlp-py` and `mlp-c` test against — and fails on the first byte of drift.
