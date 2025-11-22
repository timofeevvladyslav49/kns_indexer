package indexer

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Run(pool *pgxpool.Pool) {
	var (
		page                int
		lastBlockTimestamp  *time.Time
		lastBlockHash       *string
		lastBlockOperations []any
	)

	pool.QueryRow(
		context.Background(), "SELECT page, last_block_timestamp, last_block_hash FROM settings;",
	).Scan(&page, &lastBlockTimestamp, &lastBlockHash)

	for {
		pageMetadata := FetchPageMetadata(page)

		slog.Debug("Fetched", "pageMetadata", pageMetadata)

		history := FetchLedgerHistory(pageMetadata)

		var postCommitLogs []string

		transaction, _ := pool.Begin(context.Background())

		for _, block := range sortedBlocks(history) {
			blockTimestamp, _ := time.Parse(time.RFC3339Nano, block["date"].(string))

			// skip already processed blocks
			if lastBlockTimestamp != nil && lastBlockHash != nil && (blockTimestamp.Before(*lastBlockTimestamp) || block["$hash"] == lastBlockHash) {
				slog.Debug(fmt.Sprintf("Skipping block %v: older or equal to last processed", block["$hash"]))
				continue
			}

			slog.Debug(fmt.Sprintf("Processing block %v at %v", block["$hash"], blockTimestamp))

			operations := block["operations"].([]any)

			for _, operationRaw := range operations {
				operation := operationRaw.(map[string]any)

				if IsInscribeInstruction(operation, block["account"].(string), lastBlockOperations) {
					username := strings.ToLower(operation["description"].(string))
					commandTag, _ := transaction.Exec(
						context.Background(),
						"INSERT INTO username(username, address, owner, timestamp) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING;",
						username,
						block["account"],
						block["signer"],
						blockTimestamp,
					)
					if commandTag.RowsAffected() == 1 {
						postCommitLogs = append(
							postCommitLogs,
							fmt.Sprintf("%v inscribed username %v", block["signer"], username),
						)
					}
				} else if IsSetPrimaryNameOrCidInstruction(operation) {
					if match := SetPrimaryNamePattern.FindStringSubmatch(operation["extra"].(string)); match != nil {
						tokenAddress := match[0]

						var username string
						transaction.QueryRow(
							context.Background(),
							"UPDATE username SET is_primary = TRUE WHERE address = $1 AND owner = $2 RETURNING username;",
							tokenAddress,
							block["account"],
						).Scan(&username)

						if username != "" {
							transaction.Exec(
								context.Background(),
								"UPDATE username SET is_primary = FALSE WHERE address != $1 AND owner = $2;",
								tokenAddress,
								block["account"],
							)
							postCommitLogs = append(
								postCommitLogs,
								fmt.Sprintf("%v set primary name %v", block["account"], username),
							)
						}
					} else if match := SetCidPattern.FindStringSubmatch(operation["extra"].(string)); match != nil {
						tokenAddress, cid := match[0], match[1]

						var username string
						transaction.QueryRow(
							context.Background(),
							"UPDATE username SET cid = $1 WHERE address = $2 AND owner = $3 RETURNING username;",
							cid,
							tokenAddress,
							block["account"],
						).Scan(&username)

						if username != "" {
							postCommitLogs = append(
								postCommitLogs,
								fmt.Sprintf("%v set CID %v to %v", block["account"], cid, username),
							)
						}
					} else if IsTransferInstruction(operation) {
						var username string
						transaction.QueryRow(
							context.Background(),
							"UPDATE username SET owner = $1 WHERE address = $2 AND owner = $3 RETURNING username;",
							operation["to"],
							operation["token"],
							block["account"],
						).Scan(&username)
						if username != "" {
							postCommitLogs = append(
								postCommitLogs,
								fmt.Sprintf("%v transferred username %v to %v", block["account"], username, operation["to"]),
							)
						}
					}
				}

				lastBlockHashString := block["$hash"].(string)

				lastBlockTimestamp = &blockTimestamp
				lastBlockHash = &lastBlockHashString
				lastBlockOperations = operations
			}
		}

		if page != pageMetadata["totalPages"] {
			page += 1
		}

		transaction.Exec(
			context.Background(),
			"UPDATE settings SET page = $1, last_block_timestamp = $2, last_block_hash = $3;",
			page,
			lastBlockTimestamp,
			lastBlockHash,
		)
		transaction.Commit(context.Background())

		for postCommitLog := range postCommitLogs {
			slog.Debug("", postCommitLog)
		}

		slog.Debug("Committed settings", "page", page, "last_block_hash", lastBlockHash)

		time.Sleep(time.Second)
	}
}
