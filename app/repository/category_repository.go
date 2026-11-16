package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/transpara/backend/app/model"
)

type categoryRepository struct {
	db *sql.DB
}

// NewCategoryRepository membuat instance categoryRepository baru.
func NewCategoryRepository(db *sql.DB) model.CategoryRepository {
	return &categoryRepository{db: db}
}

const categoryColumns = `id, name, type, created_at, updated_at`

func scanCategory(s rowScanner) (*model.Category, error) {
	c := &model.Category{}
	if err := s.Scan(&c.ID, &c.Name, &c.Type, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return nil, err
	}
	return c, nil
}

// Create menyimpan kategori baru.
// UNIQUE (name, type) di database menolak kategori kembar.
func (r *categoryRepository) Create(c *model.Category) (*model.Category, error) {
	result, err := r.db.Exec(
		`INSERT INTO categories (name, type) VALUES (?, ?)`,
		c.Name, c.Type,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return nil, model.ErrCategoryDuplicate
		}
		return nil, fmt.Errorf("gagal menyimpan kategori: %w", err)
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil ID kategori baru: %w", err)
	}

	return r.FindByID(uint64(lastID))
}

// FindByID mencari kategori berdasarkan ID.
func (r *categoryRepository) FindByID(id uint64) (*model.Category, error) {
	query := `SELECT ` + categoryColumns + ` FROM categories WHERE id = ? LIMIT 1`

	c, err := scanCategory(r.db.QueryRow(query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrCategoryNotFound
		}
		return nil, fmt.Errorf("gagal mencari kategori: %w", err)
	}
	return c, nil
}

// FindAll mengembalikan daftar kategori, diurutkan per tipe lalu nama.
// Jika txType tidak kosong, hanya kategori dengan tipe itu yang diambil.
func (r *categoryRepository) FindAll(txType string) ([]model.Category, error) {
	query := `SELECT ` + categoryColumns + ` FROM categories`
	args := []any{}

	if txType != "" {
		query += ` WHERE type = ?`
		args = append(args, txType)
	}
	query += ` ORDER BY type, name`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar kategori: %w", err)
	}
	defer rows.Close()

	list := make([]model.Category, 0) // supaya JSON jadi [] bukan null
	for rows.Next() {
		c, err := scanCategory(rows)
		if err != nil {
			return nil, fmt.Errorf("gagal membaca kategori: %w", err)
		}
		list = append(list, *c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("gagal membaca daftar kategori: %w", err)
	}
	return list, nil
}