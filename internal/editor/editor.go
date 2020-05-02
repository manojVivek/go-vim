package editor

import (
	"fmt"
	"os"

	"github.com/manojVivek/go-vim/internal/logger"

	"github.com/gdamore/tcell"
	"github.com/manojVivek/go-vim/internal/actions"
	"github.com/manojVivek/go-vim/internal/fs"
	terminal "github.com/manojVivek/go-vim/internal/screen"
)

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

type Editor struct {
	currentMode              mode
	statusMessage            string
	fileName                 string
	dataBuffer               []string
	textFrame                [][]rune
	startLine                int
	cursorPosBuffer          terminal.Vertex
	cursorPosScreen          terminal.Vertex
	currentCommand           string
	lastUpDownCursorMovement cursorMovement
	firstLineInFrame         int
	lastLineInFrame          int
	screen                   *terminal.Screen
}

// NewEditor is a contructor function for the Editor
func NewEditor(f string) (Editor, error) {
	e := Editor{fileName: f, currentMode: MODE_NORMAL, lastUpDownCursorMovement: CURSOR_UP, startLine: 0, firstLineInFrame: 0}

	var err error
	// TODO: Load only the required portion in memory instead of whole file
	e.dataBuffer, err = fs.ReadFileToLines(f)
	e.statusMessage = fmt.Sprintf("\"%v\"", f)
	if err != nil {
		e.dataBuffer = make([]string, 1)
		e.statusMessage += " [New File]"
	} else {
		charCount := 0
		for _, line := range e.dataBuffer {
			charCount += len(line) + 1
		}
		e.statusMessage += fmt.Sprintf(" %vL, %vC", len(e.dataBuffer), charCount)
	}

	e.cursorPosBuffer = terminal.Vertex{0, 0}

	s, err := terminal.NewScreen()
	if err != nil {
		logger.Debug.Panicf("Error starting the screen %v", err)
		return e, err
	}
	e.screen = s
	e.syncTextFrame(false)
	e.syncCursor()
	e.screen.DisplayStatusBar(e.statusMessage, e.currentCommand)
	return e, nil
}

// HandleUserActions - function that listens the user input channel and handles it appropriately
func (e *Editor) HandleUserActions(c chan actions.Event) {
	for event := range c {
		if event.Kind != "KEY_PRESS" {
			continue
		}
		switch e.currentMode {
		case MODE_INSERT:
			if event.Value == tcell.KeyEscape {
				e.currentMode = MODE_NORMAL
				e.statusMessage = ""
				e.screen.DisplayStatusBar(e.statusMessage, e.currentCommand)
				e.fixHorizontalCursorOverflow()
				e.syncCursor()
				continue
			}
			if e.handleTextAreaCursorMovement(event) {
				e.syncCursor()
				continue
			}
			e.handleKeyInsertMode(event)
		case MODE_NORMAL:
			if e.handleTextAreaCursorMovement(event) {
				e.syncCursor()
				continue
			}
			switch event.Rune {
			case ':':
				e.currentMode = MODE_COMMAND_LINE
				e.statusMessage = ""
				e.handleKeyCommandLineMode(event)
			case 'i':
				e.currentMode = MODE_INSERT
				e.statusMessage = "-- INSERT --"
				e.screen.DisplayStatusBar(e.statusMessage, e.currentCommand)
			}
		case MODE_COMMAND_LINE:
			if event.Value == tcell.KeyEscape {
				e.currentMode = MODE_NORMAL
				e.currentCommand = ""
				e.screen.DisplayStatusBar(e.statusMessage, e.currentCommand)
				continue
			}
			if event.Value == tcell.KeyEnter {
				e.runCommand(e.currentCommand)
			}
			e.handleKeyCommandLineMode(event)
		}

	}
}

