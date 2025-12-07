package predictions

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// Handlers contains HTTP handlers for predictions
type Handlers struct {
	service *Service
}

// NewHandlers creates a new handlers instance
func NewHandlers(service *Service) *Handlers {
	return &Handlers{service: service}
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
