package main

import (
	"log"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/timothydzokoto/grpc_graphql_microservice/order"
	"github.com/tinrab/retry"
)

type Config struct {
	DatbaseURL string `envconfig:"DATABASE_URL"`
	AccountUrl string `envconfig:"ACCOUNT_SERVICE_URL"`
	CatalogUrl string `envconfig:"CATALOG_SERVICE_URL"`
}

func main() {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal(err)
	}

	var r order.Repository
	retry.ForeverSleep(2*time.Second, func(_ int) (err error) {
		r, err = order.NewPostgresRepository(cfg.DatbaseURL)
		if err != nil {
			log.Println(err)
		}
		return
	})

	defer r.Close()
	log.Println("Listening on port 8080....")

	s := order.NewService(r)
	log.Fatal(order.ListenGRPC(s, cfg.AccountUrl, cfg.CatalogUrl, 8080))

}
