package main

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestParseKeyValues(t *testing.T) {
	s := "{\"type\": \"Couch\", \"name\": \"Living Room Sofa\", \"description\": \"Brown leather sofa with golden buttons\"},{\"type\": \"Coffee Table\", \"name\": \"Mid Century Modern Coffee Table\", \"description\": \"Glass top, metal base, dimly lit"
	pairs := GetKeyValuePairs(s)

	ref := []KeyValueStringPair{
		{"type", "Couch"},
		{"name", "Living Room Sofa"},
		{"description", "Brown leather sofa with golden buttons"},
		{"type", "Coffee Table"},
		{"name", "Mid Century Modern Coffee Table"}}

	assert.Equal(t, pairs, ref)
	refArr := KeyValueStringPairArr(ref)
	scene := refArr.ToSceneryOutput()

	assert.Equal(t, SceneryOutput([]SceneryItem{
		SceneryItem{Type: "Couch", Name: "Living Room Sofa", Description: "Brown leather sofa with golden buttons"},
	}), scene)
}
