package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/edd/relaxovisionmonolith/footballdata"
	"github.com/edd/relaxovisionmonolith/predictions"
	"github.com/edd/relaxovisionmonolith/websocket"
	fiberws "github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"

	gowebly "github.com/gowebly/helpers"
)

var (
	db                  *sql.DB
	footballService     *footballdata.Service
	predictionsService  *predictions.Service
	predictionsHandlers *predictions.Handlers
	wsHub               *websocket.Hub
	wsHandler           *websocket.Handler
)

// runServer runs a new HTTP server with the loaded environment variables.
func runServer() error {
	// Initialize database
	var err error
	db, err = initDatabase()
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	// Run migrations
	if err := runMigrations(db); err != nil {
		slog.Error("Failed to run migrations", "error", err)
		// Continue even if migrations fail (they might already be applied)
	}

	// Initialize services
	initServices()

	// Initialize WebSocket hub
	wsHub = websocket.NewHub()
	wsHandler = websocket.NewHandler(wsHub)
	go wsHub.Run()

	// Validate environment variables.
	port, err := strconv.Atoi(gowebly.Getenv("BACKEND_PORT", "7000"))
	if err != nil {
		return err
	}

	// Create a new server instance with options from environment variables.
	// Note: ReadTimeout and WriteTimeout are removed to support WebSocket connections
	config := fiber.Config{
		Views:       html.NewFileSystem(http.Dir("./templates"), ".html"),
		ViewsLayout: "main",
	}

	// Create a new Fiber server.
	server := fiber.New(config)

	// Add Fiber middlewares.
	server.Use(logger.New())

	// Handle static files.
	server.Static("/static", "./static")

	// Handle index page view.
	server.Get("/", indexViewHandler)

	// Handle API endpoints.
	server.Get("/api/hello-world", showContentAPIHandler)

	// Football data endpoints
	server.Get("/api/football/competitions/:id", getCompetitionHandler)
	server.Get("/api/football/teams/:id", getTeamHandler)
	server.Get("/api/football/matches/:id", getMatchHandler)

	// Prediction endpoints
	server.Post("/api/predictions", predictionsHandlers.CreatePrediction)
	server.Get("/api/predictions/:id", predictionsHandlers.GetPrediction)
	server.Get("/api/predictions/match/:matchId", predictionsHandlers.GetMatchPredictions)

	// WebSocket endpoint
	server.Use("/ws", func(c *fiber.Ctx) error {
		if fiberws.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	server.Get("/ws", fiberws.New(wsHandler.HandleConnection))

	return server.Listen(fmt.Sprintf(":%d", port))
}

// initServices initializes all application services
func initServices() {
	// Get API keys from environment
	footballAPIKey := os.Getenv("FOOTBALL_DATA_API_KEY")
	if footballAPIKey == "" {
		footballAPIKey = "YOUR_FOOTBALL_DATA_API_KEY_HERE"
		slog.Warn("FOOTBALL_DATA_API_KEY not set, using placeholder")
	}

	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		openAIKey = "YOUR_OPENAI_API_KEY_HERE"
		slog.Warn("OPENAI_API_KEY not set, using placeholder")
	}

	// Initialize football data service
	footballClient := footballdata.NewClient(footballAPIKey)
	footballRepo := footballdata.NewRepository(db)
	footballService = footballdata.NewService(footballClient, footballRepo)

	// Initialize predictions service
	predictionsService = predictions.NewService(db, openAIKey)
	predictionsHandlers = predictions.NewHandlers(predictionsService)

	// Optional: Start background scheduler for football data sync
	// Uncomment to enable automatic data synchronization
	/*
	competitionCodes := []string{"PL", "PD", "BL1"} // Premier League, La Liga, Bundesliga
	scheduler := footballdata.NewScheduler(footballService, competitionCodes, 24*time.Hour)
	go scheduler.Start(context.Background())
	*/

	slog.Info("Services initialized successfully")
}

// Football data handlers

func getCompetitionHandler(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid competition ID",
		})
	}

	competition, err := footballService.GetCompetition(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(competition)
}

func getTeamHandler(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid team ID",
		})
	}

	team, err := footballService.GetTeam(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(team)
}

func getMatchHandler(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid match ID",
		})
	}

	match, err := footballService.GetMatch(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(match)
}
