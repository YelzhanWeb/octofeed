# RSSHub - RSS Feed Aggregator

## Overview

RSSHub is a CLI application for aggregating RSS feeds with the following features:

- Command-line interface for managing feeds
- Background worker pool for parallel feed processing
- PostgreSQL for persistent storage
- Docker Compose for easy deployment
- Hexagonal architecture for clean code organization

## Project Structure

```
rsshub/
├── cmd/rsshub/           # Application entry point
├── internal/
│   ├── domain/           # Business entities
│   ├── ports/            # Interfaces
│   ├── core/services/    # Use cases
│   └── adapters/         # Implementations
├── migrations/           # Database migrations
├── docker-compose.yml
├── Dockerfile
└── .env
```

## Prerequisites

- Go 1.23+
- Docker & Docker Compose
- PostgreSQL 16+ (if running locally)

## Quick Start

### 1. Clone and Setup

```bash
git clone <repository>
cd rsshub
cp .env.example .env
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Format Code

```bash
make install-tools
make fmt
```

### 4. Build

```bash
make build
# or
go build -o rsshub ./cmd/rsshub
```

### 5. Run with Docker

```bash
docker-compose up -d
docker exec -it rsshub_app /bin/sh
./rsshub --help
```

## Usage Examples

### Start Background Aggregator

```bash
./rsshub fetch
```

### Add Feed

```bash
./rsshub add --name "tech-crunch" --url "https://techcrunch.com/feed/"
```

### Change Interval (while running)

In another terminal:

```bash
./rsshub set-interval --duration 2m
```

### Change Workers (while running)

```bash
./rsshub set-workers --count 5
```

### List Feeds

```bash
./rsshub list --num 5
```

### Show Articles

```bash
./rsshub articles --feed-name "tech-crunch" --num 5
```

### Delete Feed

```bash
./rsshub delete --name "tech-crunch"
```

## Configuration

Edit `.env` file:

```env
CLI_APP_TIMER_INTERVAL=3m
CLI_APP_WORKERS_COUNT=3

POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=changeme
POSTGRES_DBNAME=rsshub
```

## Architecture

This project follows **Hexagonal Architecture** (Ports & Adapters):

- **Domain Layer**: Pure business logic (entities)
- **Ports**: Interfaces defining application boundaries
- **Core Services**: Use cases and business workflows
- **Adapters**: Implementations (CLI, PostgreSQL, HTTP)

Benefits:
- ✅ Testable
- ✅ Technology-agnostic core
- ✅ Easy to swap implementations
- ✅ Clear dependency direction

## Testing

```bash
go test -v -race ./...
```

## Sample RSS Feeds

- TechCrunch: https://techcrunch.com/feed/
- Hacker News: https://news.ycombinator.com/rss
- BBC News: https://feeds.bbci.co.uk/news/world/rss.xml
- The Verge: https://www.theverge.com/rss/index.xml

## License

MIT