package televise

import (
	"context"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
)

// Config defines configuration options for the Televise App.
type Config struct {
	Addr          string
	RedisAddr     string
	RedisPassword string
}

// App represents a Televise application.
type App struct {
	srv *http.Server
}

// Start starts a new app instance.
func Start(cfg Config) (*App, error) {
	ctx := context.Background()

	rc := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
	})
	if err := rc.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	PruneSessions(rc)

	ctx = context.WithValue(ctx, KeyRedis, rc)

	app := &App{
		srv: &http.Server{Addr: cfg.Addr},
	}

	var viewers int64

	go func(c *redis.Client, viewers *int64) {
		for {
			n, err := SessionCount(c)
			if err == nil {
				*viewers = n
			}
			time.Sleep(10 * time.Second)
			PruneSessions(c)
		}
	}(rc, &viewers)

	ctx = context.WithValue(ctx, KeyCount, &viewers)

	app.RegisterRoutes(ctx)

	err := app.srv.ListenAndServe()
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
	return nil
}

func allowAll(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	})
}

func (App) RegisterRoutes(ctx context.Context) {
	//http.Handle("/info", allowAll(withContext(ctx, http.HandlerFunc(HandleInfo))))
	//http.Handle("/update", allowAll(withContext(ctx, http.HandlerFunc(MetadataUpdate))))
	//http.Handle("/manifest", allowAll(withContext(ctx, http.HandlerFunc(HandleManifest))))
	//http.Handle("/vote", allowAll(withContext(ctx, http.HandlerFunc(HandleCastVote))))
	http.Handle("/token", allowAll(withContext(ctx, http.HandlerFunc(HandleToken))))
	http.Handle("/ping", allowAll(withContext(ctx, http.HandlerFunc(HandlePing))))
	http.Handle("/count", allowAll(withContext(ctx, http.HandlerFunc(HandleViewers))))
	//http.Handle("/results", allowAll(withContext(ctx, http.HandlerFunc(HandleLastResults))))
}
