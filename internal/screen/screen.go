package Screen

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/manojVivek/vim_go/internal/actions"
)

var Screen tcell.Screen

type cellData struct {
	X, Y int
	c    rune
}

// BlankScreen -  Show a blank file Screen
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

// Init - Initilizes the Screen
func Init() {
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
	Screen.Show()
}

// Close - Closes the Screen
func Close() {
	Screen.Fini()
}

// PollEvent - Waits for an event
func PollEvent() {
	ev := Screen.PollEvent()
	switch ev := ev.(type) {
	case *tcell.EventKey:
		fmt.Printf("Event %v", ev.Key())
	default:
		fmt.Printf("Event %v", ev.When())
	}
}

// PrintBuffer - function that listens the channel an prints the character as they are typed
func PrintBuffer(c chan actions.Event) {
	i := 0
	escapePressed := false
	for e := range c {
		if e.Kind != "KEY_PRESS" {
			continue
		}
		if e.Value == tcell.KeyEscape {
			escapePressed = true
			continue
		}
		if escapePressed {
			if e.Rune == 'q' {
				Close()
				os.Exit(0)
			} else {
				escapePressed = false
			}
		}

		Screen.SetContent(i, 0, e.Rune, nil, tcell.StyleDefault)
		i = i + 1
		Screen.Sync()
	}
}

func print(text []cellData) {
	for _, d := range text {
		Screen.SetContent(d.X, d.Y, d.c, nil, tcell.StyleDefault)
	}
	Screen.ShowCursor(0, 0)
	Screen.Sync()
}
