package wire

import (
	"bufio"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// loadRecords gets recursively all .bin file from recordsPath directory,
// parses their content to wire Msgs and return them as []testCase array
// with headerB and bodyB fields set.
func loadRecords(recordsPath string) ([]testCase, error) {
	// Load recursively every file path with ".bin" extension from recordsPath directory
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
		f, err := os.Open(path)
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
				return nil, err
			}

			headBytes, err := header.MarshalBinary()
			if err != nil {
				return nil, err
			}

			bodyBytes, err := body.MarshalBinary()
			if err != nil {
				return nil, err
			}

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
