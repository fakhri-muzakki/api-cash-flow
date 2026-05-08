# 💰 Cash Flow Manager - Backend API

[![Go](https://img.shields.io/badge/Go-1.23-00ADD8)](https://go.dev/)
[![Gin](https://img.shields.io/badge/Gin-1.10-00ADD8)](https://gin-gonic.com/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1)](https://www.postgresql.org/)
[![JWT](https://img.shields.io/badge/JWT-auth-000000)](https://jwt.io/)

> Backend API untuk aplikasi manajemen keuangan pribadi. Dibangun dengan Go, Gin framework, dan PostgreSQL. Mendukung autentikasi JWT, CRUD transaksi, filter periode, serta pagination.

**Frontend Repository:** [project-4-fe](https://github.com/fakhri-muzakki/project-4-fe)

---

## ✨ Fitur

- **Autentikasi** – Register, login, dan proteksi route dengan JWT
- **CRUD Transaksi** – Tambah, baca, edit, hapus pemasukan/pengeluaran
- **Filter Periode** – Transaksi dapat difilter per hari, minggu, bulan, tahun, atau custom tanggal
- **Pagination** – Dukungan `page` & `limit` untuk data transaksi

---

## 🛠️ Tech Stack

| Kategori       | Teknologi        |
| -------------- | ---------------- |
| Bahasa         | Go 1.23          |
| Web Framework  | Gin              |
| Database       | PostgreSQL 16    |
| Authentication | JWT (golang-jwt) |
| Migration      | golang-migrate   |
| Env Config     | godotenv         |

---

## 📁 Project Structure

```txt
cash-flow/
├── api/
├── cmd/
│ └── main.go # Entry point
├── internal/
│ ├── config/ # Konfigurasi (env, db)
│ ├── handler/ # HTTP handlers (auth, transaction)
│ ├── middleware/ # Auth middleware, logger, CORS
│ ├── model/ # Struct definitions
│ ├── repository/ # Database operations
│ ├── router/ # Route registration
│ ├── service/ # Business logic
├── migrations/ # SQL migration files
├── .env.example
├── go.mod
├── go.sum
└── README.md
└── vercel.json
```

---

## 🚀 Getting Started

### Prerequisites

- Go 1.23+
- PostgreSQL 16+
- Make (optional)

### Installation

```bash
git clone https://github.com/fakhri-muzakki/api-cash-flow.git
cd api-cash-flow
cp .env.example .env   # sesuaikan dengan konfigurasi database
go mod download
```

### Database Migration

```bash
# Install migrate CLI jika belum
migrate -path migrations -database "postgresql://neondb_owner:npg_pve1mUMC4Zla@ep-soft-silence-aq9khg59-pooler.c-8.us-east-1.aws.neon.tech/neondb?sslmode=require&channel_binding=require" up

```

### Run Application

```bash
air
```

Server akan berjalan di http://localhost:8080

### Build Binary

```bash
go build -o cashflow-api ./cmd

```

## 🔧 Environment Variables

| Variable           | Example                                                                 |
| ------------------ | ----------------------------------------------------------------------- |
| `DATABASE_URL`     | `postgres://postgres:postgres@localhost:5432/cash_flow?sslmode=disable` |
| `DB_HOST`          | `localhost`                                                             |
| `DB_PORT`          | `5432`                                                                  |
| `DB_USER`          | `postgres`                                                              |
| `DB_PASSWORD`      | `postgres`                                                              |
| `DB_NAME`          | `cash_flow`                                                             |
| `DB_SSLMODE`       | `disable`                                                               |
| `JWT_SECRET`       | `8b444e63237a164a1a65214861078f9c7e13c1f435ed008b3494c40bf55cfeb8`      |
| `JWT_EXPIRY_HOURS` | `24`                                                                    |
| `APP_PORT`         | `8080`                                                                  |
| `CORS_ORIGIN_URL`  | `http://localhost:3000`                                                 |

---

## 📡 API Endpoints

### 🔐 Authentication

| Method | Endpoint             | Description                 |
| ------ | -------------------- | --------------------------- |
| `POST` | `/api/auth/register` | Register user baru          |
| `POST` | `/api/auth/login`    | Login dan mengembalikan JWT |

Register Request Body:

```json
{
  "name": "Fakhri Muzakki",
  "email": "fakhri@example.com",
  "password": "secret123"
}
```

Login Request Body:

```json
{
  "email": "fakhri@example.com",
  "password": "secret123"
}
```

Login Response:

```json
{
  "token": "eyJhbGci...",
  "user": {
    "id": "uuid",
    "name": "Fakhri Muzakki",
    "email": "fakhri@example.com",
    "balance": 0
  }
}
```

### 💸 Transactions (Protected – Requires Bearer Token)

| Method   | Endpoint                | Description                                              |
| -------- | ----------------------- | -------------------------------------------------------- |
| `GET`    | `/api/transactions`     | Mengambil daftar transaksi (support filter & pagination) |
| `POST`   | `/api/transactions`     | Menambahkan transaksi baru                               |
| `PUT`    | `/api/transactions/:id` | Mengupdate transaksi berdasarkan ID                      |
| `DELETE` | `/api/transactions/:id` | Menghapus transaksi berdasarkan ID                       |

### 📌 GET Query Parameters

| Parameter    | Values                                     | Default                     |
| ------------ | ------------------------------------------ | --------------------------- |
| `page`       | `number`                                   | `1`                         |
| `limit`      | `number`                                   | `10`                        |
| `period`     | `today`, `week`, `month`, `year`, `custom` | `-`                         |
| `date_start` | `YYYY-MM-DD`                               | Required if `period=custom` |
| `date_end`   | `YYYY-MM-DD`                               | Required if `period=custom` |

GET Response:

```json
{
  "data": [...],
  "meta": {
    "page": 1,
    "limit": 10,
    "total_items": 27,
    "total_pages": 3
  }
}
```

POST/PUT Request Body:

```json
{
  "type": "income" | "expense",
  "amount": 5000000,
  "note": "Gaji bulanan",
  "date": "2026-05-01"
}
```

## 👨‍💻 Author

Fakhri Muzakki
