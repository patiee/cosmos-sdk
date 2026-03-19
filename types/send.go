package types

// Config values for custom send implementation with mint & burn logic
var (
	// MintDenom used during send transaction
	MintDenom = "stake2"

	// BondDenomBurnerAccount is an account with the permission to burn bond denom during send transaction
	BondDenomBurnerAccount = "stake_burner"

	// MintDenomMinterAccount is an account with the permission to mint denom during send transaction
	MintDenomMinterAccount = "stake2_minter"
)
