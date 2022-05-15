package main

import (
	"flag"
	"log"

	"github.com/runner-mei/GoBatis/cmd/gobatis/generator"
)

func main() {
	gen := generator.Generator{}
	gen.Flags(flag.CommandLine)
	flag.Parse()

	if err := gen.Run(flag.Args()); err != nil {
		log.Println(err)
	}
}
