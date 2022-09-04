package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Implements StakingHooks interface
type stakingHooks struct{}

func (stakingHooks) AfterValidatorCreated(sdk.Context, sdk.ValAddress) error { return nil }

func (stakingHooks) BeforeValidatorModified(sdk.Context, sdk.ValAddress) error { return nil }

func (stakingHooks) AfterValidatorRemoved(sdk.Context, sdk.ConsAddress, sdk.ValAddress) error {
	return nil
}

func (stakingHooks) AfterValidatorBonded(sdk.Context, sdk.ConsAddress, sdk.ValAddress) error {
	return nil
}

func (stakingHooks) AfterValidatorBeginUnbonding(sdk.Context, sdk.ConsAddress, sdk.ValAddress) error {
	return nil
}

func (stakingHooks) BeforeDelegationCreated(sdk.Context, sdk.AccAddress, sdk.ValAddress) error {
	return nil
}

func (stakingHooks) BeforeDelegationSharesModified(sdk.Context, sdk.AccAddress, sdk.ValAddress) error {
	return nil
}

func (stakingHooks) BeforeDelegationRemoved(sdk.Context, sdk.AccAddress, sdk.ValAddress) error {
	return nil
}

func (stakingHooks) AfterDelegationModified(sdk.Context, sdk.AccAddress, sdk.ValAddress) error {
	return nil
}

func (stakingHooks) BeforeValidatorSlashed(sdk.Context, sdk.ValAddress, sdk.Dec) error { return nil }
