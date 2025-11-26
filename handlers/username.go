package handlers

import (
	"errors"
	"kns-indexer/models"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GetUsernameSuccessResponse = models.SuccessResponse[models.Username]

// NewGetUsernameHandler godoc
// @Summary      Resolve username
// @Description  Returns username record by username
// @Tags         username
// @Accept       json
// @Produce      json
// @Param        username  path  string  true  "Username"
// @Success      200  {object}  GetUsernameSuccessResponse
// @Failure      404  {object}  models.FailureResponse
// @Failure      500  {object}  models.FailureResponse
// @Router       /api/usernames/{username} [get]
func NewGetUsernameHandler(pool *pgxpool.Pool) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		username := ctx.Params("username")

		var u models.Username

		err := pool.QueryRow(
			ctx.Context(),
			"SELECT address, owner, cid, is_primary, timestamp FROM username WHERE username = $1;",
			strings.ToLower(username),
		).Scan(&u.Address, &u.Owner, &u.CID, &u.IsPrimary, &u.Timestamp)

		if errors.Is(err, pgx.ErrNoRows) {
			return ctx.Status(fiber.StatusNotFound).JSON(
				models.FailureResponse{Status: "error", Error: "username not found"},
			)
		} else if err != nil {
			slog.Error("failed to get username", "username", username, "error", err)
			return ctx.Status(fiber.StatusInternalServerError).JSON(
				models.FailureResponse{Status: "error", Error: "internal server error"},
			)
		}

		u.Username = username

		return ctx.JSON(GetUsernameSuccessResponse{Status: "ok", Data: u})
	}
}
