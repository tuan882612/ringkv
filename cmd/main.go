package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

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
		log.Info().Msg("No address provided. Running set of nodes...")
		n1 := chord.NewNode("localhost:3000")
		time.Sleep(1 * time.Second)
		n2 := chord.NewNode("localhost:3001")
		time.Sleep(1 * time.Second)
		n3 := chord.NewNode("localhost:3002")
		time.Sleep(1 * time.Second)
		n4 := chord.NewNode("localhost:3003")
		time.Sleep(1 * time.Second)
		n5 := chord.NewNode("localhost:3004")
		n1.Bootstrap()
		time.Sleep(1 * time.Second)
		n2.JoinRing("localhost:3000")
		time.Sleep(1 * time.Second)
		n3.JoinRing("localhost:3001")
		time.Sleep(1 * time.Second)
		n4.JoinRing("localhost:3000")
		time.Sleep(1 * time.Second)
		n5.JoinRing("localhost:3000")
	} else {
		node := chord.NewNode(newAddr)

		if joinAddr == "" {
			log.Info().Msg("No join address provided. Starting as a bootstrap node...")
			node.Bootstrap()
		} else {
			log.Info().Msgf("Joining existing node at %s", joinAddr)
			node.JoinRing(joinAddr)
		}
	}

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT)
	<-sigint
	log.Info().Msg("Received SIGINT. Shutting down...")
}