func (e *Editor) handleTextAreaCursorMovement(event actions.Event) bool {
	isProcessed := true
	rangeY := len(e.dataBuffer)
	switch event.Value {
	case tcell.KeyLeft:
		if e.cursorPosBuffer.X != 0 {
			e.cursorPosBuffer.X--
			e.syncCursor()
		}
	case tcell.KeyRight:
		e.cursorPosBuffer.X++
		e.fixHorizontalCursorOverflow()
		e.syncCursor()
	case tcell.KeyDown:
		if e.cursorPosBuffer.Y+1 != rangeY {
			e.cursorPosBuffer.Y++
			e.lastUpDownCursorMovement = CURSOR_DOWN
			e.fixHorizontalCursorOverflow()
			e.fixVerticalCursorOverflow()
			e.syncCursor()
		}
	case tcell.KeyUp:
		if e.cursorPosBuffer.Y != 0 {
			e.cursorPosBuffer.Y--
			e.lastUpDownCursorMovement = CURSOR_UP
			e.fixHorizontalCursorOverflow()
			if e.cursorPosBuffer.Y < e.firstLineInFrame {
				e.firstLineInFrame--
				if e.lastLineInFrame-e.firstLineInFrame > e.screen.ScreenDim.Y-1 {
					e.lastLineInFrame--
				}
			}
			e.syncCursor()
		}
	default:
		isProcessed = false
	}
	return isProcessed
}

func (e *Editor) fixVerticalCursorOverflow() {
	if e.cursorPosBuffer.Y > e.lastLineInFrame {
		if e.lastLineInFrame+1 >= len(e.dataBuffer) {
			return
		}
		e.lastLineInFrame++
		if e.lastLineInFrame-e.firstLineInFrame > e.screen.ScreenDim.Y-2 {
			e.firstLineInFrame++
		}
		for i := e.lastLineInFrame; i >= e.firstLineInFrame; i-- {
			e.firstLineInFrame += len(e.dataBuffer[i]) / e.screen.ScreenDim.X
		}
	}
}

func (e *Editor) fixHorizontalCursorOverflow() {
	rangeX := len(e.dataBuffer[e.cursorPosBuffer.Y])
	if e.currentMode != MODE_INSERT {
		rangeX--
		if rangeX < 0 {
			rangeX = 0
		}
	}
	if rangeX < e.cursorPosBuffer.X {
		e.cursorPosBuffer.X = rangeX
	}
}

func (e *Editor) runCommand(cmd string) {
	switch cmd {
	case ":q":
		e.screen.Close()
		os.Exit(0)
	case ":wq":
		fs.WriteLinesToFile(e.fileName, e.dataBuffer)
		e.screen.Close()
		os.Exit(0)
	default:
		e.currentCommand = ""
	}
}

func (e *Editor) handleKeyCommandLineMode(event actions.Event) {
	if event.Rune == 0 {
		return
	}
	if event.Value == tcell.KeyBackspace || event.Value == tcell.KeyBackspace2 {
		e.currentCommand = e.currentCommand[:len(e.currentCommand)-1]
	} else {
		e.currentCommand += string(event.Rune)
	}
	e.screen.DisplayStatusBar(e.statusMessage, e.currentCommand)
}

