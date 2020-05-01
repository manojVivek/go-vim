package screen

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell"
	"github.com/manojVivek/go-vim/internal/actions"
	"github.com/manojVivek/go-vim/internal/fs"
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

type cursorMovement string

const (
	CURSOR_UP    cursorMovement = "UP"
	CURSOR_DOWN  cursorMovement = "DOWN"
	CURSOR_RIGHT cursorMovement = "RIGHT"
	CURSOR_LEFT  cursorMovement = "LEFT"
)

var currentMode = MODE_NORMAL
var statusMessage string
var fileName string
var dataBuffer []string
var textFrame [][]rune
var startLine int = 0
var cursorPosBuffer Vertex
var cursorPosScreen Vertex
var currentCommand string
var lastUpDownCursorMovement cursorMovement = CURSOR_UP
var firstLineInFrame int = 0
var lastLineInFrame int

// InitBuffer - Intialize the textFrame with filecontent / blank file content
func InitBuffer(f string) {
	// TODO: Load only the required portion in memory instead of whole file
	fileName = f
	var err error
	dataBuffer, err = fs.ReadFileToLines(f)
	fmt.Printf("filecontent %v %v", dataBuffer, err)
	statusMessage = fmt.Sprintf("\"%v\"", f)
	if err != nil {
		dataBuffer = make([]string, 1)
		statusMessage += " [New File]"
	} else {
		charCount := 0
		for _, line := range dataBuffer {
			charCount += len(line) + 1
		}
		statusMessage += fmt.Sprintf(" %vL, %vC", len(dataBuffer), charCount)
	}

	cursorPosBuffer = Vertex{0, 0}

	syncTextFrame(false)
	syncCursor()
	displayStatusBar()
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
				statusMessage = ""
				displayStatusBar()
				fixHorizontalCursorOverflow()
				syncCursor()
				continue
			}
			if handleTextAreaCursorMovement(e) {
				syncCursor()
				continue
			}
			handleKeyInsertMode(e)
		case MODE_NORMAL:
			if handleTextAreaCursorMovement(e) {
				syncCursor()
				continue
			}
			switch e.Rune {
			case ':':
				currentMode = MODE_COMMAND_LINE
				statusMessage = ""
				handleKeyCommandLineMode(e)
			case 'i':
				currentMode = MODE_INSERT
				statusMessage = "-- INSERT --"
				displayStatusBar()
			}
		case MODE_COMMAND_LINE:
			if e.Value == tcell.KeyEscape {
				currentMode = MODE_NORMAL
				currentCommand = ""
				displayStatusBar()
				continue
			}
			if e.Value == tcell.KeyEnter {
				runCommand(currentCommand)
			}
			handleKeyCommandLineMode(e)
		}

	}
}

func handleTextAreaCursorMovement(e actions.Event) bool {
	isProcessed := true
	rangeY := len(dataBuffer)
	switch e.Value {
	case tcell.KeyLeft:
		if cursorPosBuffer.X != 0 {
			cursorPosBuffer.X--
			syncCursor()
		}
	case tcell.KeyRight:
		cursorPosBuffer.X++
		fixHorizontalCursorOverflow()
		syncCursor()
	case tcell.KeyDown:
		if cursorPosBuffer.Y+1 != rangeY {
			cursorPosBuffer.Y++
			lastUpDownCursorMovement = CURSOR_DOWN
			fixHorizontalCursorOverflow()
			fixVerticalCursorOverflow()
			syncCursor()
		}
	case tcell.KeyUp:
		if cursorPosBuffer.Y != 0 {
			cursorPosBuffer.Y--
			lastUpDownCursorMovement = CURSOR_UP
			fixHorizontalCursorOverflow()
			if cursorPosBuffer.Y < firstLineInFrame {
				firstLineInFrame--
				if lastLineInFrame-firstLineInFrame > screenDim.Y-1 {
					lastLineInFrame--
				}
			}
			syncCursor()
		}
	default:
		isProcessed = false
	}
	return isProcessed
}

func fixVerticalCursorOverflow() {
	if cursorPosBuffer.Y > lastLineInFrame {
		if lastLineInFrame+1 >= len(dataBuffer) {
			return
		}
		lastLineInFrame++
		if lastLineInFrame-firstLineInFrame > screenDim.Y-2 {
			firstLineInFrame++
		}
		for i := lastLineInFrame; i >= firstLineInFrame; i-- {
			firstLineInFrame += len(dataBuffer[i]) / screenDim.X
		}
	}
}

