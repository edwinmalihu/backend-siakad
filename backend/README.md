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
AUTH_TOKEN_SECRET=dev-secret-change-me
AUTH_TOKEN_TTL=24h
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
- `POST /api/v1/auth/login`
- `GET /api/v1/auth/me`
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
- `GET /api/v1/master/rooms`
- `POST /api/v1/master/rooms`
- `GET /api/v1/master/rooms/{id}`
- `PUT /api/v1/master/rooms/{id}`
- `DELETE /api/v1/master/rooms/{id}`
- `GET /api/v1/master/semesters`
- `POST /api/v1/master/semesters`
- `GET /api/v1/master/semesters/{id}`
- `PUT /api/v1/master/semesters/{id}`
- `DELETE /api/v1/master/semesters/{id}`
- `GET /api/v1/student-affairs/health`
- `GET /api/v1/student-affairs/students`
- `POST /api/v1/student-affairs/students`
- `GET /api/v1/student-affairs/students/{id}`
- `PUT /api/v1/student-affairs/students/{id}`
- `DELETE /api/v1/student-affairs/students/{id}`
- `GET /api/v1/academic/health`
- `GET /api/v1/academic/homeroom-assignments`
- `POST /api/v1/academic/homeroom-assignments`
- `GET /api/v1/academic/homeroom-assignments/{id}`
- `PUT /api/v1/academic/homeroom-assignments/{id}`
- `DELETE /api/v1/academic/homeroom-assignments/{id}`
- `GET /api/v1/academic/teachers`
- `POST /api/v1/academic/teachers`
- `GET /api/v1/academic/teachers/{id}`
- `PUT /api/v1/academic/teachers/{id}`
- `DELETE /api/v1/academic/teachers/{id}`
- `GET /api/v1/academic/schedules`
- `POST /api/v1/academic/schedules`
- `GET /api/v1/academic/schedules/{id}`
- `PUT /api/v1/academic/schedules/{id}`
- `DELETE /api/v1/academic/schedules/{id}`
- `GET /api/v1/academic/subjects`
- `POST /api/v1/academic/subjects`
- `GET /api/v1/academic/subjects/{id}`
- `PUT /api/v1/academic/subjects/{id}`
- `DELETE /api/v1/academic/subjects/{id}`
- `GET /api/v1/industry-relations/health`
- `GET /api/v1/shared/health`

## Catatan

- Modul `master` dan sebagian modul `academic` sudah memiliki CRUD dasar end-to-end.
- Modul `auth` sudah menyediakan login admin berbasis `Bearer token`.
- Endpoint `schedules` sudah memiliki validasi relasi serta deteksi bentrok jadwal untuk kelas, guru, dan ruang.
- Schema database utama ada di `../db/schema.sql`.

## Contoh Request Auth

### Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "identifier": "admin",
    "password": "Admin123!"
  }'
```

### Me

```bash
curl http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer your-access-token"
```

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

## Contoh Request Rooms

### Create

```bash
curl -X POST http://localhost:8080/api/v1/master/rooms \
  -H "Content-Type: application/json" \
  -d '{
    "code": "LAB-01",
    "name": "Laboratorium 1",
    "type": "lab",
    "capacity": 36
  }'
```

### List

```bash
curl "http://localhost:8080/api/v1/master/rooms?search=LAB"
```

### Update

```bash
curl -X PUT http://localhost:8080/api/v1/master/rooms/1 \
  -H "Content-Type: application/json" \
  -d '{
    "code": "LAB-01",
    "name": "Laboratorium 1",
    "type": "computer_lab",
    "capacity": 40
  }'
```

### Delete

```bash
curl -X DELETE http://localhost:8080/api/v1/master/rooms/1
```

## Contoh Request Subjects

### Create

```bash
curl -X POST http://localhost:8080/api/v1/academic/subjects \
  -H "Content-Type: application/json" \
  -d '{
    "department_id": 1,
    "grade_level_id": 1,
    "code": "MTK-101",
    "name": "Matematika Dasar",
    "subject_type": "general",
    "kkm": 75
  }'
```

### List

```bash
curl "http://localhost:8080/api/v1/academic/subjects?search=MTK&department_id=1&grade_level_id=1"
```

### Update

```bash
curl -X PUT http://localhost:8080/api/v1/academic/subjects/1 \
  -H "Content-Type: application/json" \
  -d '{
    "department_id": 1,
    "grade_level_id": 1,
    "code": "MTK-101",
    "name": "Matematika Dasar",
    "subject_type": "compulsory",
    "kkm": 80
  }'
