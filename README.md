# Transpara — Sistem Manajemen Kas dan Transparansi Keuangan Kelompok

Backend REST API menggunakan **Go (Golang)**, **MySQL**, dan **Clean Architecture**.

Transpara (Transparent Group Financial Management System) adalah sistem pengelolaan keuangan kelompok berbasis digital untuk meningkatkan transparansi, akuntabilitas, dan pengawasan atas seluruh aktivitas keuangan kelompok atau organisasi. Bendahara mengelola kas, sedangkan setiap anggota dapat memantau kondisi keuangan secara langsung.

---

## Anggota Tim dan Pembagian Modul

| Anggota | Peran | Modul |
|---------|-------|-------|
| Naisya Gina Azzahra (434241068) | Git Master | Modul Autentikasi & Keamanan (registrasi, verifikasi OTP, login, JWT) |
| Bunga Lala Sal Sabila (434241066) | Developer | Modul Transaksi Keuangan |
| Naily Nugraheni (434241086)  | Developer | Modul Anggaran dan Voting |

---

## Role pada Sistem

| Role | Hak akses utama |
|------|-----------------|
| Bendahara (Admin) | Mengelola data anggota, iuran, transaksi, anggaran, pengajuan pengeluaran, reimbursement, dan laporan |
| Anggota (User) | Melihat saldo dan riwayat transaksi, memantau iuran, mengunggah bukti pembayaran, mengajukan reimbursement, memberi persetujuan pengeluaran |

---

## Fitur

| No | Fitur | Role | Status |
|----|-------|------|--------|
| 1 | Registrasi, verifikasi OTP, dan login | Semua | Berjalan |
| 2 | Manajemen transaksi keuangan | Bendahara | Dalam pengerjaan |
| 3 | Perencanaan dan pengawasan anggaran | Bendahara | Kerangka |
| 4 | Pengajuan dan persetujuan pengeluaran (voting) | Bendahara, Anggota | Kerangka |
| 5 | Manajemen iuran anggota | Bendahara, Anggota | Rencana |
| 6 | Reimbursement | Bendahara, Anggota | Rencana |
| 7 | Laporan dan audit trail | Bendahara, Anggota | Rencana |

---

## Rancangan Database

| Tabel | Kolom utama | Status |
|-------|-------------|--------|
| users | id, name, email, password_hash, is_verified, created_at, updated_at | Ada (migrasi 001, 002) |
| otp_tokens | id, user_id, otp_hash, expires_at, is_used, attempt_count, created_at | Ada (migrasi 003) |
| categories | id, group_id, name, type | Rencana |
| transactions | id, group_id, category_id, created_by, type, amount, description, transaction_date | Rencana |

> Tabel lain (groups, memberships, budgets, expense requests, approvals, reimbursements, audit logs) akan ditambahkan oleh modul masing-masing.
>
> **ERD:** [lampirkan gambar ERD di sini, misalnya `docs/erd.png`]

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
git clone https://github.com/naisyazh/transpara-backend.git
cd transpara-backend
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
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASS=password_mysql_anda
DB_NAME=transpara
JWT_SECRET=isi_dengan_string_acak_yang_panjang
```

Variabel lain (port aplikasi, SMTP untuk pengiriman OTP, dan sebagainya) mengikuti daftar di `.env.example`.

### 3. Buat Database dan Tabel

Masuk ke MySQL lalu buat database:

```sql
CREATE DATABASE transpara CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

Jalankan file migrasi di folder `migrations/` secara berurutan (001, 002, 003, dan seterusnya):

```bash
# macOS / Linux / Git Bash
mysql -u root -p transpara < migrations/001_create_users_table.sql
mysql -u root -p transpara < migrations/002_add_verified_to_users.sql
mysql -u root -p transpara < migrations/003_create_otp_tokens_table.sql
```

```powershell
# Windows (PowerShell)
Get-Content migrations\001_create_users_table.sql | mysql -u root -p transpara
Get-Content migrations\002_add_verified_to_users.sql | mysql -u root -p transpara
Get-Content migrations\003_create_otp_tokens_table.sql | mysql -u root -p transpara
```

Alternatif: buka tiap file di MySQL Workbench lalu jalankan berurutan.

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

```bash
# Windows (PowerShell)
curl.exe http://localhost:8080/api/v1/health

# macOS / Linux
curl http://localhost:8080/api/v1/health
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
transpara-backend/
├── internal/
│   ├── domain/            # Entitas dan interface (kontrak) — diisi tiap modul
│   ├── repository/        # Implementasi akses database
│   ├── usecase/           # Business logic
│   ├── delivery/          # Penerima request HTTP dan response
│   │   ├── middleware/    # Middleware HTTP (JWT auth, dll)
│   │   └── router/        # Pendaftaran semua route API
│   └── infrastructure/
│       └── email/         # Pengiriman email OTP via SMTP
├── pkg/
│   ├── config/            # Membaca konfigurasi dari file .env
│   ├── database/          # Koneksi ke database MySQL
│   └── helper/            # Utilitas umum (JWT, dll)
├── migrations/            # File SQL pembuat tabel, berurutan
├── main.go                # Entry point — merakit semua komponen
├── .env                   # Konfigurasi lokal (JANGAN di-commit)
├── .env.example           # Contoh konfigurasi (aman di-commit)
├── .gitignore             # Daftar file yang dikecualikan dari Git
└── go.mod                 # Daftar dependency Go
```

---

## Endpoint API

| Method | Endpoint | Deskripsi | Auth |
|--------|----------|-----------|------|
| GET | `/api/v1/health` | Cek status server | Tidak |
| POST | `/api/v1/auth/register` | Registrasi user baru | Tidak |
| POST | `/api/v1/auth/verify-otp` | Verifikasi OTP yang dikirim ke email | Tidak |
| POST | `/api/v1/auth/resend-otp` | Kirim ulang OTP | Tidak |
| POST | `/api/v1/auth/login` | Login dan dapatkan token | Tidak |
| GET | `/api/v1/auth/me` | Data user yang sedang login | JWT |

> Endpoint modul transaksi (`/api/v1/transactions`) dan anggaran (`/api/v1/budgets`) akan ditambahkan oleh masing-masing developer.

---

## Alur Kerja Tim (Git)

- Setiap anggota bekerja di branch sendiri (contoh: `feat/transaksikeuangan-domain`) dan membuat commit kecil dengan pesan yang jelas, misalnya `feat(domain): add Transaction entity`.
- Semua perubahan masuk ke `main` lewat Pull Request, dan hanya Git Master yang melakukan merge (opsi *Create a merge commit*).
- File `.env` asli tidak boleh di-commit.

---

## Teknologi yang Digunakan

- **Go** — bahasa pemrograman utama
- **Fiber v2** — HTTP framework REST API
- **MySQL** — database relasional
- **godotenv** — membaca file .env
- **bcrypt** — hash password
- **golang-jwt** — JWT token
- **SMTP (Gmail)** — pengiriman OTP lewat email