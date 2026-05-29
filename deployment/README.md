# SIAKAD Deployment

Panduan deployment SIAKAD menggunakan Docker/Podman Compose (local) dan Fly.io (production).

## Arsitektur Deployment

```
┌─────────────────────────────────────────────┐
│              SIAKAD Stack                   │
│                                             │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
│  │ Frontend │  │ Backend  │  │  MySQL   │  │
│  │  (Nginx) │  │   (Go)   │  │  8.0     │  │
│  │  :80     │  │  :18080  │  │  :3306   │  │
│  └──────────┘  └──────────┘  └──────────┘  │
└─────────────────────────────────────────────┘
```

## Docker Images

| Image | Tag | Registry | Size |
|-------|-----|----------|------|
| `edwinmalihu/backend-siakad` | `0.0.1-amd64` | Docker Hub | ~25 MB |
| `edwinmalihu/frontend-siakad` | `0.0.1-amd64` | Docker Hub | ~55 MB |
| `edwinmalihu/mysql-siakad` | `0.0.1-amd64` | Docker Hub | ~814 MB |

## Docker Hub

```bash
# Backend
podman pull edwinmalihu/backend-siakad:0.0.1-amd64

# Frontend
podman pull edwinmalihu/frontend-siakad:0.0.1-amd64

# MySQL (dengan auto-init)
podman pull edwinmalihu/mysql-siakad:0.0.1-amd64
```

## MySQL Auto-Init

MySQL container otomatis melakukan inisialisasi saat pertama kali start:

1. **Schema** — Menjalankan `schema.sql` membuat semua tabel
2. **Migrations** — Menjalankan file SQL di `db/migrations/` (idempotent)
3. **Seed RBAC** — Menjalankan `seed-rbac.sql` (roles, permissions)
4. **Admin User** — Membuat user `admin` (password: `Admin123!`)
5. **Admin Role** — Assign role administrator ke user admin

Jika database sudah ada (volume persist), hanya menjalankan migrations baru.

---

## Local Development (Docker/Podman Compose)

### Prasyarat

- Docker atau Podman terinstall
- Port 3306, 5173, 18080 available

### Cara Jalankan

```bash
# Clone repository
git clone https://github.com/edwinmalihu/SIAKAD.git
cd SIAKAD

# Jalankan dengan pre-built images
cd deployment
podman compose up -d

# Atau build sendiri (dari root directory)
podman compose -f deployment/docker-compose.yml up --build -d
```

### Akses

| Service | URL |
|---------|-----|
| Frontend | http://localhost:5173 |
| Backend API | http://localhost:18080 |
| MySQL | localhost:3306 |

### Login

| Username | Password | Role |
|----------|----------|------|
| admin | Admin123! | Administrator |

### Perintah Useful

```bash
# Lihat status
podman compose ps

# Lihat logs
podman compose logs -f backend
podman compose logs -f mysql

# Stop semua service
podman compose down

# Stop dan hapus volume (fresh start)
podman compose down -v

# Rebuild image
podman compose up --build -d
```

---

## Production Deployment (Fly.io)

### Prasyarat

- [flyctl](https://fly.io/docs/hands-on/install-flyctl/) terinstall
- Akun Fly.io
- Semua image sudah di-push ke Docker Hub

### Struktur Deployment

```
deployment/
├── docker-compose.yml      ← Local testing
├── fly-mysql.toml          ← Fly.io MySQL config
├── fly-backend.toml        ← Fly.io Backend config
├── fly-frontend.toml       ← Fly.io Frontend config
└── mysql-init/
    └── entrypoint.sh       ← MySQL auto-init script
```

### Langkah Deployment

#### 1. Login ke Fly.io

```bash
fly auth login
```

#### 2. Deploy MySQL

```bash
# Create app
fly apps create siakad-mysql

# Create volume untuk persist data
fly volumes create mysql_data --region sin --size 1 -a siakad-mysql --yes

# Deploy
fly deploy --config deployment/fly-mysql.toml -a siakad-mysql
```

#### 3. Deploy Backend

```bash
# Create app
fly apps create siakad-backend

# Deploy
fly deploy --config deployment/fly-backend.toml -a siakad-backend
```

#### 4. Deploy Frontend

```bash
# Create app
fly apps create siakad-frontend

# Deploy
fly deploy --config deployment/fly-frontend.toml -a siakad-frontend
```

### URL Production

| Service | URL |
|---------|-----|
| Frontend | https://siakad-frontend.fly.dev |
| Backend API | https://siakad-backend.fly.dev |
| MySQL | siakad-mysql.internal (internal only) |

### Fly.io Commands

```bash
# Cek status
fly status -a siakad-backend

# Lihat logs
fly logs -a siakad-backend

# SSH ke container
fly ssh console -a siakad-mysql

# Restart machine
fly machine restart <machine-id> -a siakad-backend

# Scale machine
fly machine scale count 1 -a siakad-backend
```

### Konfigurasi VM

| Service | Memory | CPU | Auto-stop |
|---------|--------|-----|-----------|
| MySQL | 1 GB | shared-1x | No |
| Backend | 512 MB | shared-1x | No |
| Frontend | 256 MB | shared-1x | Yes |

### Environment Variables

Backend (di `fly-backend.toml`):

| Variable | Value |
|----------|-------|
| `APP_PORT` | 18080 |
| `MYSQL_HOST` | siakad-mysql.internal |
| `MYSQL_USER` | siakad_user |
| `MYSQL_DATABASE` | siakad_db |
| `AUTH_TOKEN_SECRET` | (production secret) |

---

## Build Image Sendiri

### Backend

```bash
cd backend
# Build untuk amd64 (Fly.io)
podman build --platform linux/amd64 -t edwinmalihu/backend-siakad:0.0.1-amd64 .

# Push ke Docker Hub
podman push edwinmalihu/backend-siakad:0.0.1-amd64 docker.io/edwinmalihu/backend-siakad:0.0.1-amd64
```

### Frontend

```bash
cd frontend
# Build untuk amd64
podman build --platform linux/amd64 -t edwinmalihu/frontend-siakad:0.0.1-amd64 .

# Push ke Docker Hub
podman push edwinmalihu/frontend-siakad:0.0.1-amd64 docker.io/edwinmalihu/frontend-siakad:0.0.1-amd64
```

### MySQL

```bash
# Build dari root directory
podman build --platform linux/amd64 -t edwinmalihu/mysql-siakad:0.0.1-amd64 -f Dockerfile.mysql .

# Push ke Docker Hub
podman push edwinmalihu/mysql-siakad:0.0.1-amd64 docker.io/edwinmalihu/mysql-siakad:0.0.1-amd64
```

---

## Troubleshooting

### MySQL tidak bisa connect

```bash
# Cek MySQL running
fly ssh console -a siakad-mysql -C "mysqladmin ping -u root -prootpass"

# Cek tabel
fly ssh console -a siakad-mysql -C "mysql -u root -prootpass -e 'USE siakad_db; SHOW TABLES;'"
```

### Backend crash

```bash
# Lihat logs
fly logs -a siakad-backend

# Cek koneksi MySQL
fly ssh console -a siakad-backend -C "nc -zv siakad-mysql.internal 3306"
```

### Frontend 502 Bad Gateway

```bash
# Cek backend running
fly status -a siakad-backend

# Test backend langsung
curl https://siakad-backend.fly.dev/health
```
