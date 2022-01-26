package main

import (
	"fmt"
	"log"
	"os"

	"github.com/benoitkugler/binarygen"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("missing input file")
	}
	input := os.Args[1]
	err := binarygen.Generate(input)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Done.")
}
