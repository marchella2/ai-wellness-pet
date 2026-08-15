# AI Wellness Pet API

Backend REST API for the **"AI Wellness Pet"** hackathon application. Built with **Go**, **Fiber v2**, **GORM** (PostgreSQL / Supabase), and **Google Gemini AI**.

The API lets users set up a virtual pet, log daily wellness activities (water intake, sleep, journaling), and receive an empathetic AI response from their pet persona. A built-in **Logic Engine** converts daily activity into health and energy scores that determine the pet's current state (Happy, Neutral, Tired, Sad).

> 💡 **Tip:** An Indonesian version of this document is available at [`README.id.md`](./README.id.md).

## Table of Contents

- [Features](#features)
- [Tech Stack](#tech-stack)
- [Getting Started](#getting-started)
- [Project Structure](#project-structure)
- [API Endpoints](#api-endpoints)
- [Logic Engine Rules](#logic-engine-rules)
- [Error Handling](#error-handling)
- [Deployment (Render)](#deployment-render)
- [Postman Collection](#postman-collection)
- [Testing](#testing)

## Features

- **User Registration** — register a user with `POST /api/v1/user/register` before setting up a pet.
- **Pet Onboarding** — create or rename a pet with `POST /api/v1/pet/setup`.
- **Core Loop** — log daily activities and get updated pet scores + an empathetic AI message.
- **AI Companion (Milo)** — Gemini (`gemini-1.5-flash`) generates a short, empathetic reply based on the pet's state and the owner's journal. A fallback message is returned when the AI API errors or is rate-limited.
- **Demo Utilities** — reset pet state or fast-forward to a "neglected/sad" state for pitching demos.
- **Activity History** — the latest 10 daily logs per user for the frontend dashboard.
- **CORS Enabled** — `Access-Control-Allow-Origin: *` for any frontend origin.

## Tech Stack

| Layer | Technology |
| --- | --- |
| Language | Go |
| Web Framework | Fiber v2 |
| ORM | GORM |
| Database | PostgreSQL (Supabase) |
| AI | Google Gemini AI (`gemini-1.5-flash`) |
| Environment | `github.com/joho/godotenv` |

## Getting Started

### Prerequisites

- Go 1.21+ (tested with 1.26)
- A PostgreSQL database (e.g., Supabase)
- A Google Gemini API key

### 1. Clone & Install

```bash
git clone <repository-url>
cd pet-wellness-backend
go mod download
```

### 2. Configure Environment

```bash
cp .env.example .env
```

Fill in the values:

```dotenv
PORT=8080
DATABASE_URL=postgres://user:password@host:5432/dbname?sslmode=require
GEMINI_API_KEY=your_gemini_api_key_here
```

### 3. Create the Database Schema

Run this SQL once on your Supabase/PostgreSQL instance:

```sql
-- 1. Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 2. Pets table
CREATE TABLE pets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE UNIQUE,
    pet_name VARCHAR(50) NOT NULL DEFAULT 'Milo',
    health_score INT NOT NULL DEFAULT 50 CHECK (health_score BETWEEN 0 AND 100),
    energy_score INT NOT NULL DEFAULT 50 CHECK (energy_score BETWEEN 0 AND 100),
    current_state VARCHAR(20) NOT NULL DEFAULT 'Neutral',
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 3. Daily logs table
CREATE TABLE daily_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    water_glasses INT DEFAULT 0,
    sleep_hours NUMERIC(3,1) DEFAULT 0,
    journal_text TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

> **Note:** `AutoMigrate` only adds missing columns — it does **not** create or drop constraints, so your existing schema stays untouched.

### 4. Run the Server

```bash
go run .
```

The API starts on `http://localhost:8080` (or the `PORT` you set).

### 5. Run Tests

```bash
go test ./...
```

## Project Structure

```text
pet-wellness-backend/
├── main.go                        # Entry point, middleware, router wiring
├── config/
│   ├── env.go                     # Load environment variables (.env)
│   └── database.go                # GORM + PostgreSQL connection, AutoMigrate
├── models/
│   ├── user.go                    # User model
│   ├── pet.go                     # Pet model + score/state constants
│   └── daily_log.go               # DailyLog model
├── services/                      # Business logic (one struct per endpoint)
│   ├── health_service.go          # HealthService
│   ├── user_service.go            # UserService (register)
│   ├── pet_service.go             # PetService (read, setup, reset, simulate)
│   ├── activity_service.go        # ActivityService (log + logic engine + AI)
│   ├── logic_engine.go            # Score calculation & mood determination
│   ├── ai_service.go              # Gemini integration (persona Milo)
│   └── *_test.go                  # Unit tests
├── controllers/                   # HTTP handlers (one struct per endpoint)
│   ├── health_controller.go       # HealthController
│   ├── user_controller.go         # UserController
│   ├── pet_controller.go          # PetController
│   └── activity_controller.go     # ActivityController
├── routes/                        # Route registration (one struct per endpoint)
│   ├── health_router.go           # GET /health
│   ├── user_router.go             # /user routes
│   ├── pet_router.go              # /pet routes
│   └── activity_router.go         # /activity routes
└── postman/                       # Postman collection
    └── AI-Wellness-Pet.postman_collection.json
```

## API Endpoints

| Method | Endpoint | Description |
| --- | --- | --- |
| `GET` | `/health` | Server health check |
| `POST` | `/api/v1/user/register` | Register a new user (idempotent by email) |
| `GET` | `/api/v1/pet/:user_id` | Get the latest pet of a user |
| `POST` | `/api/v1/pet/setup` | Create or rename a pet (onboarding) |
| `POST` | `/api/v1/pet/:user_id/reset` | Reset pet to default state (50/50/Neutral) |
| `POST` | `/api/v1/pet/:user_id/simulate-neglect` | Force pet into neglected state (20/20/Sad) |
| `POST` | `/api/v1/activity` | Log daily activity (core loop) |
| `GET` | `/api/v1/activity/:user_id` | Get the 10 latest daily logs |

### Health Check

```http
GET /health
```

**Response — 200 OK**

```json
{
    "status": "ok"
}
```

### Register User

```http
POST /api/v1/user/register
Content-Type: application/json
```

**Request body**

```json
{
    "name": "Jane Doe",
    "email": "jane@example.com"
}
```

- `name` and `email` are **mandatory** (missing either → `400`).
- If the email is not yet registered, a new user is created.
- If the email is already registered, the endpoint is **idempotent**: it returns the existing user instead of erroring.

**Response — 201 Created** (new user)

```json
{
    "status": "success",
    "message": "User registered successfully",
    "user": {
        "id": "11111111-1111-1111-1111-111111111111",
        "name": "Jane Doe",
        "email": "jane@example.com",
        "created_at": "2026-08-16T12:00:00+07:00"
    }
}
```

**Response — 200 OK** (email already registered)

```json
{
    "status": "success",
    "message": "User already registered",
    "user": {
        "id": "11111111-1111-1111-1111-111111111111",
        "name": "Jane Doe",
        "email": "jane@example.com",
        "created_at": "2026-08-16T12:00:00+07:00"
    }
}
```

**Error — 400** `{"status": "error", "message": "name is required"}` or `{"status": "error", "message": "email is required"}`

**Error — 500** `{"status": "error", "message": "..."}` on unexpected DB failure.

### Setup Pet (Onboarding)

```http
POST /api/v1/pet/setup
Content-Type: application/json
```

**Request body**

```json
{
    "user_id": "1f4dba8f-07c9-4aa3-9738-5b1c9ec18573",
    "pet_name": "Dudu"
}
```

- `pet_name` is **mandatory** (empty → `400 pet_name is required`).
- If the user has no pet yet, a new pet is created with `health_score: 50`, `energy_score: 50`, `current_state: "Neutral"`.
- If the pet already exists, only `pet_name` is updated.
- The `user_id` must already exist — call `POST /api/v1/user/register` first to obtain one.

**Response — 200 OK**

```json
{
    "status": "success",
    "message": "Pet initialized successfully",
    "pet": {
        "id": "4d8fdaba-50d3-4e89-9659-287e534e3e9a",
        "user_id": "1f4dba8f-07c9-4aa3-9738-5b1c9ec18573",
        "pet_name": "Dudu",
        "health_score": 50,
        "energy_score": 50,
        "current_state": "Neutral",
        "created_at": "2026-08-16T01:06:34.271762+07:00",
        "updated_at": "2026-08-16T01:06:34.271762+07:00"
    }
}
```

**Error — 404 user not registered**

If the `user_id` does not exist in the `users` table (foreign key violation):

```json
{
    "status": "error",
    "message": "user not registered, please register first"
}
```

### Get Pet by User ID

```http
GET /api/v1/pet/1f4dba8f-07c9-4aa3-9738-5b1c9ec18573
```

**Response — 200 OK**

```json
{
    "status": "success",
    "pet": {
        "id": "4d8fdaba-50d3-4e89-9659-287e534e3e9a",
        "user_id": "1f4dba8f-07c9-4aa3-9738-5b1c9ec18573",
        "pet_name": "Dudu",
        "health_score": 50,
        "energy_score": 50,
        "current_state": "Neutral",
        "created_at": "2026-08-16T01:06:34.271762+07:00",
        "updated_at": "2026-08-16T01:06:34.271762+07:00"
    }
}
```

**Error — 404** `{"status": "error", "message": "pet not found"}`

### Create Activity (Core Loop)

```http
POST /api/v1/activity
Content-Type: application/json
```

**Request body**

```json
{
    "user_id": "1f4dba8f-07c9-4aa3-9738-5b1c9ec18573",
    "water_glasses": 4,
    "sleep_hours": 7.5,
    "journal_text": "Hari ini aku sangat senang bisa jalan pagi"
}
```

**Flow**

1. Validates that the pet exists (must be set up first).
2. Stores the record in `daily_logs`.
3. Runs the **Logic Engine** to recalculate scores and determine the state.
4. Updates the pet in the `pets` table.
5. Calls Gemini AI (`gemini-1.5-flash`) for an empathetic reply from the pet persona based on the new state and the journal text. Falls back to a static message on AI errors.

**Response — 200 OK**

```json
{
    "status": "success",
    "pet": {
        "id": "4d8fdaba-50d3-4e89-9659-287e534e3e9a",
        "user_id": "1f4dba8f-07c9-4aa3-9738-5b1c9ec18573",
        "pet_name": "Dudu",
        "health_score": 75,
        "energy_score": 85,
        "current_state": "Happy",
        "created_at": "2026-08-16T01:06:34.271762+07:00",
        "updated_at": "2026-08-16T01:06:34.284820+07:00"
    },
    "ai_message": "Milo senang kamu sudah mencatat aktivitas hari ini! Jangan lupa tetap minum air yang cukup dan tidur teratur ya."
}
```

**Error — 404** (pet not set up yet)

```json
{
    "status": "error",
    "message": "pet not found, please set up your pet first"
}
```

### Reset Pet State

```http
POST /api/v1/pet/1f4dba8f-07c9-4aa3-9738-5b1c9ec18573/reset
```

Restores the pet to `health_score: 50`, `energy_score: 50`, `current_state: "Neutral"`. **Daily logs are not deleted.**

**Response — 200 OK**

```json
{
    "status": "success",
    "message": "Pet state reset to default",
    "pet": {
        "id": "4d8fdaba-50d3-4e89-9659-287e534e3e9a",
        "user_id": "1f4dba8f-07c9-4aa3-9738-5b1c9ec18573",
        "pet_name": "Dudu",
        "health_score": 50,
        "energy_score": 50,
        "current_state": "Neutral"
    }
}
```

### Simulate Neglect

```http
POST /api/v1/pet/1f4dba8f-07c9-4aa3-9738-5b1c9ec18573/simulate-neglect
```

Fast-forward simulation for demos: sets the pet to `health_score: 20`, `energy_score: 20`, `current_state: "Sad"`.

**Response — 200 OK**

```json
{
    "status": "success",
    "message": "Pet is now neglected/sad",
    "pet": {
        "id": "4d8fdaba-50d3-4e89-9659-287e534e3e9a",
        "user_id": "1f4dba8f-07c9-4aa3-9738-5b1c9ec18573",
        "pet_name": "Dudu",
        "health_score": 20,
        "energy_score": 20,
        "current_state": "Sad"
    }
}
```

### Get Activity History

```http
GET /api/v1/activity/1f4dba8f-07c9-4aa3-9738-5b1c9ec18573
```

Returns the 10 latest daily logs ordered by `created_at DESC`.

**Response — 200 OK**

```json
{
    "status": "success",
    "data": [
        {
            "id": "4d358efd-713e-455d-8c7b-ea37d18632ca",
            "user_id": "1f4dba8f-07c9-4aa3-9738-5b1c9ec18573",
            "water_glasses": 4,
            "sleep_hours": 7.5,
            "journal_text": "Hari ini senang",
            "created_at": "2026-08-16T01:06:34.284802+07:00"
        }
    ]
}
```

## Logic Engine Rules

Scores start from the pet's current values and are updated on every activity:

| Condition | Effect |
| --- | --- |
| `water_glasses >= 4` | `health_score + 15` |
| `sleep_hours >= 7.0` | `energy_score + 25` |
| `journal_text` not empty | `health_score + 10`, `energy_score + 10` |

- Scores are **clamped** between `0` and `100`.
- The state is determined in the following priority order:

| Priority | Condition | State |
| --- | --- | --- |
| 1 | `(health + energy) / 2 >= 75` | `Happy` |
| 2 | `energy < 35` | `Tired` |
| 3 | `health < 40` | `Sad` |
| 4 | otherwise | `Neutral` |

## Error Handling

| HTTP Status | Scenario |
| --- | --- |
| `400` | Invalid JSON body, missing `user_id`, missing `pet_name`, or missing `name`/`email` on registration |
| `404` | Pet not found, or user not registered (FK violation during setup) |
| `500` | Database/processing failure (logged to the server console) |

> **Note:** `POST /api/v1/user/register` returns `200` (not an error) when the email is already registered — it's treated as idempotent and returns the existing user rather than a `409 Conflict`.

## Deployment (Render)

1. Push the repository to GitHub.
2. In Render, create a **New Web Service** and connect the repository.
3. Configure:

| Setting | Value |
| --- | --- |
| Build Command | `go build -o pet-wellness-backend` |
| Start Command | `./pet-wellness-backend` |

4. Add the environment variables:

| Variable | Example |
| --- | --- |
| `PORT` | `8080` |
| `DATABASE_URL` | `postgres://user:password@host:5432/dbname?sslmode=require` |
| `GEMINI_API_KEY` | `AIzaSy...` |

> `.env` is git-ignored; set all secrets via the Render dashboard.

## Postman Collection

Import `postman/AI-Wellness-Pet.postman_collection.json` into Postman. The collection includes variables (`base_url`, `user_id`, `pet_name`) and example responses for every endpoint.

## Testing

Unit tests cover the Logic Engine (`logic_engine_test.go`), the FK-violation detection (`pet_service_test.go`), and the unique-violation detection for user registration (`user_service_test.go`).

```bash
go test ./...
```

## License

This project was built for a hackathon demo. No license is specified.