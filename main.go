package main

import (
	"flag"
	"log"
	"os"
	"strings"
)

var (
	filename = flag.String("file", "", "filename")
	line     = flag.Int("line", 0, "line number of the struct literal")
	version  = flag.String("version", "", "print fillstruct version")
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("fillstruct: ")

	if strings.Contains(strings.Join(os.Args, " "), "-version") {
		log.Println(_version)
		os.Exit(0)
	}

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
