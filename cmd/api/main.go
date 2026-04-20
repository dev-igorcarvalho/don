package main

import (
	"github.com/dev-igorcarvalho/don/internal/config"
	pkgConfig "github.com/dev-igorcarvalho/don/pkg/config"
	"github.com/dev-igorcarvalho/don/pkg/logger"
	"github.com/dev-igorcarvalho/don/pkg/must"
)

func main() {
	//todo complete the main with everything need + gracefull shutdown
}

func run() error {
	cfg := must.Get(pkgConfig.Load[config.AppConfig]())
	logger.Setup(logger.Environment(cfg.Environment))
	return nil
}

func wireServer(cfg config.AppConfig) {}
