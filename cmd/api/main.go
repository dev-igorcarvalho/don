package main

import (
	"log"

	"github.com/dev-igorcarvalho/don/internal/config"
	"github.com/dev-igorcarvalho/don/pkg/logger"
)

func main() {
	cfg := must(config.Load[config.AppConfig]())
	logger.Setup(logger.Environment(cfg.Environment))
	//todo complete the main with everything need + gracefull shutdown

}

func must[T any](v *T, err error) *T {
	if err != nil {
		log.Fatalf("load failed: %v", err)
	}
	return v
}

func run() error {
	cfg := must(config.Load[config.AppConfig]())
	logger.Setup(logger.Environment(cfg.Environment))
	return nil
}

func wireServer(cfg config.AppConfig)
