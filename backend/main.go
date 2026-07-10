package main

import (
	"bufio"
	"database/sql"
	"embed"
	"gateway/internal/admin"
	"gateway/internal/cache"
	"gateway/internal/proxy"
	"gateway/internal/storage"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

//go:embed all:out
var embedFS embed.FS

func loadEnv() {
	// Try loading from parent directory (.env in root) or local directory
	paths := []string{"../.env", ".env"}
	for _, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			continue
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				// Only set if not already set in system environment
				if os.Getenv(key) == "" {
					os.Setenv(key, val)
				}
			}
		}
		break // Stop once we successfully loaded one .env file
	}
}

func main() {
	loadEnv()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/gateway?sslmode=disable"
	}

	var db *sql.DB
	var err error

	log.Printf("Connecting to database at %s...", dbURL)
	for i := 1; i <= 5; i++ {
		db, err = sql.Open("postgres", dbURL)
		if err == nil {
			err = db.Ping()
			if err == nil {
				break
			}
		}
		log.Printf("Database connection attempt %d failed: %v. Retrying in 2 seconds...", i, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Failed to establish database connection after 5 attempts: %v", err)
	}
	defer db.Close()

	pgStorage, err := storage.NewPostgresStorage(db)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	var activeCache cache.Cache
	redisURL := os.Getenv("REDIS_URL")
	if redisURL != "" {
		log.Printf("Connecting to Redis at %s...", redisURL)
		activeCache, err = cache.NewRedisCache(redisURL)
		if err != nil {
			log.Fatalf("Failed to connect to Redis: %v", err)
		}
		log.Println("Redis Cache initialized successfully. Multi-Node HA enabled.")
	} else {
		log.Println("REDIS_URL not set. Using local Memory Cache. (Single-Node mode)")
		activeCache = cache.NewMemoryCache()
	}

	proxyServer := proxy.NewProxyServer(pgStorage, activeCache)
	adminHandler := admin.NewAdminHandler(pgStorage, activeCache)

	mux := http.NewServeMux()

	// Chat completions endpoint (with Auth & Budget Middlewares)
	proxyHandler := proxyServer.AuthMiddleware(proxyServer.BudgetMiddleware(proxyServer))
	mux.Handle("/v1/chat/completions", proxyHandler)

	// Admin Dashboard APIs
	mux.Handle("/api/", adminHandler)

	// Statically embedded Next.js frontend with route fallback
	staticFS, err := fs.Sub(embedFS, "out")
	if err != nil {
		log.Fatalf("Failed to sub-filesystem: %v", err)
	}

	fileServer := http.FileServer(http.FS(staticFS))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			fileServer.ServeHTTP(w, r)
			return
		}

		// Try opening exact file path
		f, err := staticFS.Open(strings.TrimPrefix(path, "/"))
		if err == nil {
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// Try appending .html (Next.js export route matching)
		htmlPath := strings.TrimPrefix(path, "/") + ".html"
		f, err = staticFS.Open(htmlPath)
		if err == nil {
			f.Close()
			r.URL.Path = "/" + htmlPath
			fileServer.ServeHTTP(w, r)
			return
		}

		// Fallback to index.html for SPA router
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Bifrost AI Gateway starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, LoggingMiddleware(mux)); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(lrw, r)
		log.Printf("[%s] %s %d - %v", r.Method, r.URL.Path, lrw.statusCode, time.Since(start))
	})
}
