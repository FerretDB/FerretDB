package common

import (
	"bufio"
	"io"
	"strings"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

func ParseOSRelease(reader io.Reader) (map[string]string, error) {
	scanner := bufio.NewScanner(reader)

	configParams := make(map[string]string)

	for scanner.Scan() {
		str := strings.Split(scanner.Text(), "=")

		configParams[str[0]] = str[1]
	}

	if err := scanner.Err(); err != nil {
		return nil, lazyerrors.Error(err)
	}

	return configParams, nil
}
