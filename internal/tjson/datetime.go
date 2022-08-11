package tjson

import (
	"bytes"
	"encoding/json"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"time"
)

type dateTimeType time.Time

func (dt *dateTimeType) tjsontype() {}

func (dt *dateTimeType) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		panic("null data")
	}

	r := bytes.NewReader(data)
	dec := json.NewDecoder(r)

	var o time.Time
	if err := dec.Decode(&o); err != nil {
		return lazyerrors.Error(err)
	}
	if err := checkConsumed(dec, r); err != nil {
		return lazyerrors.Error(err)
	}

	*dt = dateTimeType(o)
	return nil
}

func (dt *dateTimeType) MarshalJSON() ([]byte, error) {
	res, err := json.Marshal(time.Time(*dt))
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	return res, nil
}

// check interfaces
var (
	_ tjsontype = (*dateTimeType)(nil)
)
