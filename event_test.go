package main

import (
	"testing"

	"github.com/jmoiron/sqlx/types"
	"github.com/stretchr/testify/assert"
)

var jtEvent = types.JsonText(`{"active": true, "name": "Go", "number": 12}`)

func TestEventPersist(t *testing.T) {
	var originalCount int
	_ = db.QueryRow("SELECT COUNT(*) FROM incoming").Scan(&originalCount)

	e := Event{"Mark", jtEvent}
	e.persist()

	var newCount int
	_ = db.QueryRow("SELECT COUNT(*) FROM incoming").Scan(&newCount)
	assert.Equal(t, newCount, originalCount+1)
}
