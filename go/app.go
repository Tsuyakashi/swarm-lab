package main

import (
	"embed"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"
)

//go:embed templates/*.html
var templateFS embed.FS

var tmpl = template.Must(template.ParseFS(templateFS, "templates/*.html"))

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		next.ServeHTTP(w, r)

		slog.Info("Request handled",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
			"duration", time.Since(start).String(),
		)
	})
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := struct {
		Greeting  string
		Timestamp string
	}{
		Greeting:  "世界",
		Timestamp: time.Now().Format(time.RFC3339),
	}
	if err := tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
		slog.Error("Failed to render template", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "UP"}`))
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-healthcheck" {
		resp, err := http.Get("http://localhost:8080/health")
		if err != nil {
			os.Exit(1)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			os.Exit(1)
		}
		os.Exit(0)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	mux.HandleFunc("/health", healthHandler)

	wrappedMux := loggingMiddleware(mux)

	slog.Info("Starting production server", "port", "8080")

	server := &http.Server{
		Addr:         ":8080",
		Handler:      wrappedMux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
