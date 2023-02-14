package issue

import (
	"encoding/json"
	"fmt"

	sdk "github.com/aximchain/axc-cosmos-sdk/types"

	"github.com/aximchain/flash-node/common/types"
)

// TODO: "route expressions can only contain alphanumeric characters", we need to change the cosmos sdk to support slash
// const Route  = "tokens/issue"
const (
	IssueMiniMsgType = "miniIssueMsg" //For max total supply in range 2
)

var _ sdk.Msg = IssueMiniMsg{}

type IssueMiniMsg struct {
	From        sdk.AccAddress `json:"from"`
	Name        string         `json:"name"`
	Symbol      string         `json:"symbol"`
	TotalSupply int64          `json:"total_supply"`
	Mintable    bool           `json:"mintable"`
	TokenURI    string         `json:"token_uri"`
}

func NewIssueMiniMsg(from sdk.AccAddress, name, symbol string, supply int64, mintable bool, tokenURI string) IssueMiniMsg {
	return IssueMiniMsg{
		From:        from,
		Name:        name,
		Symbol:      symbol,
		TotalSupply: supply,
		Mintable:    mintable,
		TokenURI:    tokenURI,
	}
}

// ValidateBasic does a simple validation check that
// doesn't require access to any other information.
func (msg IssueMiniMsg) ValidateBasic() sdk.Error {
	if msg.From == nil {
		return sdk.ErrInvalidAddress("sender address cannot be empty")
	}

	if err := types.ValidateIssueMiniSymbol(msg.Symbol); err != nil {
		return sdk.ErrInvalidCoins(err.Error())
	}

	if len(msg.Name) == 0 || len(msg.Name) > maxTokenNameLength {
		return sdk.ErrInvalidCoins(fmt.Sprintf("token name should have 1 ~ %v characters", maxTokenNameLength))
	}

	if len(msg.TokenURI) > types.MaxTokenURILength {
		return sdk.ErrInvalidCoins(fmt.Sprintf("token seturi should not exceed %v characters", types.MaxTokenURILength))
	}

	if msg.TotalSupply < types.MiniTokenMinExecutionAmount || msg.TotalSupply > types.MiniRangeType.UpperBound() {
		return sdk.ErrInvalidCoins(fmt.Sprintf("total supply should be between %d and %d", types.MiniTokenMinExecutionAmount, types.MiniRangeType.UpperBound()))
	}

	return nil
}

// Implements IssueMiniMsg.
func (msg IssueMiniMsg) Route() string                { return Route }
func (msg IssueMiniMsg) Type() string                 { return IssueMiniMsgType }
func (msg IssueMiniMsg) String() string               { return fmt.Sprintf("IssueMiniMsg{%#v}", msg) }
func (msg IssueMiniMsg) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{msg.From} }
func (msg IssueMiniMsg) GetSignBytes() []byte {
	b, err := json.Marshal(msg) // XXX: ensure some canonical form
	if err != nil {
		panic(err)
	}
	return b
}
func (msg IssueMiniMsg) GetInvolvedAddresses() []sdk.AccAddress {
	return msg.GetSigners()
}
