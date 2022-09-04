package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Implements GovHooks interface
type govHooks struct{}

func (govHooks) AfterProposalSubmission(sdk.Context, uint64) {}

func (govHooks) AfterProposalDeposit(sdk.Context, uint64, sdk.AccAddress) {}

func (govHooks) AfterProposalVote(sdk.Context, uint64, sdk.AccAddress) {}

func (govHooks) AfterProposalFailedMinDeposit(sdk.Context, uint64) {}

func (govHooks) AfterProposalVotingPeriodEnded(sdk.Context, uint64) {}
