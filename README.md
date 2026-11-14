# Transpara — Sistem Manajemen Kas dan Transparansi Keuangan Kelompok

Backend REST API menggunakan **Go (Golang)**, **MySQL**, dan **Clean Architecture**.

Transpara (Transparent Group Financial Management System) adalah sistem pengelolaan keuangan kelompok berbasis digital untuk meningkatkan transparansi, akuntabilitas, dan pengawasan atas seluruh aktivitas keuangan kelompok atau organisasi. Bendahara mengelola kas, sedangkan setiap anggota dapat memantau kondisi keuangan secara langsung.

---

## Anggota Tim dan Pembagian Modul

| Anggota | Peran | Modul |
|---------|-------|-------|
| Naisya Gina Azzahra (434241068) | Git Master | Modul Autentikasi & Keamanan (registrasi, verifikasi OTP, login, JWT) |
| Bunga Lala Sal Sabila (434241066) | Developer | Modul Transaksi Keuangan |
| Naily Nugraheni (434241086) | Developer | Modul Anggaran dan Voting |

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
| 3 | Perencanaan dan pengawasan anggaran | Bendahara | Berjalan |
| 4 | Pengajuan dan persetujuan pengeluaran (voting) | Bendahara, Anggota | Berjalan |
| 5 | Manajemen iuran anggota | Bendahara, Anggota | Rencana |
| 6 | Reimbursement | Bendahara, Anggota | Rencana |
| 7 | Laporan dan audit trail | Bendahara, Anggota | Rencana |

---

## Rancangan Database

| Tabel | Kolom utama | Status |
|-------|-------------|--------|
| users | id, name, email, password_hash, is_verified, role, created_at, updated_at | Ada (migrasi 001, 002, 007) |
| otp_tokens | id, user_id, otp_hash, expires_at, is_used, attempt_count, created_at | Ada (migrasi 003) |
| budgets | pos anggaran: nama, batas dana, periode | Ada (migrasi 004) |
| expense_requests | pengajuan dana beserta status | Ada (migrasi 005) |
| expense_votes | suara setuju atau tolak per pengajuan | Ada (migrasi 006) |
| categories | id, group_id, name, type | Rencana |
| transactions | id, group_id, category_id, created_by, type, amount, description, transaction_date | Rencana |

> Tabel lain (groups, memberships, reimbursements, audit logs) akan ditambahkan oleh modul masing-masing.
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

Jalankan semua file di folder `migrations/` secara berurutan (001, 002, dan seterusnya):

```bash
# macOS / Linux / Git Bash
for f in migrations/*.sql; do mysql -u root -p transpara < "$f"; done
```

