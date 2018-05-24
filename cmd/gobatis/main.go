package main

import (
	"flag"
	"log"

	"github.com/runner-mei/GoBatis/generator"
)

func main() {
	var gen = generator.Generator{}
	gen.Flags(flag.CommandLine)
	flag.Parse()

	if err := gen.Run(flag.Args()); err != nil {
		log.Println(err)
	}
}
