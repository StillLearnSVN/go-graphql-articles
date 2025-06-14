package tests

import (
	"testing"
	"time"

	"github.com/StillLearnSVN/go-graphql-articles/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestEncodeCursor(t *testing.T) {
	id := 123
	createdAt := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	cursor := models.EncodeCursor(id, createdAt)
	assert.NotEmpty(t, cursor)

	decodedID, decodedTime, err := models.DecodeCursor(cursor)
	assert.NoError(t, err)
	assert.Equal(t, id, decodedID)
	assert.Equal(t, createdAt.Unix(), decodedTime.Unix())
}

func TestDecodeCursor_InvalidCursor(t *testing.T) {
	_, _, err := models.DecodeCursor("invalid-cursor")
	assert.Error(t, err)
}

func TestDecodeCursor_InvalidFormat(t *testing.T) {
	// This would be a malformed base64 string that decodes to invalid format
	_, _, err := models.DecodeCursor("aW52YWxpZA==") // "invalid" in base64
	assert.Error(t, err)
}
