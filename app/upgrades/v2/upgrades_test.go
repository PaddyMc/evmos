package v2_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/tharsis/evmos/v2/app"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	feemarkettypes "github.com/tharsis/ethermint/x/feemarket/types"
)

type UpgradeTestSuite struct {
	suite.Suite

	ctx sdk.Context
	app *app.Evmos
}

func (suite *UpgradeTestSuite) SetupTest() {
	// setup feemarketGenesis params
	feemarketGenesis := feemarkettypes.DefaultGenesisState()
	feemarketGenesis.Params.EnableHeight = 1
	feemarketGenesis.Params.NoBaseFee = false

	suite.app = app.Setup(false, feemarketGenesis)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{
		ChainID: "evmos_9001-2",
	})
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

const upgradeHeight = 58700

func (suite *UpgradeTestSuite) TestUpdateClaims() {
	testCases := []struct {
		msg        string
		preUpdate  func()
		update     func()
		postUpdate func()
		expPass    bool
	}{
		{
			"Test Claims module migrations",
			func() {
				claimsParams := suite.app.ClaimsKeeper.GetParams(suite.ctx)
				claimsParams.EnableClaims = false
				suite.app.ClaimsKeeper.SetParams(suite.ctx, claimsParams)
				suite.Require().False(suite.app.ClaimsKeeper.GetParams(suite.ctx).EnableClaims)
			},
			func() {
				suite.ctx = suite.ctx.WithBlockHeight(upgradeHeight)
				suite.Require().NotPanics(func() {
					beginBlockRequest := abci.RequestBeginBlock{}
					suite.app.BeginBlocker(suite.ctx, beginBlockRequest)
				})
			},
			func() {
				claimsParams := suite.app.ClaimsKeeper.GetParams(suite.ctx)
				suite.Require().True(claimsParams.EnableClaims)
			},
			true,
		},
		{
			"Test ERC20 module migrations",
			func() {
				erc20 := suite.app.Erc20Keeper.GetParams(suite.ctx)
				erc20.EnableEVMHook = false
				erc20.EnableErc20 = false
				suite.app.Erc20Keeper.SetParams(suite.ctx, erc20)
				suite.Require().False(suite.app.Erc20Keeper.GetParams(suite.ctx).EnableEVMHook)
				suite.Require().False(suite.app.Erc20Keeper.GetParams(suite.ctx).EnableErc20)
			},
			func() {
				suite.ctx = suite.ctx.WithBlockHeight(upgradeHeight)
				suite.Require().NotPanics(func() {
					beginBlockRequest := abci.RequestBeginBlock{}
					suite.app.BeginBlocker(suite.ctx, beginBlockRequest)
				})
			},
			func() {
				erc20 := suite.app.Erc20Keeper.GetParams(suite.ctx)
				suite.Require().True(erc20.EnableEVMHook)
				suite.Require().True(erc20.EnableErc20)
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest()

			tc.preUpdate()
			tc.update()
			tc.postUpdate()
		})
	}
}
