package handlers

import (
	"errors"
	"kns-indexer/models"
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GetPrimaryUsernameSuccessResponse = models.SuccessResponse[models.Username]

// NewGetPrimaryUsernameHandler godoc
// @Summary      Resolve primary username
// @Description  Returns primary username by owner
// @Tags         owner
// @Accept       json
// @Produce      json
// @Param        owner  path  string  true  "Owner"
// @Success      200  {object}  GetPrimaryUsernameSuccessResponse
// @Failure      404  {object}  models.FailureResponse
// @Failure      500  {object}  models.FailureResponse
// @Router       /api/primary-username/{owner} [get]
func NewGetPrimaryUsernameHandler(pool *pgxpool.Pool) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		owner := ctx.Params("owner")

		var u models.Username

		err := pool.QueryRow(
			ctx.Context(),
			"SELECT username, address, cid, is_primary, timestamp FROM username WHERE owner = $1 AND is_primary = true;",
			owner,
		).Scan(&u.Username, &u.Address, &u.CID, &u.IsPrimary, &u.Timestamp)

		if errors.Is(err, pgx.ErrNoRows) {
			return ctx.Status(fiber.StatusNotFound).JSON(
				models.FailureResponse{Status: "error", Error: "no primary username set"},
			)
		} else if err != nil {
			slog.Error("failed to get primary username", "owner", owner, "error", err)
			return ctx.Status(fiber.StatusInternalServerError).JSON(
				models.FailureResponse{Status: "error", Error: "internal server error"},
			)
		}

		u.Owner = owner

		return ctx.JSON(GetPrimaryUsernameSuccessResponse{Status: "ok", Data: u})
	}
}
