package keeper_test

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	"go.uber.org/mock/gomock"
)

func (suite *KeeperTestSuite) getAccountsAndPrepareMocksForSendMsg() (fromAddr, toAddr sdk.AccAddress) {
	fromAddr = sdk.AccAddress([]byte("sender"))
	toAddr = sdk.AccAddress([]byte("recipient"))

	// Mocks for fund account
	suite.authKeeper.EXPECT().GetModuleAccount(gomock.Any(), "stake_burner").Return(burnerAcc).AnyTimes()
	suite.authKeeper.EXPECT().GetModuleAccount(gomock.Any(), "stake2_minter").Return(minterAcc).AnyTimes()
	suite.authKeeper.EXPECT().GetModuleAccount(gomock.Any(), "mint").Return(mintAcc).AnyTimes()

	suite.authKeeper.EXPECT().GetModuleAddress("stake_burner").Return(burnerAcc.GetAddress()).AnyTimes()
	suite.authKeeper.EXPECT().GetModuleAddress("stake2_minter").Return(minterAcc.GetAddress()).AnyTimes()
	suite.authKeeper.EXPECT().GetModuleAddress("mint").Return(mintAcc.GetAddress()).AnyTimes()

	// Mocks for send msg
	suite.authKeeper.EXPECT().GetAccount(suite.ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, addr sdk.AccAddress) sdk.AccountI {
			if addr.String() == fromAddr.String() {
				return authtypes.NewBaseAccount(fromAddr, nil, 0, 0)
			}
			if addr.String() == burnerAcc.GetAddress().String() {
				return burnerAcc
			}
			if addr.String() == minterAcc.GetAddress().String() {
				return minterAcc
			}
			if addr.String() == mintAcc.GetAddress().String() {
				return mintAcc
			}
			return nil
		}).AnyTimes()

	suite.authKeeper.EXPECT().HasAccount(suite.ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, addr sdk.AccAddress) bool {
			if addr.String() == toAddr.String() || addr.String() == fromAddr.String() {
				return true
			}
			return false
		}).AnyTimes()

	suite.authKeeper.EXPECT().NewAccountWithAddress(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, addr sdk.AccAddress) sdk.AccountI {
			return authtypes.NewBaseAccount(addr, nil, 0, 0)
		}).AnyTimes()

	suite.authKeeper.EXPECT().SetAccount(gomock.Any(), gomock.Any()).AnyTimes()

	return
}

func (suite *KeeperTestSuite) TestCustomSendCoinsNonNativeDenom() {
	fromAddr, toAddr := suite.getAccountsAndPrepareMocksForSendMsg()
	// Fund sender account with 'atom'
	initCoins := sdk.NewCoins(sdk.NewCoin("atom", math.NewInt(1000)))
	err := banktestutil.FundAccount(suite.ctx, suite.bankKeeper, fromAddr, initCoins)
	suite.Require().NoError(err)

	// Check balance of sender before send coins
	balFrom := suite.bankKeeper.GetBalance(suite.ctx, fromAddr, "atom")
	suite.Require().Equal("1000", balFrom.Amount.String(), "Sender stake balance incorrect before send")

	// Send '500atom' to recipient
	sendAmt := math.NewInt(500)
	coinsToSend := sdk.NewCoins(sdk.NewCoin("atom", sendAmt))
	err = suite.bankKeeper.SendCoins(suite.ctx, fromAddr, toAddr, coinsToSend)
	suite.Require().NoError(err)

	// Check balance of sender for atom after send
	balFrom = suite.bankKeeper.GetBalance(suite.ctx, fromAddr, "atom")
	suite.Require().Equal("500", balFrom.Amount.String())

	// Check balance of recipient for atom
	balToAtom := suite.bankKeeper.GetBalance(suite.ctx, toAddr, "atom")
	suite.Require().Equal("500", balToAtom.Amount.String())

	// No custom stake2 loaded to recipient
	balToStake2 := suite.bankKeeper.GetBalance(suite.ctx, toAddr, "stake2")
	suite.Require().Equal("0", balToStake2.Amount.String())
}

