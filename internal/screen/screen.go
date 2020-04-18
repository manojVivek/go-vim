package screen

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gdamore/tcell"
)

var screen tcell.Screen

type cellData struct {
	X, Y int
	c    rune
}

// BlankScreen -  Show a blank file screen
func BlankScreen() {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, _ := cmd.Output()
	dim := strings.Split(string(out), " ")
	heightI64, _ := strconv.ParseInt(dim[0], 10, 0)
	height := int(heightI64)
	widthI64, _ := strconv.ParseInt(strings.TrimSpace(dim[1]), 10, 0)
	width := int(widthI64)
	fmt.Printf("out: %v %v \n", height, width)
	text := make([]cellData, height)
	for i := 0; i < height; i++ {
		text[i] = cellData{0, i, '~'}
	}
	print(text)
}

// Init - Initilizes the screen
func Init() {
	s, err := tcell.NewScreen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	screen = s
	if e := s.Init(); e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
	screen.SetStyle(tcell.StyleDefault.
		Foreground(tcell.ColorBlack).
		Background(tcell.ColorWhite))
	screen.Show()
}

// Close - Closes the screen
func Close() {
	screen.Fini()
}

func print(text []cellData) {
	for _, d := range text {
		screen.SetContent(d.X, d.Y, d.c, nil, tcell.StyleDefault)
	}
	screen.Sync()
}
