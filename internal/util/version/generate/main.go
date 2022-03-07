package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"go.uber.org/zap"
)

const gitBin = "git"

func runGit(args []string, stdin io.Reader, stdout io.Writer, logger *zap.SugaredLogger) {
	if err := tryGit(args, stdin, stdout, logger); err != nil {
		logger.Fatal(err)
	}
}

func tryGit(args []string, stdin io.Reader, stdout io.Writer, logger *zap.SugaredLogger) error {
	cmd := exec.Command(gitBin, args...)
	logger.Debugf("Running %s", strings.Join(cmd.Args, " "))

	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s failed: %s", strings.Join(args, " "), err)
	}

	return nil
}

func main() {
	var wg sync.WaitGroup
	logger := zap.S().Named("git")

	//git describe --tags --dirty > version.txt
	{
		file := "version.txt"
		args := `describe --tags --dirty`

		wg.Add(1)
		go func() {
			defer wg.Done()
			out, err := os.Create(file)
			if err != nil {
				logger.Fatal("failed to create file:", file)
			}
			defer out.Close()
			runGit(strings.Split(args, " "), nil, out, logger)
		}()
	}

	//git rev-parse HEAD > commit.txt
	{
		file := "commit.txt"
		args := `rev-parse HEAD`

		wg.Add(1)
		go func() {
			defer wg.Done()
			out, err := os.Create(file)
			if err != nil {
				logger.Fatal("failed to create file:", file)
			}
			defer out.Close()
			runGit(strings.Split(args, " "), nil, out, logger)
		}()
	}

	//git branch --show-current > branch.txt
	{
		file := "branch.txt"
		args := `branch --show-current`

		wg.Add(1)
		go func() {
			defer wg.Done()
			out, err := os.Create(file)
			if err != nil {
				logger.Fatal("failed to create file:", file)
			}
			defer out.Close()
			runGit(strings.Split(args, " "), nil, out, logger)
		}()
	}

	wg.Wait()
}
