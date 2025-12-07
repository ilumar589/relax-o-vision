package predictions

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// Handlers contains HTTP handlers for predictions
type Handlers struct {
	service         *Service
	accuracyService *AccuracyService
}

// NewHandlers creates a new handlers instance
func NewHandlers(service *Service) *Handlers {
	return &Handlers{
		service:         service,
		accuracyService: NewAccuracyService(service.db),
	}
}

// CreatePrediction handles POST /api/predictions
func (h *Handlers) CreatePrediction(c *fiber.Ctx) error {
	var req PredictionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.MatchID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid match ID",
		})
	}

	prediction, err := h.service.CreatePrediction(c.Context(), req.MatchID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(prediction)
}

// GetPrediction handles GET /api/predictions/:id
func (h *Handlers) GetPrediction(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Prediction ID is required",
		})
	}

	prediction, err := h.service.GetPrediction(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(prediction)
}

// GetMatchPredictions handles GET /api/predictions/match/:matchId
func (h *Handlers) GetMatchPredictions(c *fiber.Ctx) error {
	matchIDStr := c.Params("matchId")
	matchID, err := strconv.Atoi(matchIDStr)
	if err != nil || matchID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid match ID",
		})
	}

	predictions, err := h.service.GetPredictionsByMatch(c.Context(), matchID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"predictions": predictions,
		"count":       len(predictions),
	})
}

// GetAccuracyStats handles GET /api/predictions/accuracy
func (h *Handlers) GetAccuracyStats(c *fiber.Ctx) error {
	stats, err := h.accuracyService.GetOverallStats(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(stats)
}

// GetCompetitionAccuracy handles GET /api/predictions/accuracy/competition/:id
func (h *Handlers) GetCompetitionAccuracy(c *fiber.Ctx) error {
	compIDStr := c.Params("id")
	compID, err := strconv.Atoi(compIDStr)
	if err != nil || compID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid competition ID",
		})
	}

	stats, err := h.accuracyService.GetCompetitionStats(c.Context(), compID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(stats)
}

// GetLeaderboard handles GET /api/predictions/leaderboard
func (h *Handlers) GetLeaderboard(c *fiber.Ctx) error {
	leaderboard, err := h.accuracyService.GetLeaderboard(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"leaderboard": leaderboard,
		"count":       len(leaderboard),
	})
}
