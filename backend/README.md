# SIAKAD Backend Scaffold

Backend ini dibuat sebagai **modular monolith** dengan Go dan MySQL.

## Struktur

```text
backend/
├── cmd/api
│   └── main.go
├── internal
│   ├── app
│   ├── config
│   ├── database
│   ├── httpserver
│   ├── modules
│   │   ├── academic
│   │   ├── auth
│   │   ├── industryrelations
│   │   ├── master
│   │   ├── shared
│   │   └── studentaffairs
│   └── response
├── .env.example
└── go.mod
```

## Modul

- `auth`
- `master`
- `student_affairs`
- `academic`
- `industry_relations`
- `shared`

## Menjalankan

1. Masuk ke folder backend

```bash
cd backend
```

2. Copy env

```bash
cp .env.example .env
```

3. Atur koneksi MySQL di `.env`

Contoh penting:

```env
MYSQL_HOST=192.168.64.5
MYSQL_PORT=3306
MYSQL_USER=root
MYSQL_PASSWORD=your-password
MYSQL_DATABASE=siakad_db
```

4. Download dependency

```bash
go mod tidy
```

5. Jalankan server

```bash
source .env && go run ./cmd/api
```

## Endpoint awal

- `GET /health`
- `GET /api/v1`
- `GET /api/v1/auth/health`
- `GET /api/v1/master/health`
- `GET /api/v1/student-affairs/health`
- `GET /api/v1/academic/health`
- `GET /api/v1/industry-relations/health`
- `GET /api/v1/shared/health`

## Catatan

- Saat ini endpoint masih berupa scaffold.
- Layer business logic dan repository belum diisi implementasi CRUD.
- Schema database utama ada di `../db/schema.sql`.

## Langkah berikutnya

Setelah scaffold ini, langkah ideal berikutnya:

1. buat repository untuk tabel master
2. buat auth login dasar
3. buat CRUD `students`
4. buat CRUD `teachers`
5. buat CRUD `classes`, `departments`, `academic_years`
