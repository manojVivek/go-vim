package main

import (
	"fmt"
	"os"

	"github.com/manojVivek/vim_go/internal/screen"
)

func main() {
	fmt.Printf("hello, world \n")
	fmt.Printf("%s\n", os.Args[1])
	screen.Init()
	screen.BlankScreen()
	fmt.Scanln()
	screen.Close()
}
