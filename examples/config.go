package main

import (
	"fmt"
	"log"

	"github.com/gerfey/messenger/config"
)

func main() {
	cfg, err := config.LoadConfig("examples/messenger.yaml")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Default bus:", cfg.DefaultBus)
}
