package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: zerolog.TimeFieldFormat})

	if len(os.Args) < 2 {
		log.Fatal().Msg("Error: missing listen address Usage: dht <listen-address>")
	}

	addr := os.Args[1]
	log.Info().Str("address", addr).Msg("Starting DHT node")
}
