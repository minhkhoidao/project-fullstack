package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/kyle/product/internal/platform/config"
	"github.com/kyle/product/internal/platform/logger"
	pkghttp "github.com/kyle/product/pkg/httputil"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load("gateway", ":8080")
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log := logger.New(cfg.ServiceName, cfg.LogLevel)

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:8080"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(pkghttp.RequestLogger(log))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		pkghttp.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	services := map[string]string{
		"user":         getEnv("USER_SERVICE_URL", "http://localhost:8081"),
		"product":      getEnv("PRODUCT_SERVICE_URL", "http://localhost:8082"),
		"cart":         getEnv("CART_SERVICE_URL", "http://localhost:8083"),
		"order":        getEnv("ORDER_SERVICE_URL", "http://localhost:8084"),
		"payment":      getEnv("PAYMENT_SERVICE_URL", "http://localhost:8085"),
		"inventory":    getEnv("INVENTORY_SERVICE_URL", "http://localhost:8086"),
		"review":       getEnv("REVIEW_SERVICE_URL", "http://localhost:8087"),
		"notification": getEnv("NOTIFICATION_SERVICE_URL", "http://localhost:8088"),
		"admin":        getEnv("ADMIN_SERVICE_URL", "http://localhost:8089"),
	}

	r.Route("/api/v1", func(r chi.Router) {
		r.HandleFunc("/auth/*", proxy(services["user"]))
		r.HandleFunc("/users/*", proxy(services["user"]))
		r.HandleFunc("/products/*", proxy(services["product"]))
		r.HandleFunc("/products", proxy(services["product"]))
		r.HandleFunc("/categories/*", proxy(services["product"]))
		r.HandleFunc("/categories", proxy(services["product"]))
		r.HandleFunc("/cart/*", proxy(services["cart"]))
		r.HandleFunc("/cart", proxy(services["cart"]))
		r.HandleFunc("/orders/*", proxy(services["order"]))
		r.HandleFunc("/orders", proxy(services["order"]))
		r.HandleFunc("/payments/*", proxy(services["payment"]))
		r.HandleFunc("/reviews/*", proxy(services["review"]))
		r.HandleFunc("/inventory/*", proxy(services["inventory"]))
		r.HandleFunc("/admin/*", proxy(services["admin"]))
	})

	log.Info("gateway starting", "addr", cfg.HTTPAddr)
	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: r,
	}
	return srv.ListenAndServe()
}

func proxy(target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, err := url.Parse(target)
		if err != nil {
			http.Error(w, "bad gateway", http.StatusBadGateway)
			return
		}
		p := httputil.NewSingleHostReverseProxy(u)
		p.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("proxy error to %s: %v", target, err)
			pkghttp.Error(w, http.StatusBadGateway, "service unavailable")
		}
		p.ServeHTTP(w, r)
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