func (e *Editor) handleKeyInsertMode(event actions.Event) {
	switch event.Value {
	case tcell.KeyEnter:
		newBuffer := make([]string, len(e.dataBuffer)+1)
		copy(newBuffer, e.dataBuffer)
		if e.cursorPosBuffer.Y < len(e.dataBuffer)-1 {
			for i := len(newBuffer) - 1; i > e.cursorPosBuffer.Y; i-- {
				newBuffer[i] = newBuffer[i-1]
			}
		}

		if e.cursorPosBuffer.X < len(e.dataBuffer[e.cursorPosBuffer.Y]) {
			newBuffer[e.cursorPosBuffer.Y] = e.dataBuffer[e.cursorPosBuffer.Y][:e.cursorPosBuffer.X]
			newBuffer[e.cursorPosBuffer.Y+1] = e.dataBuffer[e.cursorPosBuffer.Y][e.cursorPosBuffer.X:]
		}
		e.cursorPosBuffer.Y++
		e.cursorPosBuffer.X = 0
		e.lastUpDownCursorMovement = CURSOR_DOWN
		e.dataBuffer = newBuffer
		e.fixVerticalCursorOverflow()
		e.syncTextFrame(false)
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if e.cursorPosBuffer.X == 0 && e.cursorPosBuffer.Y == 0 {
			return
		}
		if e.cursorPosBuffer.X > 0 {
			// Delete a character in a line
			line := e.dataBuffer[e.cursorPosBuffer.Y]
			e.dataBuffer[e.cursorPosBuffer.Y] = line[:e.cursorPosBuffer.X-1] + line[e.cursorPosBuffer.X:]
			e.cursorPosBuffer.X--
		} else {
			// Merge the contents of this line to previous line
			newBuffer := make([]string, len(e.dataBuffer)-1)
			j := 0
			for i, line := range e.dataBuffer {
				if i == e.cursorPosBuffer.Y {
					continue
				}
				if i == e.cursorPosBuffer.Y-1 {
					line = line + e.dataBuffer[e.cursorPosBuffer.Y]
				}
				newBuffer[j] = line
				j++
			}
			e.cursorPosBuffer.X = len(e.dataBuffer[e.cursorPosBuffer.Y-1])
			e.cursorPosBuffer.Y--
			e.dataBuffer = newBuffer
		}
		e.syncTextFrame(false)
	default:
		if event.Rune == 0 {
			return
		}
		line := e.dataBuffer[e.cursorPosBuffer.Y]
		e.dataBuffer[e.cursorPosBuffer.Y] = line[:e.cursorPosBuffer.X] + string(event.Rune) + line[e.cursorPosBuffer.X:]
		e.cursorPosBuffer.X++
		e.syncTextFrame(false)
	}
}

func (e *Editor) syncTextFrame(dontSyncCursor bool) {

	textFrameTemp := e.fillTextFrameFromTop()

	textFrameTemp = e.fillEmptyLinesIfAnyWithTilde(textFrameTemp)

	e.textFrame = textFrameTemp
	e.screen.DisplayTextFrame(e.textFrame)
	if dontSyncCursor == false {
		e.syncCursor()
	}
}

func (e *Editor) fillEmptyLinesIfAnyWithTilde(textFrameTemp [][]rune) [][]rune {
	//Fill the empty lines in the frame with '~' char
	shouldTilde := true
	for i := len(textFrameTemp) - 1; i > 0; i-- {
		if textFrameTemp[i] != nil {
			shouldTilde = false
			continue
		}
		textFrameTemp[i] = make([]rune, e.screen.ScreenDim.X)
		if shouldTilde {
			textFrameTemp[i][0] = '~'
		}

	}
	return textFrameTemp
}

func (e *Editor) fillTextFrameFromTop() [][]rune {
	x := 0
	y := 0
	textFrameTemp := make([][]rune, e.screen.ScreenDim.Y-1)
	i := e.firstLineInFrame
	for {
		x = 0
		line := e.dataBuffer[i]
		for _, char := range line {
			if textFrameTemp[y] == nil {
				textFrameTemp[y] = make([]rune, e.screen.ScreenDim.X)
			}
			textFrameTemp[y][x] = char
			if x+1 == e.screen.ScreenDim.X {
				x = 0
				y++
			} else {
				x++
			}
		}

		if i+1 == len(e.dataBuffer) || y == e.screen.ScreenDim.Y-2 {
			break
		}
		i++
		y++
	}
	e.lastLineInFrame = i
	return textFrameTemp
}

func (e *Editor) syncCursor() {
	e.cursorPosScreen = terminal.Vertex{e.cursorPosBuffer.X, e.cursorPosBuffer.Y}
	e.cursorPosScreen.Y -= e.firstLineInFrame
	for i := e.firstLineInFrame; i < e.cursorPosBuffer.Y; i++ {
		e.cursorPosScreen.Y += (len(e.dataBuffer[i]) / e.screen.ScreenDim.X)
	}
	for e.cursorPosScreen.X > e.screen.ScreenDim.X {
		e.cursorPosScreen.X -= e.screen.ScreenDim.X
		e.cursorPosScreen.Y++
	}

	e.syncTextFrame(true)
	e.screen.DisplayCursor(e.cursorPosScreen)
}

func (e *Editor) GetTerminalScreen() (s tcell.Screen) {
	return e.screen.TerminalScreen()
}
