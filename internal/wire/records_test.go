package wire

import (
	"bufio"
	"os"
	"testing"
)

type placeholderMsgs struct {
	Header *MsgHeader
	Body   MsgBody
}

func fetchTestCases() ([]*placeholderMsgs, error) {
	// TODO: iterate over every file in directory

	// TODO: use OpenFile and fetch data on fly
	//b, err := os.ReadFile("./records/2022-22-09_00:50:53_172.20.0.3:52888.bin")
	//f, err := os.OpenFile("records/2022-22-09_00:50:53_172.20.0.3:52888.bin", os.O_RDONLY, 0o666)
	f, err := os.OpenFile("records/2022-22-09_00:50:53_172.20.0.3:52888.bin", os.O_RDONLY, 0o666)
	if err != nil {
		return nil, err
	}
	var resMsgs []*placeholderMsgs
	r := bufio.NewReader(f)
	for {
		header, body, err := ReadMessage(r)
		if err != nil {
			return resMsgs, err
		}
		resMsgs = append(resMsgs, &placeholderMsgs{header, body})
	}
}

func FuzzRecords(f *testing.F) {
	msgs, err := fetchTestCases()
	if err != nil {
		f.Error(err)
	}
	f.Log(msgs)
	//fuzzMessages(f)
}
