package logger

import (
	"fmt"
	"log"
	"os"
)

// Debug is used to log debug messages
var Debug *log.Logger

// InitLogger will initialize a debug logger to a temp file
func InitLogger() {
	file, err := os.OpenFile("/tmp/go-vim.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Error opening log file %v", err)
		return
	}

	Debug = log.New(file, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
}
