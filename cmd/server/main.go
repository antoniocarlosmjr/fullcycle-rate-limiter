package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/fullcycle-rate-limiter/config"
	db "github.com/fullcycle-rate-limiter/pkg/db/redis"
	"github.com/fullcycle-rate-limiter/pkg/http/server"
)

func main() {
	configs, err := config.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	log.Println("Creating redis ...")

	redis, err := db.NewRedis(fmt.Sprintf("%s:%s", configs.RedisHost, configs.RedisPort))
	if err != nil {
		panic(err)
	}

	r := server.NewWebServer(
		configs.MaxRequestsWithoutToken,
		configs.MaxRequestsWithToken,
		configs.TimeBlockInSecond,
		redis,
	)

	log.Println("Starting web server on port", "8080")
	err = http.ListenAndServe(fmt.Sprintf(":%s", "8080"), r)
	if err != nil {
		panic(err)
	}
}
