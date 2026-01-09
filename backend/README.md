# Mamonedz - Personal Finance Tracker API

Backend REST API untuk aplikasi personal finance tracker menggunakan Go (Gin) dan PostgreSQL.

## Tech Stack

- Go 1.21+
- Gin Framework
- PostgreSQL
- GORM

## Setup

1. Copy environment file:
```bash
cp .env.example .env
```

2. Update `.env` dengan konfigurasi database Anda

3. Install dependencies:
```bash
go mod tidy
```

4. Run server:
```bash
go run cmd/server/main.go
```

## API Endpoints

Base URL: `/api/v1`

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /health | Health check |
| GET | /expenses | Get all expenses (with filters) |
| GET | /expenses/:id | Get expense by ID |
| POST | /expenses | Create new expense |
| PUT | /expenses/:id | Update expense |
| DELETE | /expenses/:id | Delete expense |
| GET | /expenses/stats | Get statistics |

## Query Parameters

### GET /expenses
- `start_date` - Filter by start date (YYYY-MM-DD)
- `end_date` - Filter by end date (YYYY-MM-DD)
- `category` - Filter by category
- `limit` - Pagination limit (default: 10)
- `offset` - Pagination offset (default: 0)

### GET /expenses/stats
- `period` - day | week | month (default: month)

## Valid Categories

- makanan
- transportasi
- hiburan
- belanja
- kesehatan
- pendidikan
- lainnya
