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
		id, err := extractClientIP(r)
		if err != nil {
			log.Println("Failed to get remote address: ", err)
			http.Error(w, "Internal Server Error", http.StatusBadRequest)
			return
		}
		kind := "addr"
		if r.Header.Get("API_KEY") != "" {
			id = r.Header.Get("API_KEY")
			kind = "token"
		}

		key := fmt.Sprintf("%s:%s", kind, id)
		log.Println("Key: ", key)

		val, err := rl.cache.Get(key)
		if err != nil {
			log.Println("Failed to get value from cache: ", err)
			http.Error(w, "Internal Server Error", http.StatusBadGateway)
			return
		}

		if val == "" {
			val = "1"
			err := rl.cache.Set(key, val, time.Duration(rl.blockDuration)*time.Second)
			if err != nil {
				log.Println("Failed to set value in cache: ", err)
				http.Error(w, "Internal Server Error", http.StatusBadGateway)
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		count, err := strconv.Atoi(val)
		if err != nil {
			log.Println("Failed to convert value to int: ", err)
			http.Error(w, "Internal Server Error", http.StatusBadGateway)
			return
		}

		maxRequest := rl.maxIPRequests
		if kind == "token" {
			maxRequest = rl.maxTokenRequests
		}

		if count+1 > maxRequest {
			log.Println("Too many requests")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("you have reached the maximum number of requests or actions allowed within a certain time frame"))
			return
		}

		err = rl.cache.Set(key, strconv.Itoa(count+1), time.Duration(rl.blockDuration)*time.Second)
		if err != nil {
			log.Println("Filed to set value in cache: ", err)
			http.Error(w, "Internal Server Error", http.StatusBadGateway)
			return
		}

		next.ServeHTTP(w, r)
	})
}
