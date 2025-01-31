package logging

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMongoLog(t *testing.T) {
	now := time.Now()

	log := mongoLog{
		Timestamp: primitive.NewDateTimeFromTime(now),
	}

	extJSON, err := bson.MarshalExtJSON(&log, false, false)
	require.NoError(t, err)

	assert.Equal(t, fmt.Sprintf(`{"t":{"$date":"%s"}}`, now.Format(time.RFC3339)), string(extJSON))
}
