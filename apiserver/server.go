package apiserver

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/Deepjyoti-Sarmah/fast-api/config"
	"github.com/Deepjyoti-Sarmah/fast-api/store"
)

type ApiServer struct {
	config     *config.Config
	logger     *slog.Logger
	store      *store.Store
	JwtManager *JwtManager
}

func New(config *config.Config, logger *slog.Logger, store *store.Store, jwtManager *JwtManager) *ApiServer {
	return &ApiServer{
		config,
		logger,
		store,
		jwtManager,
	}
}

func (s *ApiServer) ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

func (s *ApiServer) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /ping", s.ping)
	mux.HandleFunc("POST /auth/signup", s.signupHandler())
	mux.HandleFunc("POST /auth/signin", s.signinHandler())
	mux.HandleFunc("POST /auth/refresh", s.tokenRefreshHandler())

	middleware := NextLoggerMiddleware(s.logger)
	middleware = NewAuthMiddleware(s.JwtManager, s.store.Users)

	server := &http.Server{
		Addr:    net.JoinHostPort(s.config.ApiServerHost, s.config.ApiServerPort),
		Handler: middleware(mux),
	}

	go func() {
		s.logger.Info("apiserver running", "port", s.config.ApiServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("ApiServer failed to listen and serve", "error", err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("apiserver failed to shutdown", "error", err)
		}
	}()

	wg.Wait()
	return nil
}
