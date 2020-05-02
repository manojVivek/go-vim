package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/manojVivek/go-vim/internal/actions"
	"github.com/manojVivek/go-vim/internal/editor"
	"github.com/manojVivek/go-vim/internal/logger"
)

func main() {
	logger.InitLogger()
	logger.Debug.Printf("Starting Editor")
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
	e, err := editor.NewEditor(flag.Args()[0])
	if err != nil {
		logger.Debug.Printf("Error %v", err)
		os.Exit(2)
	}
	c := actions.EventStream(e.GetTerminalScreen())
	e.HandleUserActions(c)
}
