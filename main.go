package main

import (
	"log"

	"github.com/theleeeo/thor/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Println(err)
	}
}
