package screen

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell"
	"github.com/manojVivek/vim_go/internal/actions"
)

var dataBuffer []string = make([]string, 1)
var textFrame [][]rune
var startLine int = 0
var cursorPosBuffer Vertex
var cursorPosScreen Vertex

type Vertex struct {
	X, Y int
}

// InitBuffer - Intialize the textFrame with filecontent / blank file content
func InitBuffer() {
	x := screenDim.X
	y := screenDim.Y

	textFrame = make([][]rune, y)
	for i := range textFrame {
		textFrame[i] = make([]rune, x)
	}

	// TODO: Fill the text from buffer

	cursorPosBuffer = Vertex{0, 0}

	syncTextFrame()
	syncCursor()
}

// HandleUserActions - function that listens the user input channel and handles it appropriately
func HandleUserActions(c chan actions.Event) {
	escapePressed := false
	for e := range c {
		if e.Kind != "KEY_PRESS" {
			continue
		}
		switch e.Value {
		case tcell.KeyEscape:
			escapePressed = true
		case tcell.KeyLeft:
			if cursorPosBuffer.X != 1 {
				cursorPosBuffer.X--
				syncCursor()
			}
		case tcell.KeyRight:
			if cursorPosBuffer.X != len(dataBuffer[cursorPosBuffer.Y]) {
				cursorPosBuffer.X++
				syncCursor()
			}
		default:
			if escapePressed {
				if e.Rune == 'q' {
					Close()
					os.Exit(0)
				} else {
					escapePressed = false
				}
			}
			if e.Rune == 0 {
				continue
			}
			line := dataBuffer[cursorPosBuffer.Y]
			dataBuffer[cursorPosBuffer.Y] = line[:cursorPosBuffer.X] + string(e.Rune) + line[cursorPosBuffer.X:]
			if e.Rune == 'x' {
				Close()
				fmt.Printf("Data: %v", dataBuffer)
			}
			cursorPosBuffer.X++
			syncTextFrame()
		}
	}
}

func syncTextFrame() {
	x := 0
	y := 0
	textFrame = make([][]rune, screenDim.Y)
	for i := range textFrame {
		textFrame[i] = make([]rune, screenDim.X)
	}
	for _, line := range dataBuffer {
		x = 0
		for _, char := range line {
			//fmt.Printf("char %v %v %v \n", x, y, char)
			textFrame[y][x] = char
			if x+1 == screenDim.X {
				x = 0
				y++
			} else {
				x++
			}
		}
		y++
	}
	//Fill the empty lines in the frame with '~' char
	if y < screenDim.Y {
		for ; y < screenDim.Y; y++ {
			textFrame[y][0] = '~'
		}
	}

	//fmt.Printf("Values: %v", textFrame[0])
	//os.Exit(0)

	displayTextFrame()
	syncCursor()
}

func syncCursor() {
	cursorPosScreen = Vertex{cursorPosBuffer.X, cursorPosBuffer.Y}
	for cursorPosScreen.X > screenDim.X {
		cursorPosScreen.X -= screenDim.X
		cursorPosScreen.Y++
	}

	displayCursor()
}