func fixHorizontalCursorOverflow() {
	rangeX := len(dataBuffer[cursorPosBuffer.Y])
	if currentMode != MODE_INSERT {
		rangeX--
		if rangeX < 0 {
			rangeX = 0
		}
	}
	if rangeX < cursorPosBuffer.X {
		cursorPosBuffer.X = rangeX
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
	displayStatusBar()
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
		lastUpDownCursorMovement = CURSOR_DOWN
		dataBuffer = newBuffer
		fixVerticalCursorOverflow()
		syncTextFrame(false)
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if cursorPosBuffer.X == 0 && cursorPosBuffer.Y == 0 {
			return
		}
		if cursorPosBuffer.X > 0 {
			// Delete a character in a line
			line := dataBuffer[cursorPosBuffer.Y]
			dataBuffer[cursorPosBuffer.Y] = line[:cursorPosBuffer.X-1] + line[cursorPosBuffer.X:]
			cursorPosBuffer.X--
		} else {
			// Merge the contents of this line to previous line
			newBuffer := make([]string, len(dataBuffer)-1)
			j := 0
			for i, line := range dataBuffer {
				if i == cursorPosBuffer.Y {
					continue
				}
				if i == cursorPosBuffer.Y-1 {
					line = line + dataBuffer[cursorPosBuffer.Y]
				}
				newBuffer[j] = line
				j++
			}
			cursorPosBuffer.X = len(dataBuffer[cursorPosBuffer.Y-1])
			cursorPosBuffer.Y--
			dataBuffer = newBuffer
		}
		syncTextFrame(false)
	default:
		if e.Rune == 0 {
			return
		}
		line := dataBuffer[cursorPosBuffer.Y]
		dataBuffer[cursorPosBuffer.Y] = line[:cursorPosBuffer.X] + string(e.Rune) + line[cursorPosBuffer.X:]
		cursorPosBuffer.X++
		syncTextFrame(false)
	}
}

func syncTextFrame(dontSyncCursor bool) {

	textFrameTemp := fillTextFrameFromTop()

	textFrameTemp = fillEmptyLinesIfAnyWithTilde(textFrameTemp)

	textFrame = textFrameTemp
	displayTextFrame()
	if dontSyncCursor == false {
		syncCursor()
	}
}

func fillEmptyLinesIfAnyWithTilde(textFrameTemp [][]rune) [][]rune {
	//Fill the empty lines in the frame with '~' char
	shouldTilde := true
	for i := len(textFrameTemp) - 1; i > 0; i-- {
		if textFrameTemp[i] != nil {
			shouldTilde = false
			continue
		}
		textFrameTemp[i] = make([]rune, screenDim.X)
		if shouldTilde {
			textFrameTemp[i][0] = '~'
		}

	}
	return textFrameTemp
}

func fillTextFrameFromTop() [][]rune {
	x := 0
	y := 0
	textFrameTemp := make([][]rune, screenDim.Y-1)
	i := firstLineInFrame
	for {
		x = 0
		line := dataBuffer[i]
		for _, char := range line {
			if textFrameTemp[y] == nil {
				textFrameTemp[y] = make([]rune, screenDim.X)
			}
			textFrameTemp[y][x] = char
			if x+1 == screenDim.X {
				x = 0
				y++
			} else {
				x++
			}
		}

		if i+1 == len(dataBuffer) || y == screenDim.Y-2 {
			break
		}
		i++
		y++
	}
	lastLineInFrame = i
	return textFrameTemp
}

func syncCursor() {
	cursorPosScreen = Vertex{cursorPosBuffer.X, cursorPosBuffer.Y}
	cursorPosScreen.Y -= firstLineInFrame
	for i := firstLineInFrame; i < cursorPosBuffer.Y; i++ {
		cursorPosScreen.Y += (len(dataBuffer[i]) / screenDim.X)
	}
	for cursorPosScreen.X > screenDim.X {
		cursorPosScreen.X -= screenDim.X
		cursorPosScreen.Y++
	}

	syncTextFrame(true)
	displayCursor()
}
