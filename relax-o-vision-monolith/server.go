package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/edd/relaxovisionmonolith/cache"
	"github.com/edd/relaxovisionmonolith/embeddings"
	"github.com/edd/relaxovisionmonolith/footballdata"
	"github.com/edd/relaxovisionmonolith/predictions"
	"github.com/edd/relaxovisionmonolith/predictions/providers"
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
	embeddingsService   *embeddings.Service
	embeddingsHandlers  *embeddings.Handlers
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

	// Prediction accuracy endpoints
	server.Get("/api/predictions/accuracy", predictionsHandlers.GetAccuracyStats)
	server.Get("/api/predictions/accuracy/competition/:id", predictionsHandlers.GetCompetitionAccuracy)
	server.Get("/api/predictions/leaderboard", predictionsHandlers.GetLeaderboard)

	// Semantic search endpoints
	server.Post("/api/search/teams", embeddingsHandlers.SearchTeams)
	server.Get("/api/teams/:id/similar", embeddingsHandlers.FindSimilarTeams)

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

	claudeKey := os.Getenv("CLAUDE_API_KEY")
	if claudeKey == "" {
		claudeKey = "YOUR_CLAUDE_API_KEY_HERE"
		slog.Warn("CLAUDE_API_KEY not set, using placeholder")
	}

	geminiKey := os.Getenv("GEMINI_API_KEY")
	if geminiKey == "" {
		geminiKey = "YOUR_GEMINI_API_KEY_HERE"
		slog.Warn("GEMINI_API_KEY not set, using placeholder")
	}

	// Initialize cache (use memory cache for simplicity)
	cacheConfig := cache.CacheConfig{
		Type:    "memory",
		MaxSize: 1000,
	}
	cacheImpl, err := cache.NewCache(cacheConfig)
	if err != nil {
		slog.Error("Failed to initialize cache, using memory cache", "error", err)
		cacheImpl = cache.NewMemoryCache(1000)
	}

	// Initialize football data service with caching
	footballClient := footballdata.NewClient(footballAPIKey)
	cachedClient := footballdata.NewCachedClient(footballAPIKey, cacheImpl)
	_ = cachedClient // Available for future use
	footballRepo := footballdata.NewRepository(db)
	footballService = footballdata.NewService(footballClient, footballRepo)

	// Initialize LLM providers for predictions and embeddings
	providerConfigs := []providers.ProviderConfig{
		{
			Name:    "openai",
			APIKey:  openAIKey,
			Model:   "gpt-4",
			Enabled: openAIKey != "YOUR_OPENAI_API_KEY_HERE",
			Weight:  1.0,
		},
		{
			Name:    "claude",
			APIKey:  claudeKey,
			Model:   "claude-3-5-sonnet-20241022",
			Enabled: false, // Disabled by default, can be enabled with valid key
			Weight:  1.0,
		},
		{
			Name:    "gemini",
			APIKey:  geminiKey,
			Model:   "gemini-1.5-pro",
			Enabled: false, // Disabled by default, can be enabled with valid key
			Weight:  1.0,
		},
	}

	factory := providers.NewProviderFactory(providerConfigs)
	llmProviders, err := factory.CreateProviders()
	if err != nil {
		slog.Error("Failed to create LLM providers", "error", err)
		// Fallback to just OpenAI
		llmProviders = []providers.LLMProvider{
			providers.NewOpenAIProvider(openAIKey, "gpt-4"),
		}
	}

	// Initialize predictions service
	predictionsService = predictions.NewService(db, openAIKey)
	predictionsHandlers = predictions.NewHandlers(predictionsService)

	// Initialize embeddings service
	embeddingsService = embeddings.NewService(db, llmProviders)
	embeddingsHandlers = embeddings.NewHandlers(embeddingsService)

	// Optional: Start embedding worker in background
	// embeddingsWorker := embeddings.NewWorker(embeddingsService, db, footballService)
	// go embeddingsWorker.Start(context.Background())

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
