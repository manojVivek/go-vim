package screen

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gdamore/tcell"
)

var Screen tcell.Screen
var screenDim Dimension

type Dimension struct {
	X, Y int
}

type cellData struct {
	X, Y int
	c    rune
}

// BlankScreen -  Show a blank file Screen
func BlankScreen() {

}

func updateScreenDimensions() {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, _ := cmd.Output()
	dim := strings.Split(string(out), " ")
	heightI64, _ := strconv.ParseInt(dim[0], 10, 0)
	height := int(heightI64)
	widthI64, _ := strconv.ParseInt(strings.TrimSpace(dim[1]), 10, 0)
	width := int(widthI64)
	screenDim = Dimension{width, height}
}

// Init - Initilizes the Screen
func Init(fileName string) {
	s, err := tcell.NewScreen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	Screen = s
	if e := s.Init(); e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
	Screen.SetStyle(tcell.StyleDefault.
		Foreground(tcell.ColorBlack).
		Background(tcell.ColorWhite))
	//Screen.Show()
	updateScreenDimensions()
	InitBuffer(fileName)
}

// Close - Closes the Screen
func Close() {
	Screen.Fini()
}

func displayTextFrame() {
	for y := range textFrame {
		for x := range textFrame[y] {
			Screen.SetContent(x, y, textFrame[y][x], nil, tcell.StyleDefault)
		}
	}
	Screen.Sync()
}

func displayStatusBar() {
	for x := 0; x < screenDim.Y; x++ {
		Screen.SetContent(x, screenDim.Y-1, ' ', nil, tcell.StyleDefault)
	}
	if len(statusMessage) > 0 {
		for x, c := range statusMessage {
			Screen.SetContent(x, screenDim.Y-1, c, nil, tcell.StyleDefault)
		}
	} else {
		for x, c := range currentCommand {
			Screen.SetContent(x, screenDim.Y-1, c, nil, tcell.StyleDefault)
		}
	}
	Screen.Sync()
}

func displayCursor() {
	Screen.ShowCursor(cursorPosScreen.X, cursorPosScreen.Y)
	Screen.Sync()
}
