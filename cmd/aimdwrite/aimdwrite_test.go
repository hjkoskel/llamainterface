package main

import (
	"llamainterface"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirectiveValue(t *testing.T) {
	prompt, varname, varvalue := pickDirectiveValue("qwerty$TEMP=1.2 jatkuu")
	assert.Equal(t, "TEMP", varname)
	assert.Equal(t, "1.2", varvalue)
	assert.Equal(t, "qwerty jatkuu", prompt)
}

func TestPickDirectives(t *testing.T) {
	q, errQ := PickDirectives("Test here $TEMP=42.69 $", llamainterface.DefaultQueryCompletion())
	assert.Equal(t, nil, errQ)
	assert.Equal(t, q.Temperature, 42.69)

}
