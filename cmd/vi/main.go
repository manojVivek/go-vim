package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/manojVivek/go-vim/internal/actions"
	"github.com/manojVivek/go-vim/internal/screen"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\n    go-vim [options] filename\n\n")

		flag.PrintDefaults()
		os.Exit(2)
	}
	flag.Parse()
	if len(flag.Args()) == 0 {
		flag.Usage()
	}
	screen.Init(flag.Args()[0])
	c := actions.EventStream(screen.Screen)
	screen.HandleUserActions(c)
	screen.Close()
}
