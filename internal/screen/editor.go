package screen

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell"
	"github.com/manojVivek/vim_go/internal/actions"
	"github.com/manojVivek/vim_go/internal/fs"
)

type Vertex struct {
	X, Y int
}

type mode string

const (
	MODE_NORMAL       mode = "NORMAL"
	MODE_INSERT       mode = "INSERT"
	MODE_COMMAND_LINE mode = "COMMAND_LINE"
)

var currentMode = MODE_NORMAL
var fileName string
var dataBuffer []string
var textFrame [][]rune
var startLine int = 0
var cursorPosBuffer Vertex
var cursorPosScreen Vertex
var currentCommand string

// InitBuffer - Intialize the textFrame with filecontent / blank file content
func InitBuffer(f string) {
	// TODO: Load only the required portion in memory instead of whole file
	fileName = f
	var err error
	dataBuffer, err = fs.ReadFileToLines(f)
	fmt.Printf("filecontent %v %v", dataBuffer, err)
	if err != nil {
		dataBuffer = make([]string, 1)
	}

	cursorPosBuffer = Vertex{0, 0}

	syncTextFrame()
	syncCursor()
}

// HandleUserActions - function that listens the user input channel and handles it appropriately
func HandleUserActions(c chan actions.Event) {
	for e := range c {
		if e.Kind != "KEY_PRESS" {
			continue
		}
		switch currentMode {
		case MODE_INSERT:
			if e.Value == tcell.KeyEscape {
				currentMode = MODE_NORMAL
				continue
			}
			handleKeyInsertMode(e)
		case MODE_NORMAL:
			switch e.Rune {
			case ':':
				currentMode = MODE_COMMAND_LINE
				handleKeyCommandLineMode(e)
			case 'i':
				currentMode = MODE_INSERT
			}
		case MODE_COMMAND_LINE:
			if e.Value == tcell.KeyEscape {
				currentMode = MODE_NORMAL
				currentCommand = ""
				displayCommand()
				continue
			}
			if e.Value == tcell.KeyEnter {
				runCommand(currentCommand)
			}
			handleKeyCommandLineMode(e)
		}

	}
}

func runCommand(cmd string) {
	switch cmd {
	case ":q":
		Close()
		os.Exit(0)
	case ":wq":
		fs.WriteLinesToFile(fileName, dataBuffer)
		Close()
		os.Exit(0)
	default:
		currentCommand = ""
	}
}

func handleKeyCommandLineMode(e actions.Event) {
	if e.Rune == 0 {
		return
	}
	if e.Value == tcell.KeyBackspace || e.Value == tcell.KeyBackspace2 {
		currentCommand = currentCommand[:len(currentCommand)-1]
	} else {
		currentCommand += string(e.Rune)
	}
	displayCommand()
}

func handleKeyInsertMode(e actions.Event) {
	switch e.Value {
	case tcell.KeyEnter:
		newBuffer := make([]string, len(dataBuffer)+1)
		copy(newBuffer, dataBuffer)
		if cursorPosBuffer.Y < len(dataBuffer)-1 {
			for i := len(newBuffer) - 1; i > cursorPosBuffer.Y; i-- {
				newBuffer[i] = newBuffer[i-1]
			}
		}

		if cursorPosBuffer.X < len(dataBuffer[cursorPosBuffer.Y]) {
			newBuffer[cursorPosBuffer.Y] = dataBuffer[cursorPosBuffer.Y][:cursorPosBuffer.X]
			newBuffer[cursorPosBuffer.Y+1] = dataBuffer[cursorPosBuffer.Y][cursorPosBuffer.X:]
		}
		cursorPosBuffer.Y++
		cursorPosBuffer.X = 0
		dataBuffer = newBuffer
		syncTextFrame()
	case tcell.KeyLeft:
		if cursorPosBuffer.X != 0 {
			cursorPosBuffer.X--
			syncCursor()
		}
	case tcell.KeyRight:
		if cursorPosBuffer.X != len(dataBuffer[cursorPosBuffer.Y]) {
			cursorPosBuffer.X++
			syncCursor()
		}
	case tcell.KeyDown:
		if cursorPosBuffer.Y+1 != len(dataBuffer) {
			cursorPosBuffer.Y++
			if len(dataBuffer[cursorPosBuffer.Y]) < cursorPosBuffer.X {
				cursorPosBuffer.X = len(dataBuffer[cursorPosBuffer.Y])
			}
			syncCursor()
		}
	case tcell.KeyUp:
		if cursorPosBuffer.Y != 0 {
			cursorPosBuffer.Y--
			if len(dataBuffer[cursorPosBuffer.Y]) < cursorPosBuffer.X {
				cursorPosBuffer.X = len(dataBuffer[cursorPosBuffer.Y])
			}
			syncCursor()
		}
	default:
		if e.Rune == 0 {
			return
		}
		line := dataBuffer[cursorPosBuffer.Y]
		dataBuffer[cursorPosBuffer.Y] = line[:cursorPosBuffer.X] + string(e.Rune) + line[cursorPosBuffer.X:]
		if false && e.Rune == 'x' {
			Close()
			fmt.Printf("Data: %v", dataBuffer)
		}
		cursorPosBuffer.X++
		syncTextFrame()
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
		if y == screenDim.Y {
			break
		}
	}
	//Fill the empty lines in the frame with '~' char
	for ; y < screenDim.Y-1; y++ {
		textFrame[y][0] = '~'
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
