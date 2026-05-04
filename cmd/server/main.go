package main

import (
	"context"
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Burgatski/pack-calculator/internal/handler"
	"github.com/Burgatski/pack-calculator/internal/storage"
)

var webFiles embed.FS

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	store := storage.NewMemory()
	h := handler.New(store)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	webRoot, err := fs.Sub(webFiles, "web")
	if err != nil {
		slog.Error("failed to create web sub-FS", "error", err)
		os.Exit(1)
	}
	mux.Handle("GET /", http.FileServer(http.FS(webRoot)))

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	shutdownDone := make(chan struct{})
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		slog.Info("shutting down")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("shutdown error", "error", err)
		}
		close(shutdownDone)
	}()

	slog.Info("server starting", "addr", "http://localhost:"+port)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
	<-shutdownDone
	slog.Info("server stopped")
}
