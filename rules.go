package main

import (
	"encoding/json"
	"log"
	"strconv"
)

type Rule struct {
	Key     string
	Type    string
	Setting string
	Value   string
}

func (r *Rule) Met(s *Stat) bool {
	var parsed map[string]interface{}
	err := json.Unmarshal([]byte(s.Value), &parsed)
	if err != nil {
		log.Fatal("json.Unmarshal", err)
	}

	// check if key is in the map
	// first value is actual value of key in the map
	if _, ok := parsed[r.Key]; !ok {
		return false
	}

	switch r.Type {
	case "boolean":
		return metBool(r, parsed)
	case "string":
		return metString(r, parsed)
	case "number":
		return metNumber(r, parsed)
	}

	return true
}

func metBool(r *Rule, parsed map[string]interface{}) bool {
	val := parsed[r.Key]
	needed_val, _ := strconv.ParseBool(r.Value)
	if val.(bool) != needed_val {
		return false
	}

	return true
}

func metString(r *Rule, parsed map[string]interface{}) bool {
	val := parsed[r.Key]
	needed_val := r.Value
	if val.(string) != needed_val {
		return false
	}

	return true
}

func metNumber(r *Rule, parsed map[string]interface{}) bool {
	val := parsed[r.Key].(float64)
	needed_val, _ := strconv.ParseFloat(r.Value, 64)

	switch r.Setting {
	case "eq":
		if val != needed_val {
			return false
		}
	case "gt":
		if val <= needed_val {
			return false
		}
	case "lt":
		if val >= needed_val {
			return false
		}
	}

	return true
}
