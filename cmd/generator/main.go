package main

import (
	"github.com/cnblvr/puzzles/internal/generator"
	"github.com/rs/zerolog/log"
)

func main() {
	srv, err := generator.NewService()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create generator service")
	}

	log.Info().Str("name", "generator").Msgf("service started...")
	if err := srv.Run(); err != nil {
		log.Fatal().Err(err).Msg("failed to run service")
	}
}
