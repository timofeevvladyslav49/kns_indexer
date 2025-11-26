package indexer

import (
	"os"
	"regexp"
)

const (
	TransactionsPageLimit = 100
	LaunchDate            = "2025-11-25"

	TokenName = "KNS"

	// since Keeta Network doesn't have official burn address, we'll use testnet faucet address as burn address
	BurnAddress = "keeta_aabszsbrqppriqddrkptq5awubshpq3cgsoi4rc624xm6phdt74vo5w7wipwtmiw"
)

var (
	KeetaBaseURL    = os.Getenv("KEETA_BASE_URL")
	KeetoolsBaseURL = os.Getenv("KEETOOLS_BASE_URL")

	UsernamePattern, _ = regexp.Compile(`^[a-z0-9_]{1,32}$`)

	SetPrimaryNamePattern, _ = regexp.Compile(`^set_primary_name (keeta_\w+)$`)
	SetCidPattern, _         = regexp.Compile(`^set_cid (keeta_\w+) (\w+)$`)
)
