package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	obshttp "egt.run/observer/http"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

func main() {
	env := flag.String("env", "dev", "environment (dev, prod)")
	logs := flag.String("log", "logs", "log directory")
	flag.Parse()

	log := zerolog.New(os.Stdout).With().Timestamp().Logger()
	if *env == "dev" {
		log = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	}
	if err := run(log, *env, *logs); err != nil {
		log.Fatal().Err(err).Msg("failed run")
	}
}

func run(log zerolog.Logger, env, logDir string) error {
	var templateDir, assetDir string
	switch env {
	case "dev":
		templateDir = filepath.Join("assets", "templates")
		assetDir = "assets"
	case "prod":
		templateDir = filepath.Join("public", "templates")
		assetDir = "public"
	default:
		return fmt.Errorf("unknown env: %s", env)
	}
	srv, err := obshttp.NewServer(log, env, templateDir, assetDir, logDir)
	if err != nil {
		return errors.Wrap(err, "new server")
	}
	const port = 3000
	log.Info().Int("port", port).Msg("listening")
	if err = http.ListenAndServe(fmt.Sprintf(":%d", port), srv.Mux); err != nil {
		return errors.Wrap(err, "listen and serve")
	}
	return nil
}
