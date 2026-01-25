package storage

import (
	"database/sql"
	"errors"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
)

// User represents a user in the system
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // Never expose in JSON
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserStore handles user persistence
type UserStore struct {
	db *sql.DB
}

// NewUserStore creates a new user store
func NewUserStore(dbPath string) (*UserStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, err
	}

	store := &UserStore{db: db}
	if err := store.init(); err != nil {
		return nil, err
	}

	return store, nil
}

// init creates the users table if it doesn't exist
func (s *UserStore) init() error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_username ON users(username);
	`
	_, err := s.db.Exec(query)
	return err
}

// CreateUser creates a new user with hashed password
func (s *UserStore) CreateUser(username, password string) (*User, error) {
	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Insert user
	result, err := s.db.Exec(
		"INSERT INTO users (username, password_hash) VALUES (?, ?)",
		username, string(hash),
	)
	if err != nil {
		// SQLite constraint error (UNIQUE)
		return nil, ErrUserAlreadyExists
	}

	id, _ := result.LastInsertId()
	return s.GetByID(int(id))
}

// GetByID retrieves a user by ID
func (s *UserStore) GetByID(id int) (*User, error) {
	user := &User{}
	err := s.db.QueryRow(
		"SELECT id, username, password_hash, created_at, updated_at FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetByUsername retrieves a user by username
func (s *UserStore) GetByUsername(username string) (*User, error) {
	user := &User{}
	err := s.db.QueryRow(
		"SELECT id, username, password_hash, created_at, updated_at FROM users WHERE username = ?",
		username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Authenticate verifies username and password
func (s *UserStore) Authenticate(username, password string) (*User, error) {
	user, err := s.GetByUsername(username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

// UpdatePassword updates a user's password
func (s *UserStore) UpdatePassword(userID int, newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(
		"UPDATE users SET password_hash = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		string(hash), userID,
	)
	return err
}

// EnsureDefaultUser creates the default user if no users exist
func (s *UserStore) EnsureDefaultUser(username, password string) error {
	// Check if any users exist
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return err
	}

	// If users exist, don't create default
	if count > 0 {
		return nil
	}

	// Create default user
	_, err = s.CreateUser(username, password)
	return err
}

// Close closes the database connection
func (s *UserStore) Close() error {
	return s.db.Close()
}
