# Transpara — Sistem Manajemen Kas dan Transparansi Keuangan Kelompok

Backend REST API menggunakan **Go (Golang)**, **MySQL/MariaDB**, dan **Clean Architecture**.

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
| 2 | Manajemen transaksi keuangan | Bendahara | Berjalan |
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
| categories | id, name, type, created_at, updated_at | Ada (migrasi 008) |
| transactions | id, category_id, created_by, expense_request_id, type, amount, description, transaction_date, created_at, updated_at | Ada (migrasi 009) |

Catatan: sistem ini melayani satu kelompok, sehingga tidak ada kolom `group_id`. Batasan `CHECK` pada migrasi 004, 005, dan 009 hanya dijalankan oleh MariaDB 10.2.1+ atau MySQL 8.0.16+.

---

## Struktur Proyek

```
transpara-backend/
├── main.go
├── app/
│   ├── model/          # entitas, kontrak repository, DTO, dan error domain
│   ├── repository/     # akses database
│   └── service/        # logika bisnis
├── config/             # pembacaan .env
├── database/
│   ├── mysql.go        # koneksi database
│   └── migrations/     # 001 sampai 009
├── helper/
│   ├── jwt.go
│   └── email/smtp.go   # pengiriman OTP lewat Gmail SMTP
├── middleware/
│   ├── jwt_middleware.go
│   └── role_middleware.go
├── route/              # endpoint HTTP dan pemetaan error ke status HTTP
├── .env.example
└── README.md
```

---

## Cara Menjalankan

1. Clone repository, lalu masuk ke folder proyek.
2. Salin `.env.example` menjadi `.env` dan isi nilainya (jangan commit `.env`):
   ```dotenv
   APP_PORT=8080
   APP_NAME=Transpara API

   DB_HOST=localhost
   DB_PORT=3306
   DB_USER=root
   DB_PASS=your_password_here
   DB_NAME=transpara

   JWT_SECRET=ganti_dengan_secret_yang_kuat_dan_panjang
   JWT_ISSUER=transpara-api
   JWT_EXPIRED_HOUR=24

   SMTP_HOST=smtp.gmail.com
   SMTP_PORT=587
   SMTP_USERNAME=your_email@gmail.com
   SMTP_PASSWORD=your_gmail_app_password
   SMTP_SENDER_NAME=Transpara
   ```
   `SMTP_PASSWORD` adalah **App Password** Gmail 16 karakter (butuh 2-Step Verification), bukan password login. Tulis tanpa spasi.
3. Buat database dan jalankan migrasi `001` sampai `009` **berurutan** dari folder `database/migrations/`. Di MySQL Workbench: File > Open SQL Script, lalu jalankan satu per satu. Di mysql CLI, perintah `SOURCE` bisa dipakai.
4. Jalankan server:
   ```bash
   go run main.go
   ```
   Server berjalan di `http://localhost:8080`. Cek dengan `GET /api/v1/health`.

---

## Alur Kerja Tim (Git)

- Setiap anggota bekerja di branch sendiri (contoh: `feat/transaksikeuangan-domain`) dan membuat commit kecil dengan pesan yang jelas, misalnya `feat(domain): add Transaction entity`.
- Semua perubahan masuk ke `main` lewat Pull Request, dan hanya Git Master yang melakukan merge (opsi *Create a merge commit*).
- File `.env` asli tidak boleh di-commit.

---

## Modul Autentikasi

| Method | Endpoint | Akses | Deskripsi |
|--------|----------|-------|-----------|
| POST | `/api/v1/auth/register` | Publik | Daftar akun baru dan kirim OTP ke email |
| POST | `/api/v1/auth/verify-otp` | Publik | Verifikasi akun dengan kode OTP |
| POST | `/api/v1/auth/resend-otp` | Publik | Kirim ulang kode OTP |
| POST | `/api/v1/auth/login` | Publik | Login dan dapatkan token JWT |
| GET | `/api/v1/auth/me` | Login | Data pengguna dari token |

Setelah login, kirim token di setiap request terlindungi dengan header `Authorization: Bearer <token>`.

---

## Modul Anggaran dan Voting

