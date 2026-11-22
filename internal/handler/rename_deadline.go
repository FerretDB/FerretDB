package handler

import (
	"math"
	"math/big"
	"time"

	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/wire/wirebson"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const compatGateCapacity int64 = 1 << 30

// renameTimeLimit applies a command deadline to the SQL compatibility path and
// to waiting for its gateway-wide namespace barrier. Native DocumentDB commands
// already interpret their own maxTimeMS specifications.
func renameTimeLimit(doc wirebson.AnyDocument) (time.Duration, error) {
	value, err := getOptionalParamAny(doc, "maxTimeMS", int32(0))
	if err != nil {
		return 0, err
	}
	var milliseconds float64
	switch value := value.(type) {
	case int32:
		milliseconds = float64(value)
	case int64:
		milliseconds = float64(value)
	case float64:
		milliseconds = value
	case wirebson.Decimal128:
		decimal := bson.NewDecimal128(value.H, value.L)
		if decimal.IsNaN() {
			return 0, nil
		}
		coefficient, exponent, parseErr := decimal.BigInt()
		if parseErr != nil {
			return 0, mongoerrors.New(mongoerrors.ErrBadValue, "Invalid maxTimeMS")
		}
		if coefficient.Sign() == 0 || exponent < -34 {
			return 0, nil
		}
		if exponent > 10 {
			return 0, mongoerrors.New(mongoerrors.ErrBadValue, "Invalid maxTimeMS")
		}
		if exponent >= 0 {
			coefficient.Mul(coefficient, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(exponent)), nil))
		} else {
			divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(-exponent)), nil)
			remainder := new(big.Int)
			quotient, _ := new(big.Int).QuoRem(coefficient, divisor, remainder)
			half := new(big.Int).Lsh(new(big.Int).Abs(remainder), 1).Cmp(divisor)
			if half > 0 || half == 0 && quotient.Bit(0) != 0 {
				quotient.Add(quotient, big.NewInt(int64(coefficient.Sign())))
			}
			coefficient = quotient
		}
		if !coefficient.IsInt64() {
			return 0, mongoerrors.New(mongoerrors.ErrBadValue, "Invalid maxTimeMS")
		}
		milliseconds = float64(coefficient.Int64())
	default:
		return 0, mongoerrors.New(mongoerrors.ErrBadValue, "maxTimeMS must be numeric")
	}
	if math.IsNaN(milliseconds) || math.IsInf(milliseconds, 0) || milliseconds < 0 || milliseconds > math.MaxInt32 || milliseconds != math.Trunc(milliseconds) {
		return 0, mongoerrors.New(mongoerrors.ErrBadValue, "maxTimeMS must be a nonnegative integral number within the int32 range")
	}
	return time.Duration(milliseconds) * time.Millisecond, nil
}
