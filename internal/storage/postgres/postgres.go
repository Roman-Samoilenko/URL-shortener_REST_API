package postgres

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strings"
	"time"
	"unicode"

	"github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "shortener"
	charset  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	keyLen   = 7
)

var (
	db         *sql.DB
	seededRand *rand.Rand
)

func Init() error {
	seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

	psqlInfo := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		return err
	}

	if err = db.Ping(); err != nil {
		return err
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS urls (
		id SERIAL PRIMARY KEY,
		original_url TEXT NOT NULL,
		short_key TEXT NOT NULL UNIQUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err = db.Exec(createTableSQL); err != nil {
		return err
	}

	return nil
}

func generateKey() string {
	b := make([]byte, keyLen)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func SaveURL(originalURL string) (string, error) {
	var shortKey string

	for i := 0; i < 5; i++ {
		shortKey = generateKey()
		_, err := db.Exec(
			"INSERT INTO urls (original_url, short_key) VALUES ($1, $2)",
			strings.TrimRightFunc(originalURL, unicode.IsSpace), shortKey,
		)

		if err == nil {
			return shortKey, nil
		}

		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			continue
		} else {
			return "", err
		}
	}

	return "", fmt.Errorf("failed to generate unique key")
}

func GetOriginalURL(shortKey string) (string, error) {
	var originalURL string
	err := db.QueryRow(
		"SELECT original_url FROM urls WHERE short_key = $1",
		shortKey,
	).Scan(&originalURL)
	return originalURL, err
}

func GetShortKey(OriginalURL string) (string, error) {
	var shortKey string
	err := db.QueryRow(
		"SELECT short_key FROM urls WHERE original_url = $1",
		OriginalURL,
	).Scan(&shortKey)
	return shortKey, err
}
