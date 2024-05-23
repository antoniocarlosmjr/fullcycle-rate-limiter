package middleware

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	db "github.com/fullcycle-rate-limiter/pkg/db/redis"
)

type RateLimiter struct {
	maxIPRequests    int
	maxTokenRequests int
	blockDuration    int
	cache            db.Cache
}

func extractClientIP(r *http.Request) (string, error) {
	forwardedIPs := r.Header.Get("X-Forwarded-For")
	ipList := strings.Split(forwardedIPs, ",")

	if len(ipList) > 0 {
		clientIP := net.ParseIP(strings.TrimSpace(ipList[len(ipList)-1]))
		if clientIP != nil {
			return clientIP.String(), nil
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}

	clientIP := net.ParseIP(ip)
	if clientIP != nil {
		if ip == "::1" {
			return "127.0.0.1", nil
		}
		return clientIP.String(), nil
	}

	return "", errors.New("client IP not found")
}

func NewRateLimiter(maxIPRequests, maxTokenRequests, blockDuration int, cache db.Cache) *RateLimiter {
	return &RateLimiter{
		maxIPRequests:    maxIPRequests,
		maxTokenRequests: maxTokenRequests,
		blockDuration:    blockDuration,
		cache:            cache,
	}
}

func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientID, err := extractClientIP(r)
		if err != nil {
			log.Printf("Error extracting client IP: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		requestType := "ip"
		if apiKey := r.Header.Get("API_KEY"); apiKey != "" {
			clientID = apiKey
			requestType = "api_key"
		}

		cacheKey := fmt.Sprintf("%s:%s", requestType, clientID)
		log.Printf("Cache key: %s", cacheKey)

		cacheValue, err := rl.cache.Get(cacheKey)
		if err != nil {
			log.Printf("Error fetching from cache: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if cacheValue == "" {
			if err := rl.cache.Set(cacheKey, "1", time.Duration(rl.blockDuration)*time.Second); err != nil {
				log.Printf("Error setting cache: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		requestCount, err := strconv.Atoi(cacheValue)
		if err != nil {
			log.Printf("Error converting cache value: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		maxRequests := rl.maxIPRequests
		if requestType == "api_key" {
			maxRequests = rl.maxTokenRequests
		}

		if requestCount+1 > maxRequests {
			log.Println("Rate limit exceeded")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("you have reached the maximum number of requests or actions allowed within a certain time frame"))
			return
		}

		if err := rl.cache.Set(cacheKey, strconv.Itoa(requestCount+1), time.Duration(rl.blockDuration)*time.Second); err != nil {
			log.Printf("Error updating cache: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		next.ServeHTTP(w, r)
	})
}
