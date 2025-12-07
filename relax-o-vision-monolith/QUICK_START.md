# Quick Start Guide

Get started with the Football Data Ingestion and Win Predictions modules in 5 minutes!

## Prerequisites

- Go 1.24 or later
- Docker and Docker Compose
- API Keys (optional for testing):
  - Football Data API: https://www.football-data.org/client/register
  - OpenAI API: https://platform.openai.com/api-keys

## Quick Start

### 1. Start the Infrastructure

```bash
# Start PostgreSQL and Redis
docker-compose up -d postgres redis

# Wait for PostgreSQL to be ready (about 10 seconds)
docker-compose logs -f postgres
# Press Ctrl+C when you see "database system is ready to accept connections"
```

### 2. Set Environment Variables

```bash
# Copy and edit the environment file
cat > .env << 'EOF'
DATABASE_URL=postgres://postgres:postgres@localhost:5432/relaxovision?sslmode=disable
FOOTBALL_DATA_API_KEY=YOUR_FOOTBALL_DATA_API_KEY_HERE
OPENAI_API_KEY=YOUR_OPENAI_API_KEY_HERE
EOF

# Load environment variables
export $(cat .env | xargs)
```

### 3. Build and Run

```bash
# Build the application
go build

# Run the application
./relaxovisionmonolith
```

The application will:
- Connect to PostgreSQL
- Run database migrations automatically
- Start the web server on port 7000

### 4. Test the API

```bash
# Test the basic endpoint
curl http://localhost:7000/api/hello-world

# Test football data endpoint (requires actual data in DB)
curl http://localhost:7000/api/football/competitions/2021

# Create a prediction (requires actual match data)
curl -X POST http://localhost:7000/api/predictions \
  -H "Content-Type: application/json" \
  -d '{"matchId": 123456}'
```

## Using with Docker Compose

```bash
# Build and start all services
docker-compose up --build

# Access the application
curl http://localhost:7000/

# Stop all services
docker-compose down
```

## Next Steps

### 1. Populate Football Data

To sync data from football-data.org, uncomment the scheduler in `server.go`:

```go
// Around line 110 in server.go
competitionCodes := []string{"PL", "PD", "BL1"}
scheduler := footballdata.NewScheduler(footballService, competitionCodes, 24*time.Hour)
go scheduler.Start(context.Background())
```

Rebuild and restart the application.

### 2. Test Predictions

Once you have match data, create predictions:

```bash
# Get matches first
curl http://localhost:7000/api/football/matches/123456

# Create a prediction for that match
curl -X POST http://localhost:7000/api/predictions \
  -H "Content-Type: application/json" \
  -d '{"matchId": 123456}'

# Check the prediction
curl http://localhost:7000/api/predictions/{prediction-id}
```

### 3. Explore the Database

```bash
# Connect to PostgreSQL
docker exec -it relax-o-vision-monolith-postgres-1 psql -U postgres -d relaxovision

# List tables
\dt

# View competitions
SELECT id, name, code FROM competitions;

# View predictions
SELECT match_id, home_win_prob, draw_prob, away_win_prob, confidence FROM predictions;

# Exit
\q
```

## Troubleshooting

### Database Connection Failed

Make sure PostgreSQL is running:
```bash
docker-compose ps postgres
docker-compose logs postgres
```

### Migrations Failed

Migrations are idempotent, you can safely restart the application.

Check migration status:
```sql
SELECT * FROM schema_migrations;
```

### API Keys Not Working

1. Verify your keys are correct in `.env`
2. Make sure environment variables are loaded:
   ```bash
   echo $FOOTBALL_DATA_API_KEY
   echo $OPENAI_API_KEY
   ```
3. Restart the application after changing keys

### Rate Limiting

The football-data.org free tier allows 10 requests per minute. The client handles this automatically with rate limiting.

## Development Tips

### Live Reloading

If you have Air installed:
```bash
air
```

### View Logs

```bash
# Application logs
docker-compose logs -f gowebly_fiber

# Database logs
docker-compose logs -f postgres

# All logs
docker-compose logs -f
```

### Reset Database

```bash
# Stop and remove containers
docker-compose down -v

# Start fresh
docker-compose up -d postgres
```

## API Reference

See `FOOTBALL_PREDICTIONS.md` for complete API documentation.

## Architecture

See `IMPLEMENTATION_SUMMARY.md` for technical details.

---

**Happy Coding!** 🚀
