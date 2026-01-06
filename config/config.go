package config

import (
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"log/slog"
	"os"

	_ "github.com/lib/pq"
	"ryannicoletti.net/internal/models"
	"ryannicoletti.net/posts"
)

type Config struct {
	Addr        string
	DSN         string
	Environment string
}

type Application struct {
	Logger        *slog.Logger
	DB            *sql.DB
	TemplateCache map[string]*template.Template
	Environment   string
	Posts         *models.PostModel
	Comments      *models.CommentModel
}

func Init() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.Addr, "addr", ":4000", "HTTP server port")
	flag.StringVar(&cfg.Environment, "env", "development", "Environment (development|production)")
	flag.Parse()

	err := cfg.loadDSN()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) loadDSN() error {
	if c.Environment == "production" {
		dsn := os.Getenv("PROD_DB_DSN")
		if dsn != "" {
			c.DSN = dsn
			return nil
		}
	}
	dsn := os.Getenv("DEV_DB_DSN")
	if dsn != "" {
		c.DSN = dsn
		return nil
	}
	return fmt.Errorf("database DSN not found")
}

func NewApplication(logger *slog.Logger, db *sql.DB, templateCache map[string]*template.Template, env string) *Application {
	return &Application{
		Logger:        logger,
		DB:            db,
		TemplateCache: templateCache,
		Environment:   env,
		Posts:         &models.PostModel{PostsFS: posts.Files},
		Comments:      &models.CommentModel{DB: db},
	}
}
