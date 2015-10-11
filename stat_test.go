package main

import (
	"sort"
	"testing"

	"github.com/jmoiron/sqlx/types"
	"github.com/stretchr/testify/assert"
)

func TestStatPersist(t *testing.T) {
	var originalCount int
	_ = db.QueryRow("SELECT COUNT(*) FROM incoming").Scan(&originalCount)

	s := Stat{"Mark", jt}
	s.persist()

	var newCount int
	_ = db.QueryRow("SELECT COUNT(*) FROM incoming").Scan(&newCount)
	assert.Equal(t, newCount, originalCount+1)
}

func TestStatKeys(t *testing.T) {
	var jt = types.JsonText(`{"active": true, "name": "Go", "number": "12"}`)

	s := Stat{"Mark", jt}
	result := s.keys()
	expected := []string{"active", "name", "number"}

	sort.Strings(result)
	sort.Strings(expected)
	assert.Equal(t, expected, result)
}
