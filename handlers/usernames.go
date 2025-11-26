package handlers

import (
	"kns-indexer/models"
	"log/slog"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GetUsernamesSuccessResponseData struct {
	Total     uint              `json:"total" example:"100"`
	Usernames []models.Username `json:"usernames"`
}

type GetUsernamesSuccessResponse = models.SuccessResponse[GetUsernamesSuccessResponseData]

// NewGetUsernamesHandler godoc
// @Summary      Get list of usernames
// @Description  Returns paginated list of all registered usernames with sorting by timestamp
// @Tags         username
// @Accept       json
// @Produce      json
// @Param        limit      query     int     false  "Number of records per page"                                      default(100)  minimum(1)    maximum(100)
// @Param        offset     query     int     false  "Offset for pagination (starts from 0)"                           default(0)    minimum(0)
// @Param        sortOrder  query     string  false  "Sort order by timestamp: asc or desc"                            default(desc) enums(asc,desc)
// @Success      200        {object}  GetUsernamesSuccessResponse                                      "Successfully retrieved usernames"
// @Failure      422        {object}  models.FailureResponse                                           "Invalid query parameters"
// @Failure      500        {object}  models.FailureResponse                                           "Internal server error"
// @Router       /usernames [get]
func NewGetUsernamesHandler(pool *pgxpool.Pool) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		limitStr := ctx.Query("limit", "100")
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			return ctx.Status(fiber.StatusUnprocessableEntity).JSON(
				models.FailureResponse{Status: "error", Error: "limit should be a number from 1 to 100"},
			)
		}

		offsetStr := ctx.Query("offset", "0")
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			return ctx.Status(fiber.StatusUnprocessableEntity).JSON(
				models.FailureResponse{Status: "error", Error: "offset should be positive integer"},
			)
		}

		sortOrder := ctx.Query("sortOrder", "desc")
		if sortOrder != "desc" && sortOrder != "asc" {
			return ctx.Status(fiber.StatusUnprocessableEntity).JSON(
				models.FailureResponse{Status: "error", Error: "sortOrder should be desc or asc"},
			)
		}

		var total uint

		conn, err := pool.Acquire(ctx.Context())
		if err != nil {
			slog.Error("failed to acquire connection", "error", err)
			return ctx.Status(fiber.StatusInternalServerError).JSON(
				models.FailureResponse{Status: "error", Error: "internal server error"},
			)
		}
		defer conn.Release()

		if err = conn.QueryRow(ctx.Context(), "SELECT COUNT(*) FROM username;").Scan(&total); err != nil {
			slog.Error("failed to total usernames", "error", err)
			return ctx.Status(fiber.StatusInternalServerError).JSON(
				models.FailureResponse{Status: "error", Error: "internal server error"},
			)
		}

		rows, err := conn.Query(
			ctx.Context(),
			"SELECT username, address, owner, cid, is_primary, timestamp FROM username ORDER BY timestamp "+sortOrder+" LIMIT $1 OFFSET $2;",
			limit, offset,
		)
		if err != nil {
			slog.Error("failed to get usernames", "error", err)
			return ctx.Status(fiber.StatusInternalServerError).JSON(
				models.FailureResponse{Status: "error", Error: "internal server error"},
			)
		}

		usernames, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Username])
		if err != nil {
			slog.Error("failed to parse usernames", "error", err)
			return ctx.Status(fiber.StatusInternalServerError).JSON(
				models.FailureResponse{Status: "error", Error: "internal server error"},
			)
		}

		return ctx.JSON(GetUsernamesSuccessResponse{
			Status: "ok", Data: GetUsernamesSuccessResponseData{Total: total, Usernames: usernames}},
		)
	}
}
