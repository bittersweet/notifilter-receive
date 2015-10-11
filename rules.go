package main

import (
	"encoding/json"
	"log"
	"strconv"
)

type Rule struct {
	Key      string
	Type     string
	Optional string
	Value    string
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

	if r.Type == "boolean" {
		val := parsed[r.Key]
		needed_val, _ := strconv.ParseBool(r.Value)
		if val.(bool) != needed_val {
			return false
		}
	}
	if r.Type == "string" {
		val := parsed[r.Key]
		needed_val := r.Value
		if val.(string) != needed_val {
			return false
		}
	}
	if r.Type == "number" {
		val := parsed[r.Key].(float64)
		needed_val, _ := strconv.ParseFloat(r.Value, 64)

		if r.Optional == "eq" {
			if val != needed_val {
				return false
			}
		} else if r.Optional == "gt" {
			if val <= needed_val {
				return false
			}
		} else if r.Optional == "lt" {
			if val >= needed_val {
				return false
			}
		}
	}

	return true
}
