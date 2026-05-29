# SIAKAD — Catatan Progress & Rencana Selanjutnya

## Status Terakhir: 30 Mei 2026

---

## 1. Sistem Audit Log ✅ SELESAI

### Yang sudah diimplementasi:
- **Auth Module** — Login (`POST /api/v1/auth/login`), Logout (`POST /api/v1/auth/logout`)
- **Audit Logging** — Setiap CRUD operation tercatat di `audit_logs` table
- **Login/Logout** — 1 record per sesi dengan `login_time` dan `logout_time`
- **Frontend** — Audit Logs page dengan filter, auto-refresh, online status indicator
- **User ID tracking** — User ID tercatat dari auth context

### Modul yang sudah ada audit log:
- Auth (login, login_failed, logout)
- Master (academic_years, semesters, departments, grade_levels, classes, rooms)
- Academic (teachers, subjects, schedules, homeroom_assignments, assessment_components, student_assessments, student_grades)
- Student Affairs (students, enrollments, mutations, graduations, attendances, discipline_categories, discipline_records, extracurriculars, extracurricular_members)
- Industry Relations (companies, alumni, internships, internship_logs, industry_categories)
- User Management (roles, permissions, users, user_roles, role_permissions)
- Shared (announcements, import/export)

---

## 2. Deployment ✅ SELESAI

### Docker Images (Docker Hub):
- `edwinmalihu/backend-siakad:0.0.1-amd64`
- `edwinmalihu/frontend-siakad:0.0.1-amd64`
- `edwinmalihu/mysql-siakad:0.0.1-amd64`

### Fly.io Deployment:
- **Backend:** https://siakad-backend.fly.dev
- **Frontend:** https://siakad-frontend.fly.dev
- **MySQL:** siakad-mysql.internal (internal network)

### MySQL Auto-Init:
- `deployment/mysql-init/entrypoint.sh` — Otomatis jalankan schema, migrations, seed RBAC, buat admin user
- Admin default: `admin` / `Admin123!`

### File Structure:
```
deployment/
├── docker-compose.yml
├── fly-mysql.toml
├── fly-backend.toml
├── fly-frontend.toml
├── README.md
└── mysql-init/entrypoint.sh
```

---

## 3. License Generator — RENCANA (Belum Diimplementasi)

### Status: Rencana technis sudah dibuat di repo `license-generator`

### Repo: https://github.com/edwinmalihu/license-generator.git
- File: `TECHNICAL_PLAN.md` dan `README.md`

### Konsep:
- **Dua jenis license:** Trial (7 hari, max 2x) dan Enterprise (1 tahun)
- **1 sistem terpisah** — License Generator + Dashboard Monitoring
- **Tidak perlu webhook/polling** — SIAKAD call API saat activate/validate
- **Device fingerprint** — Kombinasi hostname + MAC address

### Database (License Generator):
- `admins` — Admin login
- `license_keys` — Semua key (trial + enterprise)
- `license_history` — History activations

### Database (SIAKAD Client):
- `installed_licenses` — License lokal di SIAKAD

### API Endpoints:
- `POST /api/validate` — SIAKAD cek key
- `POST /api/activate` — SIAKAD aktifkan key
- `POST /api/trial` — Mulai trial (max 2x)
- `GET /api/status` — Cek status key
- `POST /api/admin/login` — Admin login
- `GET /api/admin/licenses` — List licenses
- `POST /api/admin/licenses/generate` — Generate key baru
- `GET /api/admin/history` — History activations
- `GET /api/admin/dashboard` — Stats overview

### Implementation Phases:
| Phase | Durasi | Isi |
|-------|--------|-----|
| Phase 1 | 3-4 hari | Backend core (auth, generator, validator, monitor) |
| Phase 2 | 2-3 hari | Frontend dashboard |
| Phase 3 | 2 hari | Integration + deployment |
| Phase 4 | 1-2 hari | SIAKAD client changes |
| **Total** | **8-11 hari** |

### SIAKAD Client Flow:
1. Install → auto trial 7 hari
2. Login → cek license lokal → expired? → redirect /license
3. H-30 → alert warning di semua module
4. Expired → blokir semua halaman, hanya lihat /license
5. Activate key → call API Generator → valid → simpan lokal

---

## 4. Akses & Credentials

### SIAKAD:
| Service | URL | Credentials |
|---------|-----|-------------|
| Frontend | https://siakad-frontend.fly.dev | admin / Admin123! |
| Backend | https://siakad-backend.fly.dev | - |
| MySQL | siakad-mysql.internal | siakad_user / siakadpass |

### License Generator (Belum deploy):
| Service | URL | Credentials |
|---------|-----|-------------|
| Frontend | - | admin / admin123 |
| Backend | - | - |
| MySQL | - | - |

---

## 5. Perintah Penting

### Deploy SIAKAD:
```bash
cd /Users/edwinmalihu/Documents/internal/SIAKAD
# Local
cd deployment && podman compose up -d

# Fly.io
fly deploy --config deployment/fly-mysql.toml -a siakad-mysql
fly deploy --config deployment/fly-backend.toml -a siakad-backend
fly deploy --config deployment/fly-frontend.toml -a siakad-frontend
```

### Build & Push Images:
```bash
# Backend
podman build --platform linux/amd64 -t edwinmalihu/backend-siakad:0.0.1-amd64 -f backend/Dockerfile backend/
podman push edwinmalihu/backend-siakad:0.0.1-amd64 docker.io/edwinmalihu/backend-siakad:0.0.1-amd64

# Frontend
podman build --platform linux/amd64 -t edwinmalihu/frontend-siakad:0.0.1-amd64 frontend/
podman push edwinmalihu/frontend-siakad:0.0.1-amd64 docker.io/edwinmalihu/frontend-siakad:0.0.1-amd64

# MySQL
podman build --platform linux/amd64 -t edwinmalihu/mysql-siakad:0.0.1-amd64 -f Dockerfile.mysql .
podman push edwinmalihu/mysql-siakad:0.0.1-amd64 docker.io/edwinmalihu/mysql-siakad:0.0.1-amd64
```

---

## 6. Yang Perlu Dilanjutkan

1. **Implement License Generator** — Mulai Phase 1 (backend core)
2. **Implement SIAKAD Client License** — Tambah license module ke SIAKAD
3. **Deploy License Generator** — Docker + Fly.io
4. **Testing integrasi** — SIAKAD ↔ License Generator
