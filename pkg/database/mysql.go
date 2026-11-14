package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql" // driver MySQL wajib di-import agar terdaftar
	"github.com/transpara/backend/config"
)

// NewMySQLConnection membuat dan mengembalikan koneksi database MySQL.
// Fungsi ini dipanggil sekali di main.go.
//
// Menggunakan *sql.DB bukan langsung koneksi tunggal, sehingga sudah
// mendukung connection pool secara otomatis (efisien untuk banyak request).
func NewMySQLConnection(cfg *config.Config) *sql.DB {
	// Format DSN (Data Source Name) untuk MySQL:
	// user:password@tcp(host:port)/dbname?parseTime=true
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=Local",
		cfg.DBUser,
		cfg.DBPass,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Gagal membuka koneksi database: %v", err)
	}

	// Uji koneksi — pastikan database benar-benar bisa dihubungi
	if err := db.Ping(); err != nil {
		log.Fatalf("Gagal terhubung ke database MySQL: %v\n"+
			"Pastikan MySQL berjalan dan konfigurasi .env sudah benar.", err)
	}

	// Konfigurasi connection pool
	db.SetMaxOpenConns(25)              // maksimal 25 koneksi terbuka sekaligus
	db.SetMaxIdleConns(5)               // maksimal 5 koneksi idle (siap pakai)
	db.SetConnMaxLifetime(5 * time.Minute) // koneksi diperbarui setiap 5 menit

	log.Println("Koneksi database MySQL berhasil")
	return db
}
