package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"ringkv/pkg/chord"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: zerolog.TimeFieldFormat})

	var (
		newAddr  string
		joinAddr string
	)

	flag.StringVar(&newAddr, "address", "", "Address of new node to bootstrap.")
	flag.StringVar(&joinAddr, "join", "", "Address of existing node to join.")
	flag.Parse()

	if newAddr == "" {
		log.Fatal().Msg("Address of new node is required.")
	}

	node := chord.NewNode(newAddr)

	if joinAddr == "" {
		log.Info().Msg("No join address provided. Starting as a bootstrap node...")
		node.Bootstrap()
	} else {
		log.Info().Msgf("Joining existing node at %s", joinAddr)
		node.JoinRing(joinAddr)
	}

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT)
	<-sigint
	log.Info().Msg("Received SIGINT. Shutting down...")
}
