package server

import (
	"log"
	"net/http"

	db "github.com/fullcycle-rate-limiter/pkg/db/redis"
	middleware "github.com/fullcycle-rate-limiter/pkg/http/middleware"
	"github.com/go-chi/chi/v5"
)

func NewWebServer(maxRequestsWithoutToken, maxTokenRequests, blockDuration int, redis db.Cache) *chi.Mux {

	log.Printf("Rate Limiter Configuration: maxRequestsWithoutToken=%d, maxTokenRequests=%d, blockDuration=%d", maxRequestsWithoutToken, maxTokenRequests, blockDuration)
	rateLimiter := middleware.NewRateLimiter(maxRequestsWithoutToken, maxTokenRequests, blockDuration, redis)

	router := chi.NewRouter()
	router.Use(rateLimiter.Limit)
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Service is running"))
	})

	return router
}
