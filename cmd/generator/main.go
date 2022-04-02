package main

import (
	"github.com/cnblvr/puzzles/app"
	"github.com/cnblvr/puzzles/internal/generator"
	"github.com/rs/zerolog/log"
)

func main() {
	app.InitHumanLogger()
	srv, err := generator.NewService()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create generator service")
	}

	log.Info().Msgf("service started...")
	if err := srv.Run(); err != nil {
		log.Fatal().Err(err).Msg("failed to run service")
	}
}
