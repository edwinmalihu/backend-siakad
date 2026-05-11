# MySQL Installation Notes on VM UTM

## Server

- Host: `192.168.64.5`
- SSH user: `devops`
- OS family: `Rocky Linux 9 / EL9`
- Installed by: `dnf`
- MySQL version: `8.0.45`

## Current Status

- Package `mysql-server` sudah ter-install.
- Service `mysqld` sudah `active` dan `enabled`.
- Root password sudah diamankan.
- MySQL saat ini listen di `*:3306`.
- Firewall saat ini **belum** membuka port `3306` untuk akses masuk.
- File penyimpanan password root ada di VM:
  - `/root/mysql-root-password.txt`

Untuk melihat password root dari user `devops`:

```bash
ssh devops@192.168.64.5
sudo cat /root/mysql-root-password.txt
```

## How To Install

### 1. Login ke VM

```bash
ssh devops@192.168.64.5
```

### 2. Install MySQL Server

```bash
sudo dnf install -y mysql-server
```

### 3. Enable dan start service

```bash
sudo systemctl enable --now mysqld
```

### 4. Cek status service

```bash
sudo systemctl status mysqld --no-pager
sudo systemctl is-active mysqld
sudo systemctl is-enabled mysqld
```

### 5. Login ke MySQL sebagai root

Ambil password root terlebih dahulu:

```bash
sudo cat /root/mysql-root-password.txt
```

Lalu login:

```bash
mysql -uroot -p
```

### 6. Verifikasi versi MySQL

```bash
mysql -uroot -p -e "SELECT VERSION();"
```

## Password Root

Password root disimpan di:

```bash
/root/mysql-root-password.txt
```

Hak akses file:

```bash
sudo ls -l /root/mysql-root-password.txt
```

## Service Management

Start:

```bash
sudo systemctl start mysqld
```

Stop:

```bash
sudo systemctl stop mysqld
```

Restart:

```bash
sudo systemctl restart mysqld
```

Lihat log:

```bash
sudo journalctl -u mysqld -n 100 --no-pager
```

## Security Notes

- Saat instalasi selesai, root MySQL saya set agar tidak bisa login tanpa password.
- Root login saat ini hanya ditujukan untuk administrasi lokal.
- Firewall belum membuka port `3306`, jadi akses remote tetap tertahan.
- Walau begitu, service MySQL saat ini listen di semua interface. Jika nanti port firewall dibuka, host lain di jaringan bisa mencoba konek.

Kalau nanti aplikasi Anda perlu koneksi dari mesin lain, lakukan secara sadar:

1. buat user database khusus aplikasi
2. buka firewall port `3306` bila memang perlu
3. batasi host koneksi user MySQL
4. jangan pakai user `root` untuk aplikasi

## Recommended Next Step

Setelah install MySQL, langkah berikut yang disarankan:

### Buat database aplikasi

```sql
CREATE DATABASE siakad_db
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;
```

### Buat user aplikasi

```sql
CREATE USER 'siakad_user'@'localhost' IDENTIFIED BY 'ganti-password-aman';
GRANT ALL PRIVILEGES ON siakad_db.* TO 'siakad_user'@'localhost';
FLUSH PRIVILEGES;
```

### Import schema awal

Jalankan dari laptop / host Anda:

```bash
mysql -h 192.168.64.5 -uroot -p siakad_db < db/schema.sql
```

Atau langsung dari dalam VM setelah file schema dipindahkan ke sana:

```bash
mysql -uroot -p siakad_db < schema.sql
```

## Useful Checks

Lihat port MySQL:

```bash
sudo ss -ltnp | grep 3306
```

Lihat package terpasang:

```bash
rpm -qa | grep '^mysql'
```

Tes login:

```bash
mysql -uroot -p -e "SHOW DATABASES;"
```