func (suite *KeeperTestSuite) TestCustomSendCoinsNativeDenom() {
	fromAddr, toAddr := suite.getAccountsAndPrepareMocksForSendMsg()

	// Fund sender account with 'stake'
	initAmt := math.NewInt(1000)
	initCoins := sdk.NewCoins(sdk.NewCoin("stake", initAmt))

	err := banktestutil.FundAccount(suite.ctx, suite.bankKeeper, fromAddr, initCoins)
	suite.Require().NoError(err)

	// Check balance of sender before send coins
	balFrom := suite.bankKeeper.GetBalance(suite.ctx, fromAddr, "stake")
	suite.Require().Equal("1000", balFrom.Amount.String(), "Sender stake balance incorrect before send")

	// Send '500stake' to recipient
	sendAmt := math.NewInt(500)
	coinsToSend := sdk.NewCoins(sdk.NewCoin("stake", sendAmt))
	err = suite.bankKeeper.SendCoins(suite.ctx, fromAddr, toAddr, coinsToSend)
	suite.Require().NoError(err)

	// Check balance of sender for stake after send
	balFrom = suite.bankKeeper.GetBalance(suite.ctx, fromAddr, "stake")
	suite.Require().Equal("500", balFrom.Amount.String(), "Sender stake balance incorrect")

	// Check balance of recipient for stake
	balToStake := suite.bankKeeper.GetBalance(suite.ctx, toAddr, "stake")
	suite.Require().Equal("250", balToStake.Amount.String(), "Recipient stake balance incorrect")

	// Check balance of recipient for stake2
	balToStake2 := suite.bankKeeper.GetBalance(suite.ctx, toAddr, "stake2")
	suite.Require().Equal("250", balToStake2.Amount.String(), "Recipient stake2 balance incorrect")

	sdkCtx := sdk.UnwrapSDKContext(suite.ctx)
	events := sdkCtx.EventManager().Events()
	var foundBurn, foundMint bool

	for _, evt := range events {
		if evt.Type == "burn" {
			for _, attr := range evt.Attributes {
				if attr.Key == "amount" && attr.Value == "250stake" {
					foundBurn = true
				}
			}
		}
		if evt.Type == "coinbase" {
			for _, attr := range evt.Attributes {
				if attr.Key == "amount" && attr.Value == "250stake2" {
					foundMint = true
				}
			}
		}
	}

	suite.Require().True(foundBurn, "Missing expected burn event for: 250stake")
	suite.Require().True(foundMint, "Missing expected mint event for: 250stake2")
}

func (suite *KeeperTestSuite) TestCustomSendCoinsUnevenAmount() {
	fromAddr, toAddr := suite.getAccountsAndPrepareMocksForSendMsg()

	// Fund sender account with 'stake'
	initCoins := sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(1000)))
	err := banktestutil.FundAccount(suite.ctx, suite.bankKeeper, fromAddr, initCoins)
	suite.Require().NoError(err)

	// Check balance of sender before send coins
	balFrom := suite.bankKeeper.GetBalance(suite.ctx, fromAddr, "stake")
	suite.Require().Equal("1000", balFrom.Amount.String(), "Sender stake balance incorrect before send")

	// Send '501stake' to recipient
	sendAmt := math.NewInt(501)
	coinsToSend := sdk.NewCoins(sdk.NewCoin("stake", sendAmt))
	err = suite.bankKeeper.SendCoins(suite.ctx, fromAddr, toAddr, coinsToSend)
	suite.Require().NoError(err)

	// Check balance of sender for stake after send
	balFrom = suite.bankKeeper.GetBalance(suite.ctx, fromAddr, "stake")
	suite.Require().Equal("499", balFrom.Amount.String())

	// Check balance of recipient for stake
	balToStake := suite.bankKeeper.GetBalance(suite.ctx, toAddr, "stake")
	suite.Require().Equal("250", balToStake.Amount.String())

	// Check balance of recipient for stake2
	balToStake2 := suite.bankKeeper.GetBalance(suite.ctx, toAddr, "stake2")
	suite.Require().Equal("250", balToStake2.Amount.String())
}

func (suite *KeeperTestSuite) TestCustomMultiSendCoins() {
	fromAddr, toAddr := suite.getAccountsAndPrepareMocksForSendMsg()
	toAddr2 := sdk.AccAddress([]byte("recipient2"))
	// setup mock true for secondary recipient
	suite.authKeeper.EXPECT().HasAccount(suite.ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, addr sdk.AccAddress) bool {
			if addr.String() == toAddr.String() || addr.String() == fromAddr.String() || addr.String() == toAddr2.String() {
				return true
			}
			return false
		}).AnyTimes()

	// Fund sender account with 'stake'
	initCoins := sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(1000)))
	err := banktestutil.FundAccount(suite.ctx, suite.bankKeeper, fromAddr, initCoins)
	suite.Require().NoError(err)

	// Send '400stake' to two recipients
	input := types.Input{
		Address: fromAddr.String(),
		Coins:   sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(400))),
	}
	outputs := []types.Output{
		{Address: toAddr.String(), Coins: sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(200)))},
		{Address: toAddr2.String(), Coins: sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(200)))},
	}
	err = suite.bankKeeper.InputOutputCoins(suite.ctx, input, outputs)
	suite.Require().NoError(err)

	// Check balance of sender for stake after send
	balFrom := suite.bankKeeper.GetBalance(suite.ctx, fromAddr, "stake")
	suite.Require().Equal("600", balFrom.Amount.String())

	// Check balance of recipient 1 for stake
	balTo1 := suite.bankKeeper.GetBalance(suite.ctx, toAddr, "stake")
	suite.Require().Equal("100", balTo1.Amount.String())

	// Check balance of recipient 1 for stake2
	balTo1Stake2 := suite.bankKeeper.GetBalance(suite.ctx, toAddr, "stake2")
	suite.Require().Equal("100", balTo1Stake2.Amount.String())

	// Check balance of recipient 2 for stake
	balTo2 := suite.bankKeeper.GetBalance(suite.ctx, toAddr2, "stake")
	suite.Require().Equal("100", balTo2.Amount.String())

	// Check balance of recipient 2 for stake2
	balTo2Stake2 := suite.bankKeeper.GetBalance(suite.ctx, toAddr, "stake2")
	suite.Require().Equal("100", balTo2Stake2.Amount.String())
}
