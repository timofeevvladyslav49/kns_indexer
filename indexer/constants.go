package indexer

import (
	"os"
	"regexp"
)

const (
	TransactionsPageLimit = 100
	LaunchDate            = "2025-12-02"

	TokenName = "KNS"

	BurnAddress = "keeta_aeaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaazpi2nodu"
)

var (
	KeetaBaseURL    = os.Getenv("KEETA_BASE_URL")
	KeetoolsBaseURL = os.Getenv("KEETOOLS_BASE_URL")

	UsernamePattern, _ = regexp.Compile(`^[a-z0-9_]{1,32}$`)

	SetPrimaryNamePattern, _ = regexp.Compile(`^set_primary_name (keeta_\w+)$`)
	SetCidPattern, _         = regexp.Compile(`^set_cid (keeta_\w+) (\w+)$`)
)
