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
	return operation["type"] == OperationTypeSetInfo &&
		operation["name"] == TokenName &&
		UsernamePattern.MatchString(strings.ToLower(operation["description"].(string))) &&
		slices.ContainsFunc(lastBlockOperations, func(opRaw any) bool {
			op := opRaw.(map[string]any)
			return op["type"] == OperationTypeCreateIdentifier && op["identifier"] == tokenAccount
		})
}

func IsTransferInstruction(operation map[string]any) bool {
	return operation["type"] == OperationTypeSend && operation["amount"] == "0x1"
}

func IsSetPrimaryNameOrCidInstruction(operation map[string]any) bool {
	return operation["type"] == OperationTypeSend &&
		operation["to"] == BurnAddress &&
		operation["extra"] != nil
}
