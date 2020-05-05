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
	startLine                int
	cursorPos                terminal.Vertex
	currentCommand           string
	lastUpDownCursorMovement cursorMovement
	firstLineInFrame         int
	lastLineInFrame          int
	screen                   *terminal.Screen
	isDirty                  bool
	userEventChannel         chan actions.Event
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
		e.statusMessage += fmt.Sprintf(" %vL, %vC", e.getLinesCount(), charCount)
	}

	e.cursorPos = terminal.Vertex{0, 0}

	s, err := terminal.NewScreen()
	if err != nil {
		logger.Debug.Panicf("Error starting the screen %v", err)
		return e, err
	}
	e.screen = s
	e.syncTextFrame(false)
	e.syncCursor()
	e.syncStatusBar()
	e.listenToEvents()
	return e, nil
}

func (e *Editor) listenToEvents() {
	e.userEventChannel = make(chan actions.Event)
	actions.EventStream(e.userEventChannel, e.screen.TerminalScreen())
	e.HandleUserActions()
}

// HandleUserActions - function that listens the user input channel and handles it appropriately
func (e *Editor) HandleUserActions() {
	for event := range e.userEventChannel {
		if event.Kind != "KEY_PRESS" {
			continue
		}
		switch e.currentMode {
		case MODE_INSERT:
			if event.Value == tcell.KeyEscape {
				e.currentMode = MODE_NORMAL
				e.statusMessage = ""
				e.syncStatusBar()
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
			e.handleKeyNormalMode(event)
		case MODE_COMMAND_LINE:
			if event.Value == tcell.KeyEscape {
				e.currentMode = MODE_NORMAL
				e.currentCommand = ""
				e.syncStatusBar()
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
	rangeY := e.getLinesCount()
	switch event.Value {
	case tcell.KeyLeft:
		if e.cursorPos.X != 0 {
			e.cursorPos.X--
			e.syncCursor()
		}
	case tcell.KeyRight:
		e.cursorPos.X++
		e.fixHorizontalCursorOverflow()
		e.syncCursor()
	case tcell.KeyDown:
		if e.cursorPos.Y+1 != rangeY {
			e.cursorPos.Y++
			e.lastUpDownCursorMovement = CURSOR_DOWN
			e.fixHorizontalCursorOverflow()
			e.fixVerticalCursorOverflow()
			e.syncCursor()
		}
	case tcell.KeyUp:
		if e.cursorPos.Y != 0 {
			e.cursorPos.Y--
			e.lastUpDownCursorMovement = CURSOR_UP
			e.fixHorizontalCursorOverflow()
			if e.cursorPos.Y < e.firstLineInFrame {
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
	for e.cursorPos.Y > e.lastLineInFrame {
		if e.lastLineInFrame+1 >= e.getLinesCount() {
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
	rangeX := e.getCurrentLineLength()
	if e.currentMode != MODE_INSERT {
		rangeX--
		if rangeX < 0 {
			rangeX = 0
		}
	}
	if rangeX < e.cursorPos.X {
		e.cursorPos.X = rangeX
	}
}

func (e *Editor) quit(force bool) {
	if e.isDirty && !force {
		e.statusMessage = "E37: No write since last change (add ! to override)"
		return
	}
	e.screen.Close()
	os.Exit(0)
}

func (e *Editor) runCommand(cmd string) {
	switch cmd {
	case ":q":
		e.quit(false)
	case ":q!":
		e.quit(true)
	case ":wq", ":x":
		if e.isDirty {
			fs.WriteLinesToFile(e.fileName, e.dataBuffer)
		}
		e.screen.Close()
		os.Exit(0)
	default:
		e.currentCommand = ""
	}
}

func (e *Editor) getCurrentLineLength() int {
	return len(e.dataBuffer[e.cursorPos.Y])
}

func (e *Editor) getLinesCount() int {
	return len(e.dataBuffer)
}

func (e *Editor) switchToInsertMode() {
	e.currentMode = MODE_INSERT
	e.statusMessage = "-- INSERT --"
	e.syncStatusBar()
}

func (e *Editor) handleKeyNormalMode(event actions.Event) {
	switch event.Rune {
	case ':':
		e.currentMode = MODE_COMMAND_LINE
		e.statusMessage = ""
		e.handleKeyCommandLineMode(event)
	case 'i':
		e.switchToInsertMode()
	case 'A':
		e.cursorPos.X = e.getCurrentLineLength()
		e.switchToInsertMode()
		e.syncCursor()
	case 'G':
		l := e.getLinesCount() - 1
		if l < 0 {
			l = 0
		}
		e.cursorPos.Y = l
		e.fixVerticalCursorOverflow()
		e.syncTextFrame(false)
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
	e.syncStatusBar()
}

func (e *Editor) handleKeyInsertMode(event actions.Event) {
	e.isDirty = true
	switch event.Value {
	case tcell.KeyEnter:
		newBuffer := make([]string, e.getLinesCount()+1)
		copy(newBuffer, e.dataBuffer)
		if e.cursorPos.Y < e.getLinesCount()-1 {
			for i := len(newBuffer) - 1; i > e.cursorPos.Y; i-- {
				newBuffer[i] = newBuffer[i-1]
			}
		}

		if e.cursorPos.X < e.getCurrentLineLength() {
			newBuffer[e.cursorPos.Y] = e.dataBuffer[e.cursorPos.Y][:e.cursorPos.X]
			newBuffer[e.cursorPos.Y+1] = e.dataBuffer[e.cursorPos.Y][e.cursorPos.X:]
		}
		e.cursorPos.Y++
		e.cursorPos.X = 0
		e.lastUpDownCursorMovement = CURSOR_DOWN
		e.dataBuffer = newBuffer
		e.fixVerticalCursorOverflow()
		e.syncTextFrame(false)
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if e.cursorPos.X == 0 && e.cursorPos.Y == 0 {
			return
		}
		if e.cursorPos.X > 0 {
			// Delete a character in a line
			line := e.dataBuffer[e.cursorPos.Y]
			e.dataBuffer[e.cursorPos.Y] = line[:e.cursorPos.X-1] + line[e.cursorPos.X:]
			e.cursorPos.X--
		} else {
			// Merge the contents of this line to previous line
			newBuffer := make([]string, e.getLinesCount()-1)
			j := 0
			for i, line := range e.dataBuffer {
				if i == e.cursorPos.Y {
					continue
				}
				if i == e.cursorPos.Y-1 {
					line = line + e.dataBuffer[e.cursorPos.Y]
				}
				newBuffer[j] = line
				j++
			}
			e.cursorPos.X = len(e.dataBuffer[e.cursorPos.Y-1])
			e.cursorPos.Y--
			e.dataBuffer = newBuffer
		}
		e.syncTextFrame(false)
	default:
		if event.Rune == 0 {
			return
		}
		line := e.dataBuffer[e.cursorPos.Y]
		e.dataBuffer[e.cursorPos.Y] = line[:e.cursorPos.X] + string(event.Rune) + line[e.cursorPos.X:]
		e.cursorPos.X++
		e.syncTextFrame(false)
	}
}
