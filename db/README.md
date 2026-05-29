# Database — SIAKAD

## Struktur Direktori

```
db/
├── schema.sql        -- DDL utama (CREATE TABLE seluruh tabel)
├── seed-rbac.sql     -- Seed data untuk RBAC (roles, permissions, role_permissions)
├── migrations/       -- Folder untuk migration file tambahan
└── README.md         -- Dokumentasi ini
```

---

## Menjalankan Schema

Untuk membuat seluruh tabel dari awal:

```bash
mysql -h <HOST> -P <PORT> -u <USER> -p <DATABASE> < db/schema.sql
```

Contoh:

```bash
mysql -h 192.168.64.5 -P 3306 -u siakad_user -p siakad_db < db/schema.sql
```

---

## Menjalankan Seed RBAC

File `db/seed-rbac.sql` berisi data awal untuk sistem **Role-Based Access Control (RBAC)**:

- **6 roles** — admin, academic, student_affairs, industry_relations, hubim, shared
- **11 permissions** — hak akses granular per modul (read/write)
- **Role-Permission assignments** — mapping role ke permission-nya masing-masing

### Cara Menjalankan

```bash
mysql -h <HOST> -P <PORT> -u <USER> -p <DATABASE> < db/seed-rbac.sql
```

Contoh:

```bash
mysql -h 192.168.64.5 -P 3306 -u siakad_user -p siakad_db < db/seed-rbac.sql
```

### Catatan Penting

- Seed menggunakan `ON DUPLICATE KEY UPDATE` sehingga **aman dijalankan berulang kali** tanpa error duplikat.
- Jalankan **setelah** `schema.sql` karena seed bergantung pada tabel `roles`, `permissions`, dan `role_permissions` yang sudah ada.
- Seed **tidak menghapus** data yang sudah ada — hanya menambah data yang belum ada.

---

## Daftar Roles

| ID | Name | Code | Description |
|----|------|------|-------------|
| 1 | Administrator | `admin` | Full system access with all permissions |
| 2 | Academic | `academic` | Access to academic module |
| 3 | Student Affairs | `student_affairs` | Access to student affairs module |
| 4 | Industry Relations | `industry_relations` | Access to industry relations module |
| 5 | HUBIM | `hubim` | Access to HUBIM module |
| 6 | Shared | `shared` | Access to shared features |

---

## Daftar Permissions

| ID | Name | Code | Description |
|----|------|------|-------------|
| 1 | Master Read | `master.read` | View master data |
| 2 | Master Write | `master.write` | Create, update, delete master data |
| 3 | Academic Read | `academic.read` | View academic data |
| 4 | Academic Write | `academic.write` | Create, update, delete academic data |
| 5 | Student Affairs Read | `student_affairs.read` | View student affairs data |
| 6 | Student Affairs Write | `student_affairs.write` | Create, update, delete student affairs data |
| 7 | Industry Relations Read | `industry_relations.read` | View industry relations data |
| 8 | Industry Relations Write | `industry_relations.write` | Create, update, delete industry relations data |
| 9 | Shared Read | `shared.read` | View shared features |
| 10 | Shared Write | `shared.write` | Create, update, delete shared features |
| 11 | User Management | `user_management.full` | Manage roles, permissions, assignments |

---

## Mapping Role → Permissions

| Role | Permissions |
|------|-------------|
| **Administrator** | ALL (admin.read, master.read/write, academic.read/write, student_affairs.read/write, industry_relations.read/write, shared.read/write, user_management.full) |
| **Academic** | master.read, master.write, academic.read, academic.write |
| **Student Affairs** | student_affairs.read, student_affairs.write |
| **Industry Relations** | industry_relations.read, industry_relations.write |
| **HUBIM** | industry_relations.read, industry_relations.write |
| **Shared** | shared.read, shared.write |

---

## API Endpoints RBAC

| Method | Path | Deskripsi |
|--------|------|-----------|
| `GET` | `/api/v1/roles` | List semua roles |
| `POST` | `/api/v1/roles` | Buat role baru |
| `GET` | `/api/v1/roles/{id}` | Detail role |
| `PUT` | `/api/v1/roles/{id}` | Update role |
| `DELETE` | `/api/v1/roles/{id}` | Hapus role (soft delete) |
| `GET` | `/api/v1/permissions` | List semua permissions |
| `POST` | `/api/v1/permissions` | Buat permission baru |
| `GET` | `/api/v1/permissions/{id}` | Detail permission |
| `PUT` | `/api/v1/permissions/{id}` | Update permission |
| `DELETE` | `/api/v1/permissions/{id}` | Hapus permission (soft delete) |
| `GET` | `/api/v1/roles/{id}/permissions` | List permissions untuk role |
| `PUT` | `/api/v1/roles/{id}/permissions` | Replace semua permissions role |
| `GET` | `/api/v1/users` | List users |
| `GET` | `/api/v1/users/{id}/roles` | List roles untuk user |
| `PUT` | `/api/v1/users/{id}/roles` | Replace semua roles user |

---

## Frontend Pages RBAC

| Path | Halaman | Deskripsi |
|------|---------|-----------|
| `/admin/roles` | Roles | CRUD roles (name, code, description) |
| `/admin/permissions` | Permissions | CRUD permissions (name, code, description) |
| `/admin/role-permissions` | Role Permissions | Assign permissions ke role |
| `/admin/user-roles` | User Roles | Assign roles ke user |

Semua halaman RBAC hanya bisa diakses oleh user dengan role `admin`.

---

## Auto-Init di Docker Container

MySQL container menggunakan `entrypoint.sh` yang otomatis melakukan inisialisasi:

### Urutan Inisialisasi

1. **Schema** — Menjalankan `01-schema.sql` (dari `schema.sql`)
2. **Migrations** — Menjalankan file di `migrations/*.sql` (idempotent, tracking via `_migrations` table)
3. **Seed RBAC** — Menjalankan `03-seed-rbac.sql`
4. **Admin User** — Membuat user `admin` (password: `Admin123!`)
5. **Admin Role** — Assign role administrator

### Behavior

- **Pertama kali** (volume kosong): Full inisialisasi (schema + migrations + seed + admin)
- **Restart** (sudah ada data): Hanya menjalankan migrations baru yang belum di-apply

### Migrations

File migration di `migrations/` dijalankan berdasarkan nama file (sorted). Progress tracking menggunakan tabel `_migrations`.

```
_migrations/
├── id              — Auto increment
├── filename        — Nama file (unique)
└── applied_at      — Timestamp penerapan
```

### Menambah Migration Baru

Buat file baru di `db/migrations/` dengan format `YYYYMMDD_nama.sql`:

```sql
-- Contoh: 20260530_add_new_column.sql
ALTER TABLE `some_table` ADD COLUMN `new_column` VARCHAR(100) DEFAULT NULL;
```

**PENTING**: Semua migration harus **idempotent** (aman dijalankan berulang kali). Gunakan `IF NOT EXISTS` / `IF EXISTS` / `IF @col_exists = 0` pattern.

### Contoh Migration Idempotent

```sql
SET @col_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'my_table' AND COLUMN_NAME = 'new_col');
SET @sql = IF(@col_exists = 0,
  'ALTER TABLE `my_table` ADD COLUMN `new_col` VARCHAR(100) DEFAULT NULL',
  'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
```
