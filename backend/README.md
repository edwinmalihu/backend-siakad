# SIAKAD Backend Scaffold

Backend ini dibuat sebagai **modular monolith** dengan Go dan MySQL.

## Dokumentasi API

Dokumentasi API awal tersedia dalam dua format:

- OpenAPI/Swagger: [docs/openapi.yaml](/Users/edwinmalihu/Documents/internal/SIAKAD/backend/docs/openapi.yaml)
- Postman collection: [siakad-backend.postman_collection.json](/Users/edwinmalihu/Documents/internal/SIAKAD/backend/docs/postman/siakad-backend.postman_collection.json)

Saat backend berjalan, file OpenAPI juga bisa diakses langsung dari:

- `GET /openapi.yaml`
- `GET /docs/openapi.yaml`

Cara pakai:

1. Import [docs/openapi.yaml](/Users/edwinmalihu/Documents/internal/SIAKAD/backend/docs/openapi.yaml) ke Swagger Editor atau tools yang mendukung OpenAPI 3.
2. Atau import [siakad-backend.postman_collection.json](/Users/edwinmalihu/Documents/internal/SIAKAD/backend/docs/postman/siakad-backend.postman_collection.json) ke Postman.
3. Ganti `baseUrl` bila server Anda tidak berjalan di `http://127.0.0.1:18080`.

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
- `GET /api/v1/master/academic-years`
- `POST /api/v1/master/academic-years`
- `GET /api/v1/master/academic-years/{id}`
- `PUT /api/v1/master/academic-years/{id}`
- `DELETE /api/v1/master/academic-years/{id}`
- `GET /api/v1/master/classes`
- `POST /api/v1/master/classes`
- `GET /api/v1/master/classes/{id}`
- `PUT /api/v1/master/classes/{id}`
- `DELETE /api/v1/master/classes/{id}`
- `GET /api/v1/master/departments`
- `POST /api/v1/master/departments`
- `GET /api/v1/master/departments/{id}`
- `PUT /api/v1/master/departments/{id}`
- `DELETE /api/v1/master/departments/{id}`
- `GET /api/v1/master/grade-levels`
- `POST /api/v1/master/grade-levels`
- `GET /api/v1/master/grade-levels/{id}`
- `PUT /api/v1/master/grade-levels/{id}`
- `DELETE /api/v1/master/grade-levels/{id}`
- `GET /api/v1/master/semesters`
- `POST /api/v1/master/semesters`
- `GET /api/v1/master/semesters/{id}`
- `PUT /api/v1/master/semesters/{id}`
- `DELETE /api/v1/master/semesters/{id}`
- `GET /api/v1/student-affairs/health`
- `GET /api/v1/academic/health`
- `GET /api/v1/industry-relations/health`
- `GET /api/v1/shared/health`

## Catatan

- Saat ini endpoint masih berupa scaffold.
- `academic_years` sudah memiliki CRUD dasar end-to-end.
- Layer business logic dan repository belum diisi implementasi CRUD.
- Schema database utama ada di `../db/schema.sql`.

## Contoh Request Academic Years

### Create

```bash
curl -X POST http://localhost:8080/api/v1/master/academic-years \
  -H "Content-Type: application/json" \
  -d '{
    "name": "2026/2027",
    "start_date": "2026-07-01",
    "end_date": "2027-06-30",
    "is_active": true
  }'
```

### List

```bash
curl "http://localhost:8080/api/v1/master/academic-years?search=2026&is_active=true"
```

### Update

```bash
curl -X PUT http://localhost:8080/api/v1/master/academic-years/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "2026/2027",
    "start_date": "2026-07-01",
    "end_date": "2027-06-30",
    "is_active": false
  }'
```

### Delete

```bash
curl -X DELETE http://localhost:8080/api/v1/master/academic-years/1
```

## Contoh Request Departments

### Create

```bash
curl -X POST http://localhost:8080/api/v1/master/departments \
  -H "Content-Type: application/json" \
  -d '{
    "code": "RPL",
    "name": "Rekayasa Perangkat Lunak",
    "program_name": "Teknik Informatika",
    "field_name": "Pengembangan Perangkat Lunak",
    "description": "Jurusan untuk pengembangan software"
  }'
```

### List

```bash
curl "http://localhost:8080/api/v1/master/departments?search=RPL"
```

### Update

```bash
curl -X PUT http://localhost:8080/api/v1/master/departments/1 \
  -H "Content-Type: application/json" \
  -d '{
    "code": "RPL",
    "name": "Rekayasa Perangkat Lunak",
    "program_name": "Teknik Informatika",
    "field_name": "Software Engineering",
    "description": "Jurusan untuk pengembangan software"
  }'
```

### Delete

```bash
curl -X DELETE http://localhost:8080/api/v1/master/departments/1
```

## Contoh Request Grade Levels

### Create

```bash
curl -X POST http://localhost:8080/api/v1/master/grade-levels \
  -H "Content-Type: application/json" \
  -d '{
    "code": "X",
    "name": "Kelas 10",
    "sort_order": 10
  }'
```

### List

```bash
curl "http://localhost:8080/api/v1/master/grade-levels?search=X"
```

### Update

```bash
curl -X PUT http://localhost:8080/api/v1/master/grade-levels/1 \
  -H "Content-Type: application/json" \
  -d '{
    "code": "X",
    "name": "Kelas 10",
    "sort_order": 1
  }'
```

### Delete

```bash
curl -X DELETE http://localhost:8080/api/v1/master/grade-levels/1
```

## Contoh Request Classes

### Create

```bash
curl -X POST http://localhost:8080/api/v1/master/classes \
  -H "Content-Type: application/json" \
  -d '{
    "academic_year_id": 1,
    "department_id": 1,
    "grade_level_id": 1,
    "name": "A",
    "is_active": true
  }'
```

### List

```bash
curl "http://localhost:8080/api/v1/master/classes?academic_year_id=1&department_id=1&grade_level_id=1&is_active=true"
```

### Update

```bash
curl -X PUT http://localhost:8080/api/v1/master/classes/1 \
  -H "Content-Type: application/json" \
  -d '{
    "academic_year_id": 1,
    "department_id": 1,
    "grade_level_id": 1,
    "name": "A1",
    "is_active": false
  }'
```

### Delete

```bash
curl -X DELETE http://localhost:8080/api/v1/master/classes/1
```

## Contoh Request Semesters

### Create

```bash
curl -X POST http://localhost:8080/api/v1/master/semesters \
  -H "Content-Type: application/json" \
  -d '{
    "academic_year_id": 1,
    "name": "Semester Ganjil",
    "code": "GANJIL",
    "is_active": true
  }'
```

### List

```bash
curl "http://localhost:8080/api/v1/master/semesters?search=ganjil&academic_year_id=1&is_active=true"
```

### Update

```bash
curl -X PUT http://localhost:8080/api/v1/master/semesters/1 \
  -H "Content-Type: application/json" \
  -d '{
    "academic_year_id": 1,
    "name": "Semester Ganjil",
    "code": "GANJIL",
    "is_active": false
  }'
```

### Delete

```bash
curl -X DELETE http://localhost:8080/api/v1/master/semesters/1
```
