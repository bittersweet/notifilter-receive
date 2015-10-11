package main

import (
	"encoding/json"
	"log"
	"strconv"
)

type rule struct {
	Key     string `json:"key"`
	Type    string `json:"type"`
	Setting string `json:"setting"`
	Value   string `json:"value"`
}

func (r *rule) Met(e *Event) bool {
	var parsed map[string]interface{}
	err := json.Unmarshal([]byte(e.Data), &parsed)
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

func metBool(r *rule, parsed map[string]interface{}) bool {
	val := parsed[r.Key]
	neededVal, _ := strconv.ParseBool(r.Value)
	if val.(bool) != neededVal {
		return false
	}

	return true
}

func metString(r *rule, parsed map[string]interface{}) bool {
	val := parsed[r.Key]
	neededVal := r.Value
	if val.(string) != neededVal {
		return false
	}

	return true
}

func metNumber(r *rule, parsed map[string]interface{}) bool {
	val := parsed[r.Key].(float64)
	neededVal, _ := strconv.ParseFloat(r.Value, 64)

	switch r.Setting {
	case "eq":
		if val != neededVal {
			return false
		}
	case "gt":
		if val <= neededVal {
			return false
		}
	case "lt":
		if val >= neededVal {
			return false
		}
	}

	return true
}
