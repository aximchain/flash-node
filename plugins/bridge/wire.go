package bridge

import (
	"github.com/aximchain/flash-node/wire"
)

// Register concrete types on wire codec
func RegisterWire(cdc *wire.Codec) {
	cdc.RegisterConcrete(BindMsg{}, "bridge/BindMsg", nil)
	cdc.RegisterConcrete(UnbindMsg{}, "bridge/UnbindMsg", nil)
	cdc.RegisterConcrete(TransferOutMsg{}, "bridge/TransferOutMsg", nil)
}
