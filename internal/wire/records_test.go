package wire

import (
	"bufio"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func fetchTestCases() ([]testCase, error) {
	// TODO: iterate over every file in directory
	recordsPath := "../../records"

	var recordFilesPaths []string
	err := filepath.WalkDir(recordsPath, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(entry.Name()) == ".bin" {
			recordFilesPaths = append(recordFilesPaths, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var resMsgs []testCase
	for _, path := range recordFilesPaths {
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
			resMsgs = append(
				resMsgs,
				testCase{
					msgHeader: header,
					msgBody:   body,
				},
			)
		}
	}
	return resMsgs, nil
}

func FuzzRecords(f *testing.F) {
	msgs, err := fetchTestCases()
	if err != nil {
		f.Error(err)
	}
	f.Log(msgs)
	fuzzMessages(f, msgs)
}
