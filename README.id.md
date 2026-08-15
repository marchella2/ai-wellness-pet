# AI Wellness Pet API

Backend REST API untuk aplikasi **"AI Wellness Pet"** (hackathon). Dibangun dengan **Go**, **Fiber v2**, **GORM** (PostgreSQL / Supabase), dan **Google Gemini AI**.

API ini memungkinkan pengguna menyiapkan hewan peliharaan virtual, mencatat aktivitas harian (konsumsi air, tidur, jurnal), lalu mendapatkan respons AI yang penuh empati dari persona peliharaannya. **Logic Engine** bawaan mengubah aktivitas harian menjadi skor kesehatan dan energi yang menentukan kondisi peliharaan saat ini (Happy, Neutral, Tired, Sad).

> 💡 **Catatan:** Versi bahasa Inggris tersedia di [`README.md`](./README.md).

## Daftar Isi

- [Fitur](#fitur)
- [Tech Stack](#tech-stack)
- [Memulai](#memulai)
- [Struktur Project](#struktur-project)
- [Daftar Endpoint API](#daftar-endpoint-api)
- [Aturan Logic Engine](#aturan-logic-engine)
- [Penanganan Error](#penanganan-error)
- [Deployment (Render)](#deployment-render)
- [Postman Collection](#postman-collection)
- [Testing](#testing)

## Fitur

- **Onboarding Pet** — buat atau ganti nama pet lewat `POST /api/v1/pet/setup`.
- **Core Loop** — catat aktivitas harian dan dapatkan skor pet terbaru + pesan AI yang empatik.
- **AI Companion (Milo)** — Gemini (`gemini-1.5-flash`) membuat respons singkat dan empatik berdasarkan kondisi pet serta jurnal pemilik. Pesan fallback otomatis diberikan saat API AI error/rate-limited.
- **Utilitas Demo** — reset kondisi pet atau fast-forward ke kondisi "diabaikan/sedih" untuk demo di depan juri.
- **Riwayat Aktivitas** — 10 log aktivitas terbaru per pengguna untuk dashboard frontend.
- **CORS Aktif** — `Access-Control-Allow-Origin: *` untuk semua origin frontend.

## Tech Stack

| Layer | Teknologi |
| --- | --- |
| Bahasa | Go |
| Web Framework | Fiber v2 |
| ORM | GORM |
| Database | PostgreSQL (Supabase) |
| AI | Google Gemini AI (`gemini-1.5-flash`) |
| Environment | `github.com/joho/godotenv` |

## Memulai

### Prasyarat

- Go 1.21+ (diuji dengan 1.26)
- Database PostgreSQL (mis. Supabase)
- Google Gemini API key

### 1. Clone & Install

```bash
git clone <repository-url>
cd pet-wellness-backend
go mod download
```

### 2. Konfigurasi Environment

```bash
cp .env.example .env
```

Isi nilainya:

```dotenv
PORT=8080
DATABASE_URL=postgres://user:password@host:5432/dbname?sslmode=require
GEMINI_API_KEY=your_gemini_api_key_here
```

### 3. Buat Skema Database

Jalankan SQL ini sekali di Supabase/PostgreSQL Anda:

```sql
-- 1. Tabel Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 2. Tabel Pets
CREATE TABLE pets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE UNIQUE,
    pet_name VARCHAR(50) NOT NULL DEFAULT 'Milo',
    health_score INT NOT NULL DEFAULT 50 CHECK (health_score BETWEEN 0 AND 100),
    energy_score INT NOT NULL DEFAULT 50 CHECK (energy_score BETWEEN 0 AND 100),
    current_state VARCHAR(20) NOT NULL DEFAULT 'Neutral',
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 3. Tabel Daily Logs
CREATE TABLE daily_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    water_glasses INT DEFAULT 0,
    sleep_hours NUMERIC(3,1) DEFAULT 0,
    journal_text TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

> **Catatan:** `AutoMigrate` hanya menambahkan kolom yang belum ada — **tidak** membuat/menghapus constraint, sehingga skema yang sudah ada tidak diubah.

### 4. Jalankan Server

```bash
go run .
```

API berjalan di `http://localhost:8080` (atau sesuai `PORT` yang Anda set).

### 5. Jalankan Test

```bash
go test ./...
```

## Struktur Project

```text
pet-wellness-backend/
├── main.go                        # Entry point, middleware, wiring router
├── config/
│   ├── env.go                     # Load environment variables (.env)
│   └── database.go                # Koneksi GORM + PostgreSQL, AutoMigrate
├── models/
│   ├── user.go                    # Model User
│   ├── pet.go                     # Model Pet + konstanta skor/state
│   └── daily_log.go               # Model DailyLog
├── services/                      # Logika bisnis (satu struct per endpoint)
│   ├── health_service.go          # HealthService
│   ├── pet_service.go             # PetService (read, setup, reset, simulate)
│   ├── activity_service.go        # ActivityService (log + logic engine + AI)
│   ├── logic_engine.go            # Kalkulasi skor & penentu kondisi
│   ├── ai_service.go              # Integrasi Gemini (persona Milo)
│   └── *_test.go                  # Unit tests
├── controllers/                   # Handler HTTP (satu struct per endpoint)
│   ├── health_controller.go       # HealthController
│   ├── pet_controller.go          # PetController
│   └── activity_controller.go     # ActivityController
├── routes/                        # Registrasi route (satu struct per endpoint)
│   ├── health_router.go           # GET /health
│   ├── pet_router.go              # Route /pet
│   └── activity_router.go         # Route /activity
└── postman/                       # Postman collection
    └── AI-Wellness-Pet.postman_collection.json
```

## Daftar Endpoint API

| Method | Endpoint | Deskripsi |
| --- | --- | --- |
| `GET` | `/health` | Cek kesehatan server |
| `GET` | `/api/v1/pet/:user_id` | Ambil pet terbaru milik user |
| `POST` | `/api/v1/pet/setup` | Buat atau ganti nama pet (onboarding) |
| `POST` | `/api/v1/pet/:user_id/reset` | Reset pet ke kondisi default (50/50/Neutral) |
| `POST` | `/api/v1/pet/:user_id/simulate-neglect` | Paksa pet ke kondisi diabaikan (20/20/Sad) |
| `POST` | `/api/v1/activity` | Catat aktivitas harian (core loop) |
| `GET` | `/api/v1/activity/:user_id` | Ambil 10 log aktivitas terbaru |

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

- `pet_name` **wajib diisi** (kosong → `400 pet_name is required`).
- Jika user belum punya pet, pet baru dibuat dengan `health_score: 50`, `energy_score: 50`, `current_state: "Neutral"`.
- Jika pet sudah ada, hanya kolom `pet_name` yang diperbarui.

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

**Error — 404 user tidak terdaftar**

Jika `user_id` tidak ada di tabel `users` (foreign key violation):

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

**Alur**

1. Validasi pet sudah ada (harus di-setup terlebih dahulu).
2. Simpan record ke tabel `daily_logs`.
3. Jalankan **Logic Engine** untuk menghitung ulang skor dan menentukan kondisi.
4. Update pet di tabel `pets`.
5. Panggil Gemini AI (`gemini-1.5-flash`) untuk respons empatik dari persona pet berdasarkan kondisi baru dan teks jurnal. Fallback otomatis jika AI error.

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

**Error — 404** (pet belum di-setup)

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

Mengembalikan pet ke `health_score: 50`, `energy_score: 50`, `current_state: "Neutral"`. **Riwayat daily_logs tidak dihapus.**

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

Simulasi fast-forward untuk demo: set pet ke `health_score: 20`, `energy_score: 20`, `current_state: "Sad"`.

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

Mengembalikan 10 log aktivitas terbaru, diurutkan berdasarkan `created_at DESC`.

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

## Aturan Logic Engine

Skor dimulai dari nilai pet saat ini dan diperbarui setiap ada aktivitas:

| Kondisi | Efek |
| --- | --- |
| `water_glasses >= 4` | `health_score + 15` |
| `sleep_hours >= 7.0` | `energy_score + 25` |
| `journal_text` tidak kosong | `health_score + 10`, `energy_score + 10` |

- Skor di-**clamp** antara `0` dan `100`.
- Kondisi ditentukan berdasarkan urutan prioritas berikut:

| Prioritas | Kondisi | State |
| --- | --- | --- |
| 1 | `(health + energy) / 2 >= 75` | `Happy` |
| 2 | `energy < 35` | `Tired` |
| 3 | `health < 40` | `Sad` |
| 4 | selain itu | `Neutral` |

## Penanganan Error

| HTTP Status | Skenario |
| --- | --- |
| `400` | Body JSON tidak valid, `user_id` kosong, atau `pet_name` kosong |
| `404` | Pet tidak ditemukan, atau user tidak terdaftar (FK violation saat setup) |
| `500` | Gagal akses database/pemrosesan (dicatat di log server) |

## Deployment (Render)

1. Push repository ke GitHub.
2. Di Render, buat **New Web Service** dan hubungkan repository.
3. Konfigurasi:

| Setting | Nilai |
| --- | --- |
| Build Command | `go build -o pet-wellness-backend` |
| Start Command | `./pet-wellness-backend` |

4. Tambahkan environment variables:

| Variable | Contoh |
| --- | --- |
| `PORT` | `8080` |
| `DATABASE_URL` | `postgres://user:password@host:5432/dbname?sslmode=require` |
| `GEMINI_API_KEY` | `AIzaSy...` |

> `.env` di-ignore oleh git; semua secret diisi lewat dashboard Render.

## Postman Collection

Import `postman/AI-Wellness-Pet.postman_collection.json` ke Postman. Collection mencakup variabel (`base_url`, `user_id`, `pet_name`) dan contoh response untuk setiap endpoint.

## Testing

Unit tests mencakup Logic Engine (`logic_engine_test.go`) dan deteksi FK violation (`pet_service_test.go`).

```bash
go test ./...
```

## Lisensi

Project ini dibuat untuk demo hackathon. Tidak ada lisensi yang ditetapkan.
