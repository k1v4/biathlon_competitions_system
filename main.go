package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("run app by: go run main.go <config.json> <events.txt>")
	}

	fmt.Println(os.Args)
}
