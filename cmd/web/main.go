package main

import (
	"context"
	"database/sql"
	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"html/template"
	"time"
	"wa3wa3.snippetbox/internal/models"

	"flag"
	"log/slog"
	"net/http"
	"os"
)

type application struct {
	logger         *slog.Logger
	snippets       models.SnippetModelInterface
	users          models.UserModelInterface
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
}

func main() {

	addr := flag.String("addr", ":4000", "HTTP network address")
	dsn := flag.String("dsn", "postgres://web:pass@127.0.0.1:5432/snippetbox", "postgres data source name")

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}))

	db, err := openDB(*dsn)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()

	// initialize and cache templates into the app as a dependency
	templateCache, err := newTemplateCache()

	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	formDecoder := form.NewDecoder()

	sessionManager := scs.New()
	pool, err := pgxpool.New(context.Background(), "postgres://web:pass@127.0.0.1:5432/snippetbox")
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	defer pool.Close()
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Store = pgxstore.New(pool)
	sessionManager.Cookie.Secure = true
	app := &application{
		logger,
		&models.SnippetModel{DB: db},
		&models.UserModel{DB: db},
		templateCache,
		formDecoder,
		sessionManager,
	}

	logger.Info("starting server", "addr", *addr)
	server := http.Server{
		Addr:         *addr,
		Handler:      app.routes(),
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	err = server.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")

	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
