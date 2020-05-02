package screen

import (
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gdamore/tcell"
)

// Dimension is a struct that has X and Y field to represent a rectangle
type Dimension struct {
	X, Y int
}

type cellData struct {
	X, Y int
	c    rune
}

// Screen is a struct that has the logic to display the editor state on the terminal
type Screen struct {
	tScreen   tcell.Screen
	screenDim Dimension
}

func (s *Screen) updateScreenDimensions() {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, _ := cmd.Output()
	dim := strings.Split(string(out), " ")
	heightI64, _ := strconv.ParseInt(dim[0], 10, 0)
	height := int(heightI64)
	widthI64, _ := strconv.ParseInt(strings.TrimSpace(dim[1]), 10, 0)
	width := int(widthI64)
	s.screenDim = Dimension{width, height}
}

// Init - Initilizes the Screen
func NewScreen() (*Screen, error) {
	s, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if e := s.Init(); e != nil {
		return nil, err
	}
	s.SetStyle(tcell.StyleDefault.
		Foreground(tcell.ColorBlack).
		Background(tcell.ColorWhite))
	//Screen.Show()
	screen := Screen{tScreen: s}
	screen.updateScreenDimensions()
	return &screen, nil
}

// Close - Closes the Screen
func (s *Screen) Close() {
	s.tScreen.Fini()
}

func (s *Screen) DisplayTextFrame() {
	for y := range textFrame {
		for x := range textFrame[y] {
			s.tScreen.SetContent(x, y, textFrame[y][x], nil, tcell.StyleDefault)
		}
	}
	s.tScreen.Show()
}

func (s *Screen) DisplayStatusBar() {
	for x := 0; x < s.screenDim.X; x++ {
		s.tScreen.SetContent(x, s.screenDim.Y-1, ' ', nil, tcell.StyleDefault)
	}
	if len(statusMessage) > 0 {
		for x, c := range statusMessage {
			s.tScreen.SetContent(x, s.screenDim.Y-1, c, nil, tcell.StyleDefault)
		}
	} else {
		for x, c := range currentCommand {
			s.tScreen.SetContent(x, s.screenDim.Y-1, c, nil, tcell.StyleDefault)
		}
	}
	s.tScreen.Show()
}

func (s *Screen) DisplayCursor() {
	s.tScreen.ShowCursor(cursorPosScreen.X, cursorPosScreen.Y)
	s.tScreen.Show()
}
