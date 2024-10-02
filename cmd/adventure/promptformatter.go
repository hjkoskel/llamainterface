/*
Prompt formatter

Handles prompt formatting for LLM
Possible to change LLM to some other than llama3.1 instruct

TODO:

todo add part of of llama interface library
*/

package main

import (
	"fmt"
	"strings"
)

type PromptExample struct {
	Input  string
	Result string
}

type PromptFormatter interface {
	QueryStop() []string
	ExtractText(userName string, agentName string, prompt string) string
	Format(
		systemPrompt string,
		examples []PromptExample,
		userName string,
		agentName string,
		content string) string
	AppendStopIfNeeded(prompt string) string
}

type Llama31Formatter struct {
}

func (p Llama31Formatter) AppendStopIfNeeded(prompt string) string {
	if strings.HasSuffix(strings.TrimSpace(prompt), "<|eot_id|>") {
		return prompt
	}
	return prompt + "<|eot_id|>\n"
}

func (p Llama31Formatter) QueryStop() []string {
	return []string{"<|eot_id|>"}
}

func (p Llama31Formatter) ExtractText(userName string, agentName string, prompt string) string {
	s := strings.ReplaceAll(prompt, "<|start_header_id|>system<|end_header_id|>", "")
	s = strings.ReplaceAll(s, "<|start_header_id|>player<|end_header_id|>", "")
	s = strings.ReplaceAll(s, "<|start_header_id|>dungeonmaster<|end_header_id|>", "")

	s = strings.ReplaceAll(s, fmt.Sprintf("<|start_header_id|>%s<|end_header_id|>", userName), "")
	s = strings.ReplaceAll(s, fmt.Sprintf("<|start_header_id|>%s<|end_header_id|>", agentName), "")
	s = strings.ReplaceAll(s, "<|eot_id|>", "")
	return strings.TrimSpace(s)
}

func (p Llama31Formatter) Format(systemPrompt string, examples []PromptExample, userName string, agentName string, content string) string {
	var sb strings.Builder

	if 0 < len(systemPrompt) {
		sb.WriteString("<|start_header_id|>system<|end_header_id|>\n\n")
		sb.WriteString(systemPrompt)
		sb.WriteString("<|eot_id|>\n")
	}

	for _, exm := range examples {
		sb.WriteString("<|start_header_id|>")
		sb.WriteString(userName)
		sb.WriteString("<|end_header_id|>\n\n")
		sb.WriteString(exm.Input)
		sb.WriteString("<|eot_id|><|start_header_id|>")
		sb.WriteString(agentName)
		sb.WriteString("<|eot_id|>\n")
	}
	if 0 < len(content) {
		sb.WriteString("<|start_header_id|>")
		sb.WriteString(userName)
		sb.WriteString("<|end_header_id|>\n\n")
		sb.WriteString(strings.TrimSpace(content))
		sb.WriteString("<|eot_id|><|start_header_id|>")
		sb.WriteString(agentName)
		sb.WriteString("<|end_header_id|>") //LLM will predict what comes after this
	}
	return sb.String()
}
