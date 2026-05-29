# SIAKAD Backend

Backend API untuk sistem informasi akademik (SIAKAD) dibangun dengan Go 1.23 dan MySQL 8.0.

## Arsitektur

Modular monolith dengan pattern `Module` interface per domain:

```
cmd/api/main.go          → Entry point, signal handling, graceful shutdown
internal/
├── app/app.go           → Bootstrap: config, DB, modules, server
├── config/config.go     → .env loader, env var parsing
├── database/mysql.go    → MySQL connection, health check
├── httpserver/server.go  → HTTP server, auth middleware, base routes
├── response/json.go     → JSON/error response helpers
└── modules/
    ├── auth/             → Login, logout, token (HMAC-SHA256), middleware
    ├── master/           → Academic years, semesters, departments, grade levels, classes, rooms
    ├── academic/         → Teachers, subjects, schedules, assessments, grades
    ├── studentaffairs/   → Students, enrollments, mutations, attendances, discipline, extracurriculars
    ├── industryrelations/→ Companies, internships, alumni, industry categories
    ├── shared/           → Announcements, audit logs, student search, import/export
    └── usermanagement/   → Roles, permissions, users CRUD, role-permission assignments
```

## Dependensi

| Package | Version | Fungsi |
|---------|---------|--------|
| `go-sql-driver/mysql` | v1.8.1 | MySQL driver |
| `xuri/excelize/v2` | v2.9.0 | Excel import/export |
| `golang.org/x/crypto` | v0.31.0 | bcrypt password hashing |

## Konfigurasi

Buat file `.env` di direktori `backend/`:

```env
APP_NAME=SIAKAD Backend
APP_ENV=development
APP_HOST=0.0.0.0
APP_PORT=18080
AUTH_TOKEN_SECRET=your-secret-key
AUTH_TOKEN_TTL=24h
MYSQL_ENABLED=true
MYSQL_HOST=127.0.0.1
MYSQL_PORT=3306
MYSQL_USER=siakad_user
MYSQL_PASSWORD=your-password
MYSQL_DATABASE=siakad_db
```

Atau set environment variable langsung.

## Menjalankan

### Local Development

```bash
cd backend
cp .env.example .env   # Edit sesuai kebutuhan
go run ./cmd/api
```

Server berjalan di `http://localhost:18080`.

### Build Binary

```bash
cd backend
go build -o siakad-backend ./cmd/api
./siakad-backend
```

### Docker

```bash
cd backend
docker build -t siakad-backend .
docker run -p 18080:18080 \
  -e MYSQL_HOST=host.docker.internal \
  -e MYSQL_USER=siakad_user \
  -e MYSQL_PASSWORD=your-password \
  -e MYSQL_DATABASE=siakad_db \
  siakad-backend
```

## API Endpoints

### Authentication
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/login` | Login |
| POST | `/api/v1/auth/logout` | Logout |
| GET | `/api/v1/auth/me` | Get current user |

### Master Data
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET/POST | `/api/v1/master/academic-years` | Tahun akademik |
| GET/POST | `/api/v1/master/semesters` | Semester |
| GET/POST | `/api/v1/master/departments` | Jurusan |
| GET/POST | `/api/v1/master/grade-levels` | Tingkat |
| GET/POST | `/api/v1/master/classes` | Kelas |
| GET/POST | `/api/v1/master/rooms` | Ruangan |

### Academic
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET/POST | `/api/v1/academic/teachers` | Guru |
| GET/POST | `/api/v1/academic/subjects` | Mata pelajaran |
| GET/POST | `/api/v1/academic/schedules` | Jadwal |
| GET/POST | `/api/v1/academic/homeroom-assignments` | Wali kelas |
| GET/POST | `/api/v1/academic/assessment-components` | Komponen penilaian |
| GET/POST | `/api/v1/academic/student-assessments` | Penilaian siswa |
| GET/POST | `/api/v1/academic/student-grades` | Nilai siswa |

### Student Affairs
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET/POST | `/api/v1/student-affairs/students` | Siswa |
| GET/POST | `/api/v1/student-affairs/enrollments` | Enrollment |
| GET/POST | `/api/v1/student-affairs/mutations` | Mutasi |
| GET/POST | `/api/v1/student-affairs/graduations` | Kelulusan |
| GET/POST | `/api/v1/student-affairs/attendances` | Kehadiran |
| GET/POST | `/api/v1/student-affairs/discipline-categories` | Kategori disiplin |
| GET/POST | `/api/v1/student-affairs/discipline-records` | Catatan disiplin |
| GET/POST | `/api/v1/student-affairs/extracurriculars` | Ekstrakurikuler |
| GET/POST | `/api/v1/student-affairs/extracurricular-members` | Keanggotaan ekstrakurikuler |

### Industry Relations
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET/POST | `/api/v1/industry-relations/categories` | Kategori industri |
| GET/POST | `/api/v1/industry-relations/companies` | Perusahaan |
| GET/POST | `/api/v1/industry-relations/internships` | Magang |
| GET/POST | `/api/v1/industry-relations/internship-logs` | Log magang |
| GET/POST | `/api/v1/industry-relations/alumni` | Alumni |

### Shared
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET/POST | `/api/v1/shared/announcements` | Pengumuman |
| GET | `/api/v1/shared/audit-logs` | Log audit |
| GET | `/api/v1/shared/student-search` | Pencarian siswa |
| POST | `/api/v1/shared/import/{module}` | Import Excel |
| GET | `/api/v1/shared/import/{module}/template` | Download template |
| GET | `/api/v1/shared/export/{module}` | Export Excel |

### User Management
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET/POST | `/api/v1/roles` | Role |
| GET/POST | `/api/v1/permissions` | Permission |
| GET/POST | `/api/v1/users` | User |
| PUT | `/api/v1/roles/{id}/permissions` | Assign permission ke role |
| PUT | `/api/v1/users/{id}/roles` | Assign role ke user |

## Struktur Module

Setiap module mengikuti pattern:

```
module/
├── types.go        → Struct entity, create/update request
├── repository.go   → SQL queries, CRUD, validasi referensi, soft delete
├── handler.go      → HTTP handlers, JSON decode/encode, validasi, audit log
└── module.go       → Wiring handler, registrasi routes
```

## Fitur

- **Soft Delete** — Semua entity pakai `deleted_at` + VIRTUAL columns untuk unique constraint
- **RBAC** — 6 role, 11 permission, role-permission dan user-role junction tables
- **Audit Logging** — Setiap create/update/delete tercatat di `audit_logs`
- **Import/Export Excel** — Support import/export untuk students, teachers, departments, grade levels, academic years
- **Auth** — Custom HMAC-SHA256 token dengan role-based access

## Default Credentials

| Username | Password | Role |
|----------|----------|------|
| admin | Admin123! | Administrator |
