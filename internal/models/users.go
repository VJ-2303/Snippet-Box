package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             int
	Name           string
	Email          string
	HashedPassword []byte
	Created        time.Time
}

type UserModel struct {
	DB *sql.DB
}

func (m *UserModel) Insert(name, email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}
	query := `INSERT INTO users (name, email, hashed_password, created_at)
			  VALUES(?,?,?,UTC_TIMESTAMP())`

	_, err = m.DB.Exec(query, name, email, string(hashedPassword))
	if err != nil {
		var mySQLError *mysql.MySQLError
		if errors.As(err, &mySQLError) {
			if mySQLError.Number == 1062 && strings.Contains(mySQLError.Message, "users_uc_email") {
				return ErrDuplicateEmail
			}
		}
		return err
	}
	return nil
}

func (m *UserModel) Authenticate(email, pasword string) (int, error) {

	var id int
	var hashedPassword []byte

	query := `SELECT id, hashed_password FROM users WHERE email = ?`

	err := m.DB.QueryRow(query, email).Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(pasword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}
	return id, nil
}

func (m *UserModel) Exists(id int) (bool, error) {
	var exists bool

	query := `SELECT EXISTS(SELECT true FROM users WHERE id = ?)`

	err := m.DB.QueryRow(query, id).Scan(&exists)
	return exists, err
}

func (m *UserModel) Get(id int) (*User, error) {
	if id < 1 {
		return nil, ErrNoRecord
	}
	var user User

	query := `SELECT name, email, created_at FROM users WHERE id = ?`

	err := m.DB.QueryRow(query, id).Scan(
		&user.Name,
		&user.Email,
		&user.Created,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	return &user, nil
}

func (m *UserModel) PasswordChange(id int, currentPassword, newPassword string) error {

	var hashed_password string

	query := `SELECT hashed_password from users WHERE id = ?`

	err := m.DB.QueryRow(query, id).Scan(&hashed_password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNoRecord
		} else {
			return err
		}
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashed_password), []byte(currentPassword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidCredentials
		}
		return err
	}
	new_hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}
	query = `UPDATE users SET hashed_password = ? WHERE id = ?`
	_, err = m.DB.Exec(query, new_hash, id)
	if err != nil {
		return err
	}
	return nil
}
