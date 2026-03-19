package types

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MintBurnKeeper is an interface for mint and burn for custom send coins implementation
type MintBurnKeeper interface {
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error

	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}
