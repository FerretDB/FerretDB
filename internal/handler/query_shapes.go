package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/netip"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/v2/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
)

// shapeRecorder is opt-in laboratory instrumentation, not a production sanitizer.
// Field names are retained, so only synthetic data is permitted. Authentication
// and user-management commands are excluded entirely. Scalar values are omitted.
type shapeRecorder struct {
	mu        sync.Mutex
	path      string
	until     time.Time
	remaining int
	bytesLeft int
	peer      netip.Addr
}

func newShapeRecorder() (*shapeRecorder, error) {
	path := os.Getenv("FERRETDB_QUERY_SHAPES_PATH")
	if path == "" {
		return nil, nil
	}
	if os.Getenv("FERRETDB_QUERY_SHAPES_SYNTHETIC_ONLY") != "true" {
		return nil, fmt.Errorf("query shape capture requires synthetic-only acknowledgement")
	}
	var peer netip.Addr
	if s := os.Getenv("FERRETDB_QUERY_SHAPES_PEER"); s != "" {
		var err error
		peer, err = netip.ParseAddr(s)
		if err != nil {
			return nil, fmt.Errorf("invalid capture peer")
		}
	}
	duration := 5 * time.Minute
	if s := os.Getenv("FERRETDB_QUERY_SHAPES_DURATION"); s != "" {
		var err error
		duration, err = time.ParseDuration(s)
		if err != nil || duration <= 0 || duration > 10*time.Minute {
			return nil, fmt.Errorf("query shape duration must be positive and at most ten minutes")
		}
	}
	path += "." + time.Now().UTC().Format("20060102T150405.000000000") + ".jsonl"
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("create shape capture: %w", err)
	}
	if err = f.Close(); err != nil {
		return nil, err
	}
	return &shapeRecorder{path: path, until: time.Now().Add(duration), remaining: 10000, bytesLeft: 8 * 1024 * 1024, peer: peer}, nil
}

type shapeField struct {
	Name  string `json:"name"`
	Shape any    `json:"shape"`
}

func shapeValue(v any, depth int) any {
	if depth > 20 {
		return "depth-limit"
	}
	switch v := v.(type) {
	case wirebson.AnyDocument:
		d, err := v.Decode()
		if err != nil {
			return "invalid-document"
		}
		out := make([]shapeField, 0, len(d.FieldNames()))
		for _, key := range d.FieldNames() {
			if len(out) == 128 {
				break
			}
			out = append(out, shapeField{Name: key, Shape: shapeValue(d.Get(key), depth+1)})
		}
		return out
	case wirebson.AnyArray:
		a, err := v.Decode()
		if err != nil {
			return "invalid-array"
		}
		out := make([]any, 0)
		for item := range a.Values() {
			if len(out) == 32 {
				break
			}
			out = append(out, shapeValue(item, depth+1))
		}
		return out
	default:
		return fmt.Sprintf("%T", v)
	}
}

func (r *shapeRecorder) record(ctx context.Context, req *middleware.Request, resp *middleware.Response) {
	if r == nil {
		return
	}
	if r.peer.IsValid() && conninfo.Get(ctx).Peer.Addr().Unmap() != r.peer.Unmap() {
		return
	}
	cmd := req.Document().Command()
	switch cmd {
	case "saslStart", "saslContinue", "authenticate", "createUser", "updateUser", "dropUser", "usersInfo":
		return
	}
	db, _ := req.Document().Get("$db").(string)
	if db != "admin" && !strings.HasPrefix(db, "unifi-ferretdb-test-") {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.remaining == 0 || time.Now().After(r.until) {
		return
	}
	doc, err := req.DocumentDeep()
	if err != nil {
		return
	}
	code := 0
	ok := false
	if resp != nil {
		code = int(resp.ErrorCode())
		ok = resp.OK()
	}
	b, err := json.Marshal(struct {
		Command string `json:"command"`
		Shape   any    `json:"shape"`
		OK      bool   `json:"ok"`
		Code    int    `json:"code"`
	}{cmd, shapeValue(doc, 0), ok, code})
	if err != nil || len(b) > 65536 || len(b)+1 > r.bytesLeft {
		return
	}
	f, err := os.OpenFile(r.path, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		r.remaining = 0
		return
	}
	_, err = f.Write(append(b, '\n'))
	_ = f.Close()
	if err != nil {
		r.remaining = 0
		return
	}
	r.remaining--
	r.bytesLeft -= len(b) + 1
}
