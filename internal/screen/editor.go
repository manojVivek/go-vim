package screen

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell"
	"github.com/manojVivek/vim_go/internal/actions"
)

var textFrame [][]rune
var lineCount int = 0
var cursorPos Vertex

type Vertex struct {
	X, Y int
}

func InitBuffer() {
	x := screenDim.X
	y := screenDim.Y
	fmt.Println("ScreenDim: %v", screenDim)

	textFrame = make([][]rune, y)
	for i := range textFrame {
		textFrame[i] = make([]rune, x)
	}

	fmt.Println("TextFrame: %v", x)

	// TODO: Fill the text from buffer

	//Fill the empty lines in the frame with '~' char
	if lineCount < y+1 {
		for i := lineCount; i < y; i++ {
			textFrame[i][0] = '~'
		}
	}
	cursorPos = Vertex{0, 0}
	syncTextFrame()
}

// HandleUserActions - function that listens the user input channel and handles it appropriately
func HandleUserActions(c chan actions.Event) {
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
		textFrame[cursorPos.Y][cursorPos.X] = e.Rune
		cursorPos.X++
		if cursorPos.X == screenDim.X {
			cursorPos.X = 0
			cursorPos.Y++
		}
		syncTextFrame()
	}
}
