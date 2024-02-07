package native

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/eth/tracers"
	"github.com/ethereum/go-ethereum/mamoru"
)

func init() {
	tracers.DefaultDirectory.Register("mamoru", newMamoruCallStackTracer, false)
}

var _ vm.EVMLogger = (*mamoru.CallStackTracer)(nil)
var _ tracers.Tracer = (*mamoru.CallStackTracer)(nil)

// Used to register the tracer with the tracer manager
func newMamoruCallStackTracer(ctx *tracers.Context, cfg json.RawMessage) (tracers.Tracer, error) {
	var config mamoru.CallTracerConfig
	if cfg != nil {
		if err := json.Unmarshal(cfg, &config); err != nil {
			return nil, err
		}
	}
	// First callframe contains tx context info
	// and is populated on start and end.
	return &mamoru.CallStackTracer{Txs: nil, Source: "console", Config: config}, nil
}
