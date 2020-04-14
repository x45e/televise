package televise

import (
	"context"
	"database/sql"
	"log"
	"net/http"

	// DB driver
	_ "github.com/denisenkom/go-mssqldb"
)

// Config defines configuration options for the Televise App.
type Config struct {
	Addr string
	DB string
}

// App represents a Televise application.
type App struct {
	srv *http.Server
	db *sql.DB
}

// Start starts a new app instance.
func Start(cfg Config) (*App, error) {
	db, err := sql.Open("sqlserver", cfg.DB)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	ctx = context.Background()
	ctx = context.WithValue(ctx, KeyDB, db)

	log.Println("Running migrations...")
	err = Migrate(db)
	if err != nil {
		return nil, err
	}
	log.Println("Finished migrating")

	app := &App{
		srv: &http.Server{Addr: cfg.Addr},
		db: db,
	}

	app.RegisterRoutes(ctx)

	err = app.srv.ListenAndServe()
	if err != http.ErrServerClosed {
		return nil, err
	}
	return app, nil
}

// Close closes the HTTP server and DB connections.
func (app *App) Close() error {
	if app.srv != nil {
		err := app.srv.Shutdown(context.Background())
		if err != nil {
			return err
		}
		app.srv = nil
	}
	if app.db != nil {
		err := app.db.Close()
		if err != nil {
			return err
		}
		app.db = nil
	}
	return nil
}

func allowAll(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	})
}

func (App) RegisterRoutes(ctx context.Context) {
	http.Handle("/info", allowAll(withContext(ctx, http.HandlerFunc(HandleInfo))))
	http.Handle("/update", allowAll(withContext(ctx, http.HandlerFunc(MetadataUpdate))))
}