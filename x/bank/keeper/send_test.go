package keeper_test

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
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

func (suite *KeeperTestSuite) TestCustomSendCoinsNativeDenom() {
	fromAddr, toAddr := suite.getAccountsAndPrepareMocksForSendMsg()

	// Fund sender account with 'stake'
	initAmt := math.NewInt(1000)
	initCoins := sdk.NewCoins(sdk.NewCoin("stake", initAmt))

	err := banktestutil.FundAccount(suite.ctx, suite.bankKeeper, fromAddr, initCoins)
	suite.Require().NoError(err)

	balFrom := suite.bankKeeper.GetBalance(suite.ctx, fromAddr, "stake")
	suite.Require().Equal("1000", balFrom.Amount.String(), "Sender stake balance incorrect before send")

	sendAmt := math.NewInt(500)
	coinsToSend := sdk.NewCoins(sdk.NewCoin("stake", sendAmt))

	// Send '500stake' to recipient
	err = suite.bankKeeper.SendCoins(suite.ctx, fromAddr, toAddr, coinsToSend)
	suite.Require().NoError(err)

	balFrom = suite.bankKeeper.GetBalance(suite.ctx, fromAddr, "stake")
	suite.Require().Equal("500", balFrom.Amount.String(), "Sender stake balance incorrect")
	balToStake := suite.bankKeeper.GetBalance(suite.ctx, toAddr, "stake")
	suite.Require().Equal("250", balToStake.Amount.String(), "Recipient stake balance incorrect")
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