```powershell
# Windows (PowerShell)
Get-ChildItem migrations\*.sql | Sort-Object Name | ForEach-Object {
  Get-Content $_.FullName | mysql -u root -p transpara
}
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
│   ├── domain/            # Entitas, DTO, interface repository, dan error khusus
│   ├── repository/        # Implementasi akses database MySQL
│   ├── usecase/           # Business logic
│   ├── delivery/          # Penerima request HTTP dan pengirim response
│   │   ├── middleware/    # RequireAuth (JWT) dan RequireRole (pembatasan role)
│   │   └── router/        # Pendaftaran semua route API
│   └── infrastructure/
│       └── email/         # Pengiriman email OTP via SMTP
├── pkg/
│   ├── config/            # Membaca konfigurasi dari file .env
│   ├── database/          # Koneksi ke database MySQL
│   └── helper/            # Utilitas umum (JWT)
├── migrations/            # File SQL migrasi (001-007), jalankan berurutan
├── docs/
│   └── postman/           # Koleksi Postman untuk pengujian manual
├── main.go                # Entry point, merakit semua komponen
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

> Endpoint modul anggaran dan voting ada di bagian **Modul Anggaran dan Voting** di bawah. Endpoint modul transaksi (`/api/v1/transactions`) akan ditambahkan oleh developer modul transaksi.

---

## Alur Kerja Tim (Git)

- Setiap anggota bekerja di branch sendiri (contoh: `feat/transaksikeuangan-domain`) dan membuat commit kecil dengan pesan yang jelas, misalnya `feat(domain): add Transaction entity`.
- Semua perubahan masuk ke `main` lewat Pull Request, dan hanya Git Master yang melakukan merge (opsi *Create a merge commit*).
- File `.env` asli tidak boleh di-commit.

---

## Modul Anggaran dan Voting

Bendahara membuat pos anggaran, anggota mengajukan penggunaan dana, lalu seluruh anggota memberi suara. Pengajuan yang disetujui mayoritas baru boleh dicairkan oleh modul transaksi.

### Persiapan Database

Migrasi modul ini ada di `migrations/004` sampai `migrations/007` dan ikut dijalankan pada langkah 3 di atas. Kalau ingin menjalankannya terpisah, pakai `SOURCE` di MySQL client (jalan juga di Windows PowerShell):

```sql
USE transpara;
SOURCE migrations/004_create_budgets_table.sql;
SOURCE migrations/005_create_expense_requests_table.sql;
SOURCE migrations/006_create_expense_votes_table.sql;
SOURCE migrations/007_add_role_to_users.sql;
```

Setiap user baru otomatis berrole `anggota`. Role `bendahara` hanya bisa disetel langsung lewat database:

```sql
UPDATE users SET role = 'bendahara' WHERE email = 'email_bendahara@contoh.com';
```

Perubahan role langsung berlaku tanpa login ulang, karena role dibaca dari database pada setiap request.

### Endpoint

Semua endpoint wajib memakai header `Authorization: Bearer <token>`.

| Method | Endpoint | Akses | Deskripsi |
|--------|----------|-------|-----------|
| GET | `/api/v1/budgets` | Login | Daftar pos anggaran beserta sisa dana |
| GET | `/api/v1/budgets/:id` | Login | Detail satu pos anggaran |
| POST | `/api/v1/budgets` | Bendahara | Buat pos anggaran |
| PUT | `/api/v1/budgets/:id` | Bendahara | Ubah pos anggaran |
| GET | `/api/v1/expense-requests` | Login | Daftar pengajuan, filter `?status=pending\|approved\|rejected` |
| GET | `/api/v1/expense-requests/:id` | Login | Detail pengajuan beserta hitungan suara |
| POST | `/api/v1/expense-requests` | Login | Buat pengajuan dana |
| POST | `/api/v1/expense-requests/:id/vote` | Login | Beri suara setuju atau tolak |

### Contoh Request

Buat pos anggaran (`POST /api/v1/budgets`, tanggal berformat `YYYY-MM-DD`):

```json
{
  "name": "Kegiatan Bakti Sosial",
  "description": "Dana untuk kegiatan bulan November",
  "max_amount": 1000000,
  "start_date": "2026-11-01",
  "end_date": "2026-11-30"
}
```

Buat pengajuan dana (`POST /api/v1/expense-requests`):

```json
{
  "budget_id": 1,
  "title": "Beli konsumsi",
  "amount": 500000,
  "reason": "Konsumsi untuk peserta kegiatan"
}
```

Beri suara (`POST /api/v1/expense-requests/:id/vote`, nilai `yes` atau `no`):

```json
{
  "vote": "yes"
}
```

Format response sama dengan modul lain: `success`, `message`, dan `data` (hanya saat sukses).

### Aturan Bisnis

- **Pembatasan role**: hanya bendahara yang boleh membuat dan mengubah pos anggaran. Anggota biasa menerima `403 Forbidden`.
- **Validasi pengajuan**: pos anggaran harus ada, tanggal hari ini harus berada di dalam periode pos (tanggal akhir ikut dihitung), dan nominal tidak boleh melebihi sisa dana. Status awal pengajuan adalah `pending`.
- **Sisa dana** = `max_amount` dikurangi total pengajuan berstatus `pending` dan `approved`. Pengajuan `rejected` tidak dihitung.
- **Total anggota voting** = semua user dengan `is_verified = 1`, bendahara ikut dihitung.
- **Disetujui** (`approved`): suara setuju melewati 50% total anggota. Contoh 4 anggota, 3 suara setuju.
- **Ditolak** (`rejected`): suara tolak sudah mencapai 50% total anggota, sehingga suara setuju tidak mungkin melewati 50% lagi. Contoh 4 anggota, 2 suara tolak.
- Saat status berubah menjadi `approved` atau `rejected`, kolom `decided_at` otomatis terisi.
- Satu user hanya boleh memberi satu suara per pengajuan, dan vote hanya bisa dilakukan saat status masih `pending`.
- Proses vote berjalan dalam satu DB transaction dengan penguncian baris, sehingga hitungan tetap akurat walau dua orang vote bersamaan.
- Pos anggaran tidak bisa diubah ke `max_amount` yang lebih kecil dari dana yang sudah diajukan.

### Kode Response

| Kode | Arti |
|------|------|
| 400 | Input tidak valid (format JSON, validasi field, format tanggal, nilai `vote`) |
| 401 | Token tidak ada atau tidak valid |
| 403 | Role tidak punya akses (anggota mengakses fitur bendahara) |
| 404 | Pos anggaran atau pengajuan tidak ditemukan |
| 409 | Sudah vote, atau voting sudah ditutup karena status bukan `pending` |
| 422 | Aturan bisnis tidak terpenuhi (di luar periode, melebihi sisa dana, dan sejenisnya) |

### Penghubung ke Modul Transaksi

Pencairan dana hanya boleh dilakukan dari pengajuan berstatus `approved`. Modul transaksi memanggil `ExpenseUseCase.EnsureApproved(id)` sebelum mencairkan dana. Fungsi ini mengembalikan `domain.ErrRequestNotApproved` jika status masih `pending` atau `rejected` (dipetakan ke `422`), dan `domain.ErrExpenseNotFound` jika ID tidak ada.

### Pengujian Manual

Koleksi Postman tersedia di `docs/postman/transpara-anggaran-voting.postman_collection.json`. Siapkan tepat 4 akun terverifikasi (1 bendahara dan 3 anggota) sebelum menjalankannya.

---

## Teknologi yang Digunakan

- **Go** — bahasa pemrograman utama
- **Fiber v2** — HTTP framework REST API
- **MySQL** — database relasional
- **godotenv** — membaca file .env
- **bcrypt** — hash password
- **golang-jwt** — JWT token
- **SMTP (Gmail)** — pengiriman OTP lewat email