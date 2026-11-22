package clientconn

import (
	"bufio"
	"bytes"
	"context"
	"testing"

	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"
)

type unacknowledgedHandler struct{ processed int }

func (*unacknowledgedHandler) Run(context.Context)              {}
func (*unacknowledgedHandler) Describe(chan<- *prometheus.Desc) {}
func (*unacknowledgedHandler) Collect(chan<- prometheus.Metric) {}
func (h *unacknowledgedHandler) Handle(_ context.Context, req *middleware.Request) (*middleware.Response, error) {
	h.processed++
	if req.Document().Get("fail") == true {
		return middleware.ResponseErr(req, mongoerrors.New(mongoerrors.ErrDuplicateKey, "synthetic duplicate")), nil
	}
	return middleware.ResponseDoc(req, wirebson.MustDocument("ok", float64(1)))
}

func TestUnacknowledgedWritesExecuteWithoutReplies(t *testing.T) {
	t.Parallel()
	h := &unacknowledgedHandler{}
	m := middleware.New(&middleware.NewOpts{Mode: middleware.NormalMode, DocDB: h, Metrics: middleware.NewMetrics(), L: testutil.Logger(t)})
	c := &conn{m: m}
	var output bytes.Buffer
	writer := bufio.NewWriter(&output)
	for id := int32(1); id <= 3; id++ {
		msg := wire.MustOpMsg("ping", int32(1), "$db", "admin", "fail", id == 2)
		if id < 3 {
			msg.Flags = wire.OpMsgFlags(wire.OpMsgMoreToCome)
		}
		req, err := middleware.RequestWire(&wire.MsgHeader{OpCode: wire.OpCodeMsg, RequestID: id, MessageLength: int32(wire.MsgHeaderLen + msg.Size())}, msg)
		if err != nil {
			t.Fatal(err)
		}
		if err = c.processRequest(context.Background(), req, writer); err != nil {
			t.Fatal(err)
		}
		if id < 3 && output.Len() != 0 {
			t.Fatal("unacknowledged operation produced a reply")
		}
	}
	if h.processed != 3 {
		t.Fatal("an unacknowledged operation was skipped")
	}
	header, _, err := wire.ReadMessage(bufio.NewReader(&output))
	if err != nil {
		t.Fatal(err)
	}
	if header.ResponseTo != 3 || output.Len() != 0 {
		t.Fatal("later acknowledged response was contaminated")
	}
}
