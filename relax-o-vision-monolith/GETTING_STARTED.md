# Getting Started with Relax-O-Vision

This guide will help you get the Relax-O-Vision application up and running on your local machine.

## Table of Contents

- [Prerequisites](#prerequisites)
- [API Keys Setup](#api-keys-setup)
- [Local Development Setup](#local-development-setup)
- [Docker Setup](#docker-setup)
- [Common Operations](#common-operations)
- [Architecture Overview](#architecture-overview)
- [Troubleshooting](#troubleshooting)

## Prerequisites

Before you begin, ensure you have the following installed on your system:

- **Go 1.25+** - Download from [https://golang.org/dl/](https://golang.org/dl/)
- **Docker and Docker Compose** - Download from [https://www.docker.com/get-started](https://www.docker.com/get-started)
- **Node.js/Bun** (for frontend assets) - Download from [https://bun.sh/](https://bun.sh/)
- **Git** - Download from [https://git-scm.com/downloads](https://git-scm.com/downloads)

### Optional Tools

- **Air** (for hot reload during development) - Install with: `go install github.com/air-verse/air@latest`
- **Templ** (for template generation) - Install with: `go install github.com/a-h/templ/cmd/templ@latest`
- **golangci-lint** (for code linting) - Install from [https://golangci-lint.run/usage/install/](https://golangci-lint.run/usage/install/)

## API Keys Setup

This application requires API keys from various services. Follow these steps to obtain them:

### 1. Football Data API (Required)

The football data API provides match, team, and competition information.

1. Visit [https://www.football-data.org/](https://www.football-data.org/)
2. Sign up for a free account
3. Navigate to your account dashboard
4. Copy your API key
5. **Note:** Free tier allows 10 requests per minute

### 2. OpenAI API (Required for Predictions)

OpenAI powers the match prediction feature.

1. Visit [https://platform.openai.com/](https://platform.openai.com/)
2. Sign up or log in
3. Navigate to API Keys section
4. Create a new API key
5. Copy and save it securely

### 3. Claude API (Optional)

Claude by Anthropic can be used as an alternative LLM provider.

1. Visit [https://console.anthropic.com/](https://console.anthropic.com/)
2. Sign up or log in
3. Navigate to API Keys
4. Create a new API key
5. Copy and save it securely

### 4. Gemini API (Optional)

Google's Gemini can also be used as an LLM provider.

1. Visit [https://makersuite.google.com/](https://makersuite.google.com/)
2. Sign up or log in
3. Get your API key
4. Copy and save it securely

## Local Development Setup

Follow these steps to run the application locally:

### 1. Clone the Repository

```bash
git clone https://github.com/ilumar589/relax-o-vision.git
cd relax-o-vision/relax-o-vision-monolith
```

### 2. Set Up Environment Variables

```bash
# Copy the example environment file
cp .env.example .env

# Edit the .env file with your actual API keys
# Use your preferred text editor (nano, vim, code, etc.)
nano .env
```

Fill in your API keys in the `.env` file:

```env
FOOTBALL_DATA_API_KEY=your_actual_football_data_api_key
OPENAI_API_KEY=your_actual_openai_api_key
CLAUDE_API_KEY=your_actual_claude_api_key  # Optional
GEMINI_API_KEY=your_actual_gemini_api_key  # Optional
```

### 3. Start Dependencies (PostgreSQL and Redis)

```bash
# Start only the database and cache services
docker-compose up -d postgres redis

# Verify services are running
docker-compose ps
```

### 4. Install Go Dependencies

```bash
go mod download
```

### 5. Run the Application

**Option A: Run directly with Go**

```bash
go run .
```

**Option B: Use Air for hot reload (recommended for development)**

```bash
# Make sure Air is installed
go install github.com/air-verse/air@latest

# Start with Air
air
```

### 6. Access the Application

Open your browser and navigate to:

```
http://localhost:7000
```

## Docker Setup

To run the entire application stack with Docker:

### Start All Services

```bash
# Build and start all services (app, database, Redis)
docker-compose up -d

# View logs
docker-compose logs -f

# View logs for a specific service
docker-compose logs -f gowebly_fiber
```

### Rebuild the Application

After making code changes:

```bash
# Rebuild and restart the app service
docker-compose up -d --build gowebly_fiber
```

### Stop All Services

```bash
# Stop all services but keep data
docker-compose down

# Stop all services and remove volumes (resets database)
docker-compose down -v
```

## Common Operations

### View Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f gowebly_fiber
docker-compose logs -f postgres
docker-compose logs -f redis
```

### Database Operations

```bash
# Connect to PostgreSQL
docker-compose exec postgres psql -U postgres -d relaxovision

# Run migrations (automatic on app start)
# Migrations are in the ./migrations folder
```

### Reset Everything

```bash
# Stop services and remove all data
docker-compose down -v

# Start fresh
docker-compose up -d
```

### Check Service Health

```bash
# List running services
docker-compose ps

# Check specific service status
docker-compose ps postgres
docker-compose ps redis
```

## Architecture Overview

### Components

- **Backend:** Go with Fiber web framework
- **Database:** PostgreSQL with pgvector extension for embeddings
- **Cache:** Redis for fast data access with 30-day TTL
- **Frontend:** htmx with Alpine.js and Tailwind CSS

### Data Flow

1. **API Requests** → Fiber handlers
2. **Cache Check** → Redis (fast) → PostgreSQL (if miss)
3. **Data Fetching** → Football Data API (if cache expired)
4. **Caching Strategy:**
   - Competition data: 30 days
   - Team data: 30 days
   - Match data: 30 days (refreshed more frequently for live matches)
   - Standings: 30 days

### Features

- **Match Predictions:** AI-powered match outcome predictions
- **Semantic Search:** Find similar teams using embeddings
- **Real-time Updates:** WebSocket support for live match updates
- **Caching:** Efficient multi-layer caching (Redis + PostgreSQL)

## Experimental Features (Go 1.25+)

This application uses experimental Go features available in Go 1.25+:

### Green Tea GC

Improves garbage collection performance with lower latency and better memory management.

- Enabled with: `GOEXPERIMENT=greenteagc`
- Set automatically in Docker
- Benefits: Reduced pause times, better throughput

### JSON v2 (`encoding/json/v2`)

Go 1.25 includes a new, experimental JSON implementation that provides significant improvements over the standard library.

- Enabled with: `GOEXPERIMENT=jsonv2`
- Set automatically in Docker
- Available packages:
  - `encoding/json/v2` - Major revision of encoding/json
  - `encoding/json/jsontext` - Lower-level JSON syntax processing

**Benefits:**
- Better performance for JSON marshaling/unmarshaling
- More consistent behavior
- Support for `omitzero` tag to omit zero values
- Better handling of maps with non-string keys
- Improved error messages

The application uses both experimental features together: `GOEXPERIMENT=greenteagc,jsonv2`

## Troubleshooting

### Port Already in Use

```bash
# If port 7000 is already in use, change BACKEND_PORT in .env
BACKEND_PORT=8000

# Or stop the conflicting service
lsof -ti:7000 | xargs kill
```

### Database Connection Issues

```bash
# Check if PostgreSQL is running
docker-compose ps postgres

# Check PostgreSQL logs
docker-compose logs postgres

# Restart PostgreSQL
docker-compose restart postgres
```

### Redis Connection Issues

```bash
# Check if Redis is running
docker-compose ps redis

# Test Redis connection
docker-compose exec redis redis-cli ping
# Should return: PONG
```

### API Rate Limiting

The free tier of football-data.org allows only 10 requests per minute:

- Application has built-in rate limiting
- Data is cached for 30 days to minimize API calls
- If you hit the limit, wait 60 seconds before retrying

### Missing Dependencies

```bash
# Clean and reinstall Go modules
go clean -modcache
go mod download

# Rebuild the application
go build
```

### Docker Build Fails

```bash
# Clean Docker cache
docker-compose down
docker system prune -a

# Rebuild from scratch
docker-compose build --no-cache
docker-compose up -d
```

## Next Steps

- Explore the API endpoints at `http://localhost:7000/api/`
- Read the main [README.md](README.md) for project overview
- Check out [FOOTBALL_PREDICTIONS.md](FOOTBALL_PREDICTIONS.md) for prediction features
- Review [IMPLEMENTATION_SUMMARY.md](IMPLEMENTATION_SUMMARY.md) for technical details

## Support

For issues or questions:

1. Check the [Troubleshooting](#troubleshooting) section
2. Review existing GitHub issues
3. Create a new issue with detailed information

---

Happy coding! 🎉
