package main

import (
	"log"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/timothydzokoto/grpc_graphql_microservice/account"
	"github.com/tinrab/retry"
)

type Config struct {
	DatabaseUrl string `envconfig:"DATABASE_URL"`
}

func main() {
	var cfg Config

	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal(err)
	}

	// Connection to database
	var r account.Repository
	retry.ForeverSleep(2*time.Second, func(_ int) (err error) {
		r, err = account.NewPostgresRepository(cfg.DatabaseUrl)
		if err != nil {
			log.Println(err)
		}
		return
	})
	defer r.Close()
	log.Println("Listening on port 8080")

	// Service
	s := account.NewService(r)
	log.Fatal(account.ListenGRPC(s, 8080))
}
