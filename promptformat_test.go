package llamainterface

import (
	"testing"
)

// Helper function to compare two LLMMessages slices
func compareMessages(t *testing.T, expected, actual LLMMessages) {
	if len(expected) != len(actual) {
		t.Fatalf("Expected %d messages, but got %d", len(expected), len(actual))
	}
	for i := range expected {
		if expected[i].Type != actual[i].Type || expected[i].Content != actual[i].Content {
			t.Errorf("Message %d mismatch.\nExpected: %+v\nActual: %+v", i, expected[i], actual[i])
		}
	}
}

// Test cases for ToPrompt
func TestToPrompt_Success(t *testing.T) {
	formatter := &Llama3InstructFormatter{}

	messages := LLMMessages{
		{Type: "system", Content: "sysGoesHere"},
		{Type: "user", Content: "Hello!"},
		{Type: "assistant", Content: ""},
	}
	expected := `<|begin_of_text|><|start_header_id|>system<|end_header_id|>

sysGoesHere<|eot_id|>
<|start_header_id|>user<|end_header_id|>

Hello!<|eot_id|>
<|start_header_id|>assistant<|end_header_id|>

`

	prompt, err := formatter.ToPrompt(messages)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if prompt != expected {
		t.Errorf("Expected:\n%s\nGot:\n|%s|", expected, prompt)
	}
}

func TestToPrompt_ReservedStringInType(t *testing.T) {
	formatter := &Llama3InstructFormatter{}
	messages := LLMMessages{
		{Type: "<|begin_of_text|>", Content: "Hello!"},
	}

	_, err := formatter.ToPrompt(messages)
	if err == nil {
		t.Fatal("Expected error for reserved string in Type but got nil")
	}
	expectedErr := "the <|begin_of_text|> is not user prompt type"
	if err.Error() != expectedErr {
		t.Errorf("Expected error: %s, but got: %v", expectedErr, err)
	}
}

func TestToPrompt_ReservedStringInContent(t *testing.T) {
	formatter := &Llama3InstructFormatter{}
	messages := LLMMessages{
		{Type: "user", Content: "<|begin_of_text|>"},
	}

	_, err := formatter.ToPrompt(messages)
	if err == nil {
		t.Fatal("Expected error for reserved string in Content but got nil")
	}
	expectedErr := "message0 contains reserved string <|begin_of_text|> in content"
	if err.Error() != expectedErr {
		t.Errorf("Expected error: %s, but got: %v", expectedErr, err)
	}
}

func TestToPrompt_NonEmptyLastContent(t *testing.T) {
	formatter := &Llama3InstructFormatter{}
	messages := LLMMessages{
		{Type: "user", Content: "Hello!"},
		{Type: "assistant", Content: "Non-empty content"},
	}

	_, err := formatter.ToPrompt(messages)
	if err == nil {
		t.Fatal("Expected error for non-empty last content but got nil")
	}
	expectedErr := "last content is not empty content: Non-empty content"
	if err.Error() != expectedErr {
		t.Errorf("Expected error: %s, but got: %v", expectedErr, err)
	}
}

// Test cases for Parse
func TestParse_Success(t *testing.T) {
	formatter := &Llama3InstructFormatter{}
	input := `<|begin_of_text|><|start_header_id|>system<|end_header_id|>

<|eot_id|>
<|start_header_id|>user<|end_header_id|>

Hello!<|eot_id|>
<|start_header_id|>assistant<|end_header_id|>

`

	expected := LLMMessages{
		{Type: "system", Content: ""},
		{Type: "user", Content: "Hello!"},
		{Type: "assistant", Content: ""},
	}

	messages, err := formatter.Parse(input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	compareMessages(t, expected, messages)
}
