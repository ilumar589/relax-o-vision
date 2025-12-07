# Football Data & Predictions Modules

This document describes the new football data ingestion and AI-powered win predictions modules added to the relax-o-vision-monolith application.

## Overview

Two new modules have been added:

1. **Football Data Ingestion** (`footballdata/`) - Fetches and stores data from football-data.org API
2. **Win Predictions** (`predictions/`) - AI-powered match predictions using OpenAI agents

## Database Setup

### Prerequisites

- PostgreSQL with pgvector extension
- The database migrations will be run automatically on application startup

### Migrations

Migrations are located in the `migrations/` directory:
- `001_create_competitions.sql` - Creates competitions table
- `002_create_teams.sql` - Creates teams table
- `003_create_matches.sql` - Creates matches table
- `004_create_predictions.sql` - Creates predictions table

All tables include:
- JSONB columns for flexible nested data
- Vector columns (pgvector) for future semantic search capabilities
- Appropriate indexes for performance

## Configuration

### Environment Variables

Set the following environment variables:

```bash
# Database connection
DATABASE_URL=postgres://postgres:postgres@localhost:5432/relaxovision?sslmode=disable

# API Keys
FOOTBALL_DATA_API_KEY=your_football_data_api_key_here
OPENAI_API_KEY=your_openai_api_key_here
```

### Dapr Secrets (Optional)

If using Dapr, configure secrets in `dapr/secrets.json`:

```json
{
  "football-data-api-key": "your_football_data_api_key_here",
  "openai-api-key": "your_openai_api_key_here"
}
```

A template file is provided at `dapr/secrets.json.template`.

## Football Data Module

### Features

- HTTP client for football-data.org API with rate limiting
- Support for fetching competitions, teams, matches, and standings
- PostgreSQL storage with JSONB and vector columns
- Background scheduler for periodic data synchronization

### API Endpoints

#### Get Competition
```
GET /api/football/competitions/:id
```

#### Get Team
```
GET /api/football/teams/:id
```

#### Get Match
```
GET /api/football/matches/:id
```

### Background Sync

To enable automatic data synchronization, uncomment the scheduler code in `server.go`:

```go
competitionCodes := []string{"PL", "PD", "BL1"} // Premier League, La Liga, Bundesliga
scheduler := footballdata.NewScheduler(footballService, competitionCodes, 24*time.Hour)
go scheduler.Start(context.Background())
```

## Predictions Module

### Features

- Multiple AI agents for different analysis perspectives:
  - **Statistical Agent** - Analyzes historical statistics
  - **Form Agent** - Evaluates recent team form
  - **Head-to-Head Agent** - Analyzes historical matchups
  - **Aggregator Agent** - Combines insights from other agents

- Dapr workflow orchestration (optional)
- PostgreSQL storage for predictions

### API Endpoints

#### Create Prediction
```
POST /api/predictions
Content-Type: application/json

{
  "matchId": 123456
}
```

#### Get Prediction
```
GET /api/predictions/:id
```

#### Get Match Predictions
```
GET /api/predictions/match/:matchId
```

### Response Format

```json
{
  "id": "uuid",
  "matchId": 123456,
  "homeWinProb": 0.45,
  "drawProb": 0.30,
  "awayWinProb": 0.25,
  "confidence": 0.75,
  "status": "completed",
  "reasoning": "Based on statistical analysis and recent form...",
  "agentOutputs": [
    {
      "agentType": "statistical",
      "homeWinProb": 0.48,
      "drawProb": 0.28,
      "awayWinProb": 0.24,
      "confidence": 0.80,
      "reasoning": "...",
      "keyFactors": ["Home advantage", "Better goal difference"]
    }
  ],
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z"
}
```

## Running with Docker Compose

The updated `docker-compose.yml` includes:
- PostgreSQL with pgvector extension
- Redis for Dapr pubsub (optional)
- Application service with database connectivity

```bash
docker-compose up -d
```

## Architecture

### Football Data Flow
1. API Client fetches data from football-data.org
2. Service layer processes and validates data
3. Repository layer stores in PostgreSQL
4. Scheduler runs periodic syncs (optional)

### Predictions Flow
1. Request received for match prediction
2. Match data fetched from database
3. Multiple AI agents analyze the match in parallel
4. Aggregator agent combines insights
5. Final prediction saved to database and returned

## Future Enhancements

- Support for multiple LLM providers (Claude, Gemini, etc.)
- Real-time prediction updates via WebSockets
- Enhanced statistical models
- Semantic search using pgvector embeddings
- Prediction accuracy tracking and model improvements

## Rate Limits

- **football-data.org free tier**: 10 requests per minute
- The client includes automatic rate limiting to respect this constraint

## Notes

- Placeholder API keys are used by default if environment variables are not set
- The application will log warnings if API keys are not configured
- Migrations are idempotent and safe to run multiple times
- Vector embeddings are prepared but not currently populated (future enhancement)
