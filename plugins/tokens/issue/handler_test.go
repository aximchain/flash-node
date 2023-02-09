package issue

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aximchain/axc-cosmos-sdk/baseapp"
	sdk "github.com/aximchain/axc-cosmos-sdk/types"
	"github.com/aximchain/axc-cosmos-sdk/x/auth"
	"github.com/aximchain/axc-cosmos-sdk/x/bank"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/aximchain/flash-node/common/testutils"
	"github.com/aximchain/flash-node/common/types"
	"github.com/aximchain/flash-node/common/upgrade"
	"github.com/aximchain/flash-node/plugins/tokens/store"
	"github.com/aximchain/flash-node/wire"
)

func setup() (sdk.Context, sdk.Handler, auth.AccountKeeper, store.Mapper) {
	ms, capKey1, capKey2, _ := testutils.SetupThreeMultiStoreForUnitTest()
	cdc := wire.NewCodec()
	cdc.RegisterInterface((*types.IToken)(nil), nil)
	cdc.RegisterConcrete(&types.Token{}, "axcchain/Token", nil)
	cdc.RegisterConcrete(&types.MiniToken{}, "axcchain/MiniToken", nil)
	tokenMapper := store.NewMapper(cdc, capKey1)
	accountKeeper := auth.NewAccountKeeper(cdc, capKey2, auth.ProtoBaseAccount)
	bankKeeper := bank.NewBaseKeeper(accountKeeper)
	handler := NewHandler(tokenMapper, bankKeeper)
	accountStore := ms.GetKVStore(capKey2)
	accountStoreCache := auth.NewAccountStoreCache(cdc, accountStore, 10)
	ctx := sdk.NewContext(ms, abci.Header{ChainID: "mychainid", Height: 1},
		sdk.RunTxModeDeliver, log.NewNopLogger()).
		WithAccountCache(auth.NewAccountCache(accountStoreCache))
	return ctx, handler, accountKeeper, tokenMapper
}

func setChainVersion() {
	upgrade.Mgr.AddUpgradeHeight(upgrade.BEP8, -1)
}

func resetChainVersion() {
	upgrade.Mgr.Config.HeightMap = nil
}

func TestHandleIssueToken(t *testing.T) {
	ctx, handler, accountKeeper, tokenMapper := setup()
	_, acc := testutils.NewAccount(ctx, accountKeeper, 100e8)

	ctx = ctx.WithValue(baseapp.TxHashKey, "000")
	msg := NewIssueMsg(acc.GetAddress(), "New AXC", "NNB", 100000e8, false)
	sdkResult := handler(ctx, msg)
	require.Equal(t, true, sdkResult.Code.IsOK())
	token, err := tokenMapper.GetToken(ctx, "NNB-000")
	require.NoError(t, err)
	expectedToken, err := types.NewToken("New AXC", "NNB-000", 100000e8, acc.GetAddress(), false)
	require.Equal(t, expectedToken, token)

	sdkResult = handler(ctx, msg)
	require.Contains(t, sdkResult.Log, "symbol(NNB) already exists")
}

func TestHandleMintToken(t *testing.T) {
	ctx, handler, accountKeeper, tokenMapper := setup()
	_, acc := testutils.NewAccount(ctx, accountKeeper, 100e8)
	mintMsg := NewMintMsg(acc.GetAddress(), "NNB-000", 10000e8)
	sdkResult := handler(ctx, mintMsg)
	require.Contains(t, sdkResult.Log, "symbol(NNB-000) does not exist")

	issueMsg := NewIssueMsg(acc.GetAddress(), "New AXC", "NNB", 100000e8, true)
	ctx = ctx.WithValue(baseapp.TxHashKey, "000")
	sdkResult = handler(ctx, issueMsg)
	require.Equal(t, true, sdkResult.Code.IsOK())

	sdkResult = handler(ctx, mintMsg)
	require.Equal(t, true, sdkResult.Code.IsOK())

	token, err := tokenMapper.GetToken(ctx, "NNB-000")
	require.NoError(t, err)
	expectedToken, err := types.NewToken("New AXC", "NNB-000", 110000e8, acc.GetAddress(), true)
	require.Equal(t, expectedToken, token)

	invalidMintMsg := NewMintMsg(acc.GetAddress(), "NNB-000", types.TokenMaxTotalSupply)
	sdkResult = handler(ctx, invalidMintMsg)
	require.Contains(t, sdkResult.Log, "mint amount is too large")

	_, acc2 := testutils.NewAccount(ctx, accountKeeper, 100e8)
	invalidMintMsg = NewMintMsg(acc2.GetAddress(), "NNB-000", types.TokenMaxTotalSupply)
	sdkResult = handler(ctx, invalidMintMsg)
	require.Contains(t, sdkResult.Log, "only the owner can mint token NNB")

	// issue a non-mintable token
	issueMsg = NewIssueMsg(acc.GetAddress(), "New AXC2", "NNB2", 100000e8, false)
	ctx = ctx.WithValue(baseapp.TxHashKey, "000")
	sdkResult = handler(ctx, issueMsg)
	require.Equal(t, true, sdkResult.Code.IsOK())

	mintMsg = NewMintMsg(acc.GetAddress(), "NNB2-000", 10000e8)
	sdkResult = handler(ctx, mintMsg)
	require.Contains(t, sdkResult.Log, "token(NNB2-000) cannot be minted")

	// mint native token
	invalidMintMsg = NewMintMsg(acc.GetAddress(), "AXC", 10000e8)
	require.Contains(t, invalidMintMsg.ValidateBasic().Error(), "cannot mint native token")
}
