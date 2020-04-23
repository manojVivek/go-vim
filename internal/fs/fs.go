package fs

import (
	"bufio"
	"os"
	"path/filepath"
)

// ReadFileToLines - Reads the given file into a string array of lines and returns it
func ReadFileToLines(fileName string) ([]string, error) {
	fileAbsPath, err := filepath.Abs(fileName)
	file, err := os.Open(fileAbsPath)
	var lines []string
	if err != nil {
		return lines, err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return lines, err
	}
	return lines, err
}

// WriteLinesToFile - Writes the given []string to the given file
func WriteLinesToFile(fileName string, lines []string) error {
	fileAbsPath, err := filepath.Abs(fileName)
	file, err := os.OpenFile(fileAbsPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	datawriter := bufio.NewWriter(file)

	for _, data := range lines {
		_, err = datawriter.WriteString(data + "\n")
		if err != nil {
			return err
		}
	}
	datawriter.Flush()
	return nil
}