```

### Delete

```bash
curl -X DELETE http://localhost:8080/api/v1/academic/subjects/1
```

## Contoh Request Teachers

### Create

```bash
curl -X POST http://localhost:8080/api/v1/academic/teachers \
  -H "Content-Type: application/json" \
  -d '{
    "nip": "198901012015011001",
    "nuptk": "1234567890123456",
    "full_name": "Budi Santoso",
    "gender": "male",
    "address": "Jl. Pendidikan No. 1",
    "phone": "081234567890",
    "email": "budi.santoso@example.com",
    "employment_status": "permanent",
    "position": "Guru Matematika",
    "photo_url": "https://example.com/photo.jpg",
    "status": "active"
  }'
```

### List

```bash
curl "http://localhost:8080/api/v1/academic/teachers?search=budi&gender=male&status=active"
```

### Update

```bash
curl -X PUT http://localhost:8080/api/v1/academic/teachers/1 \
  -H "Content-Type: application/json" \
  -d '{
    "nip": "198901012015011001",
    "nuptk": "1234567890123456",
    "full_name": "Budi Santoso",
    "gender": "male",
    "address": "Jl. Pendidikan No. 2",
    "phone": "081234567890",
    "email": "budi.santoso@example.com",
    "employment_status": "permanent",
    "position": "Wali Kelas",
    "photo_url": "https://example.com/photo.jpg",
    "status": "active"
  }'
```

### Delete

```bash
curl -X DELETE http://localhost:8080/api/v1/academic/teachers/1
```

## Contoh Request Homeroom Assignments

### Create

```bash
curl -X POST http://localhost:8080/api/v1/academic/homeroom-assignments \
  -H "Content-Type: application/json" \
  -d '{
    "teacher_id": 1,
    "class_id": 1,
    "academic_year_id": 1,
    "semester_id": 1
  }'
```

### List

```bash
curl "http://localhost:8080/api/v1/academic/homeroom-assignments?academic_year_id=1&semester_id=1"
```

### Update

```bash
curl -X PUT http://localhost:8080/api/v1/academic/homeroom-assignments/1 \
  -H "Content-Type: application/json" \
  -d '{
    "teacher_id": 2,
    "class_id": 1,
    "academic_year_id": 1,
    "semester_id": 1
  }'
```

### Delete

```bash
curl -X DELETE http://localhost:8080/api/v1/academic/homeroom-assignments/1
```

## Contoh Request Schedules

### Create

```bash
curl -X POST http://localhost:8080/api/v1/academic/schedules \
  -H "Content-Type: application/json" \
  -d '{
    "class_id": 1,
    "subject_id": 1,
    "teacher_id": 1,
    "room_id": 1,
    "academic_year_id": 1,
    "semester_id": 1,
    "day_of_week": 1,
    "start_time": "07:00",
    "end_time": "08:30"
  }'
```

### List

```bash
curl "http://localhost:8080/api/v1/academic/schedules?academic_year_id=1&semester_id=1&day_of_week=1"
```

### Update

```bash
curl -X PUT http://localhost:8080/api/v1/academic/schedules/1 \
  -H "Content-Type: application/json" \
  -d '{
    "class_id": 1,
    "subject_id": 1,
    "teacher_id": 1,
    "room_id": 1,
    "academic_year_id": 1,
    "semester_id": 1,
    "day_of_week": 1,
    "start_time": "08:00",
    "end_time": "09:30"
  }'
```

### Delete

```bash
curl -X DELETE http://localhost:8080/api/v1/academic/schedules/1
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

## Contoh Request Students

### Create

```bash
curl -X POST http://localhost:8080/api/v1/student-affairs/students \
  -H "Content-Type: application/json" \
  -d '{
    "nis": "2026001",
    "nisn": "9988776655",
    "full_name": "Andi Saputra",
    "gender": "male",
    "birth_place": "Padang",
    "birth_date": "2010-01-12",
    "address": "Jalan Merdeka",
    "phone": "081234567890",
    "entry_year": 2026,
    "status": "active"
  }'
```

### List

```bash
curl "http://localhost:8080/api/v1/student-affairs/students?search=Andi&gender=male&status=active&entry_year=2026"
```

### Update

```bash
curl -X PUT http://localhost:8080/api/v1/student-affairs/students/1 \
  -H "Content-Type: application/json" \
  -d '{
    "nis": "2026001",
    "nisn": "9988776655",
    "full_name": "Andi Saputra",
    "gender": "male",
    "birth_place": "Padang",
    "birth_date": "2010-01-12",
    "address": "Jalan Sudirman",
    "phone": "081234567891",
    "entry_year": 2026,
    "status": "active"
  }'
```

### Delete

```bash
curl -X DELETE http://localhost:8080/api/v1/student-affairs/students/1
```
