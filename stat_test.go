package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatPersist(t *testing.T) {
	var originalCount int
	_ = db.QueryRow("select count(*) from incoming").Scan(&originalCount)

	s := Stat{"Mark", jt}
	s.persist()

	var newCount int
	_ = db.QueryRow("select count(*) from incoming").Scan(&newCount)
	assert.Equal(t, newCount, originalCount+1)
}

func TestStatKeys(t *testing.T) {
	s := Stat{"Mark", jt}
	keys := s.keys()
	expected := []string{"active", "name", "number"}
	assert.Equal(t, expected, keys)
}
