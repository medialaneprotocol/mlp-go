# mlp-go

Go reference implementation of the Media Lane Protocol (MLP) wire format.

Mirrors `medialaneprotocol/mlp-core` (Rust) byte-for-byte. Currently implements
the PDU codec: fragment header (22 bytes), TLV header extension, and update
metadata. Full sender/receiver and scheduler are planned.

## Usage

```go
import "github.com/medialaneprotocol/mlp-go/pkg/mlp"

meta := &mlp.UpdateMeta{UpdateID: 7, UpdateBase: 6, ...}
ext := meta.ToExt()
frag := &mlp.Fragment{Header: mlp.FragmentHeader{...}, Ext: ext, Payload: []byte{1,2,3,4}}
b, _ := frag.Encode()
```

## Test

```sh
go test ./...
```

Golden-vector test ensures byte-for-byte parity with the Rust reference.
