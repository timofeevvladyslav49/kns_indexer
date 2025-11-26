package indexer

import (
	"slices"
	"strings"
)

func IsInscribeInstruction(
	operation map[string]any,
	tokenAccount string,
	lastBlockOperations []any,
) bool {
	return int(operation["type"].(float64)) == OperationTypeSetInfo &&
		operation["name"] == TokenName &&
		UsernamePattern.MatchString(strings.ToLower(operation["description"].(string))) &&
		slices.ContainsFunc(lastBlockOperations, func(opRaw any) bool {
			op := opRaw.(map[string]any)
			return int(op["type"].(float64)) == OperationTypeCreateIdentifier && op["identifier"] == tokenAccount
		})
}

func IsTransferInstruction(operation map[string]any) bool {
	return int(operation["type"].(float64)) == OperationTypeSend && operation["amount"] == "0x1"
}

func IsSetPrimaryNameOrCidInstruction(operation map[string]any) bool {
	return int(operation["type"].(float64)) == OperationTypeSend &&
		operation["to"] == BurnAddress &&
		operation["extra"] != nil
}
