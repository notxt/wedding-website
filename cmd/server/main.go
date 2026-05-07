package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/notxt/wedding-website/internal/config"
	"github.com/notxt/wedding-website/internal/handlers"
	"github.com/notxt/wedding-website/internal/static"
	"github.com/notxt/wedding-website/internal/templates"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()

	tmpls, err := templates.Load()
	if err != nil {
		logger.Error("template load failed", "err", err)
		os.Exit(1)
	}

	staticHandler, err := static.Handler()
	if err != nil {
		logger.Error("static handler init failed", "err", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.Handle("GET /static/", staticHandler)
	mux.HandleFunc("GET /{$}", handlers.Home(tmpls))
	mux.HandleFunc("GET /faqs", handlers.FAQs(tmpls))
	mux.HandleFunc("GET /information", handlers.Page(tmpls, "information.html"))
	mux.HandleFunc("GET /information/itinerary", handlers.Page(tmpls, "information-itinerary.html"))
	mux.HandleFunc("GET /information/travel", handlers.Page(tmpls, "information-travel.html"))
	mux.HandleFunc("GET /information/things-to-do", handlers.Page(tmpls, "information-things-to-do.html"))

	addr := net.JoinHostPort("", cfg.Port)
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("server listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		if err != nil {
			logger.Error("server failed", "err", err)
			os.Exit(1)
		}
	case sig := <-stop:
		logger.Info("shutdown signal received", "signal", sig.String())
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			logger.Error("graceful shutdown failed", "err", err)
			os.Exit(1)
		}
		logger.Info("server stopped cleanly")
	}
}