Bendahara membuat pos anggaran, anggota mengajukan penggunaan dana, lalu seluruh anggota memberi suara. Pengajuan yang disetujui mayoritas baru boleh dicairkan oleh modul transaksi.

### Persiapan Database

Migrasi modul ini ada di `migrations/004` sampai `migrations/007` dan ikut dijalankan pada langkah 3 di atas. Kalau ingin menjalankannya terpisah dengan mysql CLI, pakai `SOURCE`:

```sql
USE transpara;
SOURCE database/migrations/004_create_budgets_table.sql;
SOURCE database/migrations/005_create_expense_requests_table.sql;
SOURCE database/migrations/006_create_expense_votes_table.sql;
SOURCE database/migrations/007_add_role_to_users.sql;
```

`SOURCE` hanya dikenali mysql CLI. Di MySQL Workbench, buka dan jalankan setiap berkas lewat File > Open SQL Script.

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

Pencairan dana hanya boleh dilakukan dari pengajuan berstatus `approved`. Modul transaksi memanggil `ExpenseService.EnsureApproved(id)` sebelum mencairkan dana. Fungsi ini mengembalikan `model.ErrRequestNotApproved` jika status masih `pending` atau `rejected` (dipetakan ke `422`), dan `model.ErrExpenseNotFound` jika ID tidak ada.

### Pengujian Manual

Koleksi Postman tersedia di `docs/postman/transpara-anggaran-voting.postman_collection.json`. Siapkan tepat 4 akun terverifikasi (1 bendahara dan 3 anggota) sebelum menjalankannya.

---

## Modul Transaksi Keuangan

Bendahara mencatat pemasukan dan pengeluaran kas, sedangkan seluruh anggota dapat melihat riwayat transaksi dan saldo. Pengeluaran **wajib** terhubung ke pengajuan dana berstatus `approved` dari modul anggaran, sehingga uang hanya keluar setelah disetujui anggota.

### Persiapan Database

Migrasi modul ini ada di `database/migrations/008` dan `009`, dan harus dijalankan **setelah** migrasi 004 sampai 007 karena tabel `transactions` memiliki foreign key ke `expense_requests`.

```sql
USE transpara;
SOURCE database/migrations/008_create_categories_table.sql;
SOURCE database/migrations/009_create_transactions_table.sql;
```

Migrasi 008 sekaligus mengisi tujuh kategori awal, sehingga bendahara bisa langsung mencatat transaksi:

| Tipe | Kategori awal |
|------|---------------|
| income | Iuran Anggota, Donasi, Lain-lain |
| expense | Konsumsi, Perlengkapan, Kegiatan, Lain-lain |

### Endpoint

Semua endpoint wajib memakai header `Authorization: Bearer <token>`.

| Method | Endpoint | Akses | Deskripsi |
|--------|----------|-------|-----------|
| GET | `/api/v1/categories` | Login | Daftar kategori, filter `?type=income\|expense` |
| POST | `/api/v1/categories` | Bendahara | Buat kategori |
| GET | `/api/v1/transactions` | Login | Daftar transaksi, terbaru di atas |
| GET | `/api/v1/transactions/summary` | Login | Ringkasan total pemasukan, pengeluaran, dan saldo kas |
| GET | `/api/v1/transactions/:id` | Login | Detail satu transaksi |
| POST | `/api/v1/transactions` | Bendahara | Catat pemasukan atau pengeluaran |

Filter pada `GET /api/v1/transactions` (semuanya opsional dan bisa digabung):

| Parameter | Nilai | Keterangan |
|-----------|-------|------------|
| `type` | `income` atau `expense` | Saring berdasarkan tipe |
| `category_id` | angka | Saring berdasarkan kategori |
| `from` | `YYYY-MM-DD` | Tanggal awal, ikut dihitung |
| `to` | `YYYY-MM-DD` | Tanggal akhir, ikut dihitung |

### Contoh Request

Buat kategori (`POST /api/v1/categories`, `type` bernilai `income` atau `expense`):

```json
{
  "name": "Sponsor",
  "type": "income"
}
```

Catat pemasukan (`POST /api/v1/transactions`, `amount` wajib):

