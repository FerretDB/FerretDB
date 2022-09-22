package wire

import (
	"bufio"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func fetchRecords(recordsPath string) ([]testCase, error) {

	// Fetch resursively every file path with ".bin" extension from recordsPath directory
	var recordFiles []string
	err := filepath.WalkDir(recordsPath, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(entry.Name()) == ".bin" {
			recordFiles = append(recordFiles, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Read every record file, parse their content to wire messages
	// and store them in the testCase struct
	var resMsgs []testCase
	for _, path := range recordFiles {
		f, err := os.OpenFile(path, os.O_RDONLY, 0o666)
		if err != nil {
			return nil, err
		}

		r := bufio.NewReader(f)

		for {
			header, body, err := ReadMessage(r)
			if err == io.EOF {
				break
			}
			if err != nil {
				return resMsgs, err
			}

			headBytes := []byte(header.String())
			bodyBytes := []byte(body.String())
			resMsgs = append(
				resMsgs,
				testCase{
					headerB: headBytes,
					bodyB:   bodyBytes,
				},
			)
		}
	}
	return resMsgs, nil
}

func FuzzRecords(f *testing.F) {
	msgs, err := fetchRecords("./records/")
	if err != nil {
		f.Error(err)
	}
	f.Log(msgs)
	fuzzMessages(f, msgs)
}
