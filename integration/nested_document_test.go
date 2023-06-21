package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestCreateNestedDocument(t *testing.T) {
	t.Run("0", func(t *testing.T) {
		embdyDoc := bson.M{}
		doc := CreateNestedDocument(0)
		assert.Equal(t, embdyDoc, doc)
	})

	t.Run("1", func(t *testing.T) {
		embdyDoc := bson.M{"0": nil}
		doc := CreateNestedDocument(1)
		assert.Equal(t, embdyDoc, doc)
	})

	t.Run("2", func(t *testing.T) {
		embdyDoc := bson.M{"0": bson.M{"1": nil}}
		doc := CreateNestedDocument(2)
		assert.Equal(t, embdyDoc, doc)
	})
}