```json
{
  "type": "income",
  "category_id": 1,
  "amount": 2000000,
  "description": "Iuran bulan November",
  "transaction_date": "2026-11-15"
}
```

Catat pengeluaran dari pengajuan yang disetujui. Nominal **tidak** diisi karena diambil dari pengajuan:

```json
{
  "type": "expense",
  "category_id": 4,
  "expense_request_id": 1,
  "description": "Konsumsi kegiatan",
  "transaction_date": "2026-11-15"
}
```

Contoh data ringkasan saldo (`GET /api/v1/transactions/summary`):

```json
{
  "total_income": 2000000,
  "total_expense": 500000,
  "balance": 1500000
}
```

Format response sama dengan modul lain: `success`, `message`, dan `data` (hanya saat sukses).

### Aturan Bisnis

- **Pembatasan role**: hanya bendahara yang boleh membuat kategori dan mencatat transaksi. Anggota biasa menerima `403 Forbidden`, tetapi boleh melihat kategori, daftar transaksi, dan saldo.
- **Kategori**: harus ada, dan tipenya harus sama dengan tipe transaksi. Pasangan nama dan tipe kategori bersifat unik (huruf besar dan kecil tidak dibedakan), tetapi nama yang sama boleh ada di tipe yang berbeda.
- **Pemasukan** (`income`): `amount` wajib diisi dan harus lebih dari 0. Pemasukan tidak boleh menyertakan `expense_request_id`.
- **Pengeluaran** (`expense`): `expense_request_id` wajib diisi dan pengajuannya harus berstatus `approved` (dicek lewat `ExpenseService.EnsureApproved`). Nominal diambil dari pengajuan, bukan dari input, sehingga dana yang cair sama persis dengan yang disetujui anggota. Field `amount` pada request diabaikan.
- **Satu pengajuan hanya bisa dicairkan sekali.** Dijaga `UNIQUE KEY` di database, sehingga aman walau dua request masuk bersamaan.
- **Tanggal**: `transaction_date` berformat `YYYY-MM-DD`.
- **Saldo kas** = total pemasukan dikurangi total pengeluaran, dihitung di database.
- **Batasan saat ini**: pencairan hanya dibatasi status `approved` dan sisa dana pos anggaran. Saldo kas tidak diperiksa, sehingga `balance` bisa negatif jika pemasukan belum dicatat.
- Database menjaga aturan yang sama lewat `CHECK`: pengeluaran wajib punya pengajuan dan pemasukan tidak boleh punya.

### Kode Response

| Kode | Arti |
|------|------|
| 400 | Input tidak valid (format JSON, validasi field, format tanggal, nilai `type`, parameter filter, ID bukan angka) |
| 401 | Token tidak ada atau tidak valid |
| 403 | Role tidak punya akses (anggota mencatat transaksi atau membuat kategori) |
| 404 | Kategori, transaksi, atau pengajuan tidak ditemukan |
| 409 | Pengajuan sudah pernah dicairkan, atau kategori dengan nama dan tipe yang sama sudah ada |
| 422 | Aturan bisnis tidak terpenuhi (pengajuan belum disetujui, pengeluaran tanpa pengajuan, pemasukan dengan pengajuan, nominal pemasukan kosong, tipe kategori tidak sesuai) |

### Pengujian Manual

Koleksi Postman tersedia di `docs/postman/transpara-transaksi.postman_collection.json`. Siapkan tepat 4 akun terverifikasi (1 bendahara dan 3 anggota) sebelum menjalankannya, karena pencairan dana membutuhkan pengajuan yang disetujui voting. Urutan skenario: buat kategori, catat pemasukan, buat pos anggaran dan pengajuan, voting sampai `approved`, lalu cairkan dana dan cek saldo.

---

## Teknologi yang Digunakan

- **Go** — bahasa pemrograman utama
- **Fiber v2** — HTTP framework REST API
- **MySQL / MariaDB** — database relasional (MariaDB 10.2.1+ atau MySQL 8.0.16+ agar batasan `CHECK` aktif)
- **godotenv** — membaca file .env
- **bcrypt** — hash password dan OTP
- **golang-jwt** — JWT token
- **go-playground/validator** — validasi input request
- **SMTP (Gmail)** — pengiriman OTP lewat email