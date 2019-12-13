package main

import (
	"flag"
	"github.com/pmerzlyakov/gocha/chat"
	"log"
)

var (
	configFile = flag.String("config", "./config.json", "Path to config file")
)

func main() {
	flag.Parse()

	cfg, err := chat.LoadConfig(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	s, err := chat.NewServer(*cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(s.ListenAndServe())
}
