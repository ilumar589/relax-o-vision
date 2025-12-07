package embeddings

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// Handlers handles HTTP endpoints for embeddings and search
type Handlers struct {
	service *Service
}

// NewHandlers creates new embedding handlers
func NewHandlers(service *Service) *Handlers {
	return &Handlers{
		service: service,
	}
}

// SearchTeamsRequest represents a search request for teams
type SearchTeamsRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

// SearchMatchesRequest represents a search request for matches
type SearchMatchesRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

// SearchTeams searches for teams by semantic similarity
func (h *Handlers) SearchTeams(c *fiber.Ctx) error {
	var req SearchTeamsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Limit == 0 {
		req.Limit = 10
	}

	teams, err := h.service.SearchSimilarTeams(c.Context(), req.Query, req.Limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"teams": teams,
		"count": len(teams),
	})
}

// FindSimilarTeams finds teams similar to a given team
func (h *Handlers) FindSimilarTeams(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid team ID",
		})
	}

	limitStr := c.Query("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	teams, err := h.service.FindSimilarTeam(c.Context(), id, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"teams": teams,
		"count": len(teams),
	})
}
