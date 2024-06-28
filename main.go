package main

import (
	"errors"
	"flag"
	"log"
	"os"
)

var errNotFound = errors.New("no struct literal found at selection")

var File string

var (
	filename = flag.String("file", "", "filename")
	line     = flag.Int("line", 0, "line number of the struct literal")
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("fillstruct: ")

	flag.Parse()

	if *line == 0 || *filename == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	h := newHandler(*filename, *line)
	if err := h.preCheck(); err != nil {
		log.Fatal(err)
	}
	if err := h.travel(); err != nil {
		log.Fatal(err)
	}
	if err := h.writeBack(); err != nil {
		log.Fatal(err)
	}

}
