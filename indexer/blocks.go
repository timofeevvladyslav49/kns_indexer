package indexer

import (
	"sort"
)

func sortedBlocks(history map[string]any) []map[string]any {
	var blocks []map[string]any
	for _, h := range history["history"].([]any) {
		for _, b := range h.(map[string]any)["voteStaple"].(map[string]any)["blocks"].([]any) {
			blocks = append(blocks, b.(map[string]any))
		}
	}
	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i]["date"].(string) < blocks[j]["date"].(string)
	})
	return blocks
}
