package indexer

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
)

var client = &http.Client{}

func FetchPageMetadata(page int) map[string]any {
	values := url.Values{
		"limit":     {strconv.Itoa(TransactionsPageLimit)},
		"page":      {strconv.Itoa(page)},
		"sortOrder": {"asc"},
		"dateFrom":  {LaunchDate},
	}

	resp, _ := client.Get(KeetoolsBaseURL + "/api/staples/metadata?" + values.Encode())
	defer resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
	return result
}

func FetchLedgerHistory(pageMetadata map[string]any) map[string]any {
	values := url.Values{
		"limit": {strconv.Itoa(TransactionsPageLimit)},
	}
	if pageMetadata["startBlocksHash"] != nil {
		values.Set("start", pageMetadata["startBlocksHash"].(string))
	}

	resp, _ := client.Get(KeetaBaseURL + "/api/node/ledger/history?" + values.Encode())
	defer resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
	return result
}
