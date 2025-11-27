package handlers

import (
	"errors"
	"kns-indexer/models"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/proxy"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewDomainHandler(pool *pgxpool.Pool) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		hostname := ctx.Hostname()
		username := strings.SplitN(hostname, ".", 2)[0]
		if username == hostname {
			return ctx.Next()
		}

		var cid *string

		err := pool.QueryRow(
			ctx.Context(), "SELECT cid FROM username WHERE username = $1;", strings.ToLower(username),
		).Scan(&cid)

		if errors.Is(err, pgx.ErrNoRows) || cid == nil {
			return ctx.Next()
		} else if err != nil {
			slog.Error("failed to get CID by username", "username", username, "error", err)
			return ctx.Status(fiber.StatusInternalServerError).JSON(
				models.FailureResponse{Status: "error", Error: "internal server error"},
			)
		}

		return proxy.Do(ctx, "https://dweb.link/ipfs/"+*cid+ctx.OriginalURL())
	}
}
