# Transpara — Sistem Manajemen Kas dan Transparansi Keuangan Kelompok

Backend REST API menggunakan **Go (Golang)**, **MySQL**, dan **Clean Architecture**.

---

## Anggota Tim dan Pembagian Modul

| Developer | Modul |
|-----------|-------|
| Developer 1 | Modul Autentikasi & Keamanan (registrasi, login, JWT) |
| Developer 2 | Modul Transaksi Keuangan |
| Developer 3 | Modul Anggaran dan Voting |

---

## Prasyarat

Pastikan sudah terinstal di komputer Anda:

- [Go](https://go.dev/dl/) versi 1.21 atau lebih baru
- [MySQL](https://dev.mysql.com/downloads/) versi 8.0 atau lebih baru
- [Git](https://git-scm.com/)

---

## Cara Menjalankan Project

### 1. Clone Repository

```bash
git clone https://github.com/<username>/transpara.git
cd transpara
```

### 2. Siapkan File `.env`

Salin file contoh dan isi dengan konfigurasi lokal Anda:

```bash
# Windows (PowerShell)
Copy-Item .env.example .env

# macOS / Linux
cp .env.example .env
```

Buka file `.env` dan isi nilai yang sesuai:

```env
DB_USER=root
DB_PASS=password_mysql_anda
DB_NAME=transpara
JWT_SECRET=isi_dengan_string_acak_yang_panjang
```

### 3. Buat Database MySQL

Masuk ke MySQL dan jalankan perintah berikut:

```sql
CREATE DATABASE transpara CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 4. Install Dependency Go

```bash
go mod tidy
```

### 5. Jalankan Server

```bash
go run main.go
```

Server akan berjalan di `http://localhost:8080`.

### 6. Uji dengan Health Check

Buka browser atau Postman, akses:

```
GET http://localhost:8080/api/v1/health
```

Respons yang diharapkan:

```json
{
  "success": true,
  "message": "Server Transpara berjalan dengan baik"
}
```

---

## Struktur Folder

```
transpara/
├── config/           # Membaca konfigurasi dari file .env
├── database/         # Koneksi ke database MySQL
├── domain/           # Struct entitas dan interface (kontrak) — isi oleh tiap modul
├── repository/       # Implementasi akses database — isi oleh tiap modul
├── usecase/          # Business logic / Use Case — isi oleh tiap modul
├── middleware/        # Middleware HTTP (JWT auth, dll)
├── router/           # Pendaftaran semua route API
├── main.go           # Entry point — merakit semua komponen
├── .env              # Konfigurasi lokal (JANGAN di-commit)
├── .env.example      # Contoh konfigurasi (aman di-commit)
├── .gitignore        # Daftar file yang dikecualikan dari Git
└── go.mod            # Daftar dependency Go
```

---

## Endpoint API

| Method | Endpoint | Deskripsi | Auth |
|--------|----------|-----------|------|
| GET | `/api/v1/health` | Cek status server | Tidak |
| POST | `/api/v1/auth/register` | Registrasi user baru | Tidak |
| POST | `/api/v1/auth/login` | Login dan dapatkan token | Tidak |
| GET | `/api/v1/auth/me` | Data user yang sedang login | JWT |

> Endpoint auth akan tersedia setelah Modul Autentikasi selesai.

---

## Teknologi yang Digunakan

- **Go** — bahasa pemrograman utama
- **Fiber v2** — HTTP framework REST API
- **MySQL** — database relasional
- **godotenv** — membaca file .env
- **bcrypt** — hash password (akan ditambahkan di modul auth)
- **golang-jwt** — JWT token (akan ditambahkan di modul auth)
