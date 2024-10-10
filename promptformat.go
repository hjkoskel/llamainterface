/*
https://github.com/huggingface/chat-ui/blob/main/PROMPTS.md
*/
package llamainterface

import (
	"fmt"
	"strings"
)

// Optional default values. Use system
const (
	LLMMESSAGETYPE_SYSTEM    = "system"
	LLMMESSAGETYPE_USER      = "user"
	LLMMESSAGETYPE_ASSISTANT = "assistant"
)

// LLMJobMessage tries to be same way as on openai
type LLMMessage struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

type LLMMessages []LLMMessage //Leave last entry content as

func (p *LLMMessages) GetWithType(typ string) []string {
	result := []string{}
	for _, a := range *p {
		if a.Type == typ {
			result = append(result, a.Content)
		}
	}
	return result
}

func isOneOfStrings(s string, arr []string) bool {
	for _, a := range arr {
		if s == a {
			return true
		}
	}
	return false
}

// Check sanity of llmMessages. One system prompt, starts with  system,user,assistant,user,assistant
func (p *LLMMessages) SanityCheck(system string, assistant string, users []string) error {
	if len(*p) < 1 {
		return fmt.Errorf("only %v messages", len(*p))
	}
	/*
		if (*p)[0].Type != system {
			return fmt.Errorf("first is not system prompt")
		}*/
	if 1 < len(p.GetWithType(system)) {
		return fmt.Errorf("too many %d system prompts:%s", len(p.GetWithType(system)), system)
	}

	/* Do not care
	userTurn := true
	for i, d := range *p {
		if d.Type == system {
			continue
		}
		if userTurn {
			if !isOneOfStrings(d.Type, users) {
				return fmt.Errorf("the %s is not user prompt type (valid users are %#v) index=%v", d.Type, users, i)
			}
		} else {
			if d.Type != assistant {
				return fmt.Errorf("the %s is not assistant: %s  index=%v", d.Type, assistant, i)
			}
		}
		userTurn = !userTurn
	}
	*/
	return nil
}

type PromptFormatter interface {
	Cleanup(s string) string                   //Just removes all fishy control things and return string only
	ToPrompt(data LLMMessages) (string, error) //Fails if contains control characters
	Parse(s string) (LLMMessages, error)
	QueryStop() []string
}

func stringContainsAnyStrings(s string, candicates []string) string {
	for _, c := range candicates {
		if strings.Contains(s, c) {
			return c
		}
	}
	return ""
}

type Llama2Formatter struct {
	System    string
	Users     []string
	Assistant string
}

func (p *Llama2Formatter) QueryStop() []string {
	return []string{"<|eot_id|>"}
}

func (p *Llama2Formatter) SetDefaultsIfNeeded() {
	if len(p.System) == 0 {
		p.System = LLMMESSAGETYPE_SYSTEM
	}
	if len(p.Users) == 0 {
		p.Users = []string{LLMMESSAGETYPE_USER}
	}
	if len(p.Assistant) == 0 {
		p.Assistant = LLMMESSAGETYPE_ASSISTANT
	}
}

func (p *Llama2Formatter) ToPrompt(data LLMMessages) (string, error) {
	var sb strings.Builder
	p.SetDefaultsIfNeeded()
	errSanity := data.SanityCheck(p.System, p.Assistant, p.Users)
	if errSanity != nil {
		return "", errSanity
	}

	sysPrompts := data.GetWithType(LLMMESSAGETYPE_SYSTEM)
	if 1 < len(sysPrompts) {
		return "", fmt.Errorf("too many system prompts %v", len(sysPrompts))
	}
	sysprompt := ""
	if 0 < len(sysPrompts) {
		sysprompt = sysPrompts[0]
	}

	sb.WriteString(fmt.Sprintf("<s>[INST] <<SYS>>\n%s\n<</SYS>>\n\n\n", sysprompt))
	count := 0
	for _, d := range data {
		if d.Type == LLMMESSAGETYPE_SYSTEM { //Assume system,user,system...
			continue
		}
		if count%2 == 0 {
			sb.WriteString(" " + d.Content + " [/INST]")
		} else {
			sb.WriteString(" " + d.Content + " </s><s>[INST]")
		}
		count++
	}
	return sb.String(), nil
}

/*
TODO IMPLEMENT WHEN NEEED!
func (p *Llama2Formatter) Parse(s string) (LLMMessages, error) {

}
*/

/*
Llama3Instructformatter
https://medium.com/@renjuhere/llama-3-first-look-c19d99b4933b

there can be many users
*/

type Llama3InstructFormatter struct {
	System    string
	Users     []string //could be many? or not?
	Assistant string
}

func (p *Llama3InstructFormatter) QueryStop() []string {
	return []string{"<|eot_id|>"}
}

func (p *Llama3InstructFormatter) setDefaultsIfNeeded() {
	if len(p.System) == 0 {
		p.System = LLMMESSAGETYPE_SYSTEM
	}
	if len(p.Users) == 0 {
		p.Users = []string{LLMMESSAGETYPE_USER}
	}
	if len(p.Assistant) == 0 {
		p.Assistant = LLMMESSAGETYPE_ASSISTANT
	}
}

func (p *Llama3InstructFormatter) Cleanup(s string) string {
	//Used for
	splitter := "<|end_header_id|>"
	arr := strings.Split(s, splitter)
	s = arr[len(arr)-1]

	s = strings.ReplaceAll(s, "<|begin_of_text|>", "")
	s = strings.ReplaceAll(s, "<|start_header_id|>", "")
	s = strings.ReplaceAll(s, "<|end_header_id|>", "")
	s = strings.ReplaceAll(s, "<|eot_id|>", "")
	return s
}

func (p *Llama3InstructFormatter) ToPrompt(data LLMMessages) (string, error) {
	var sb strings.Builder
	reserved := []string{"<|begin_of_text|>", "<|start_header_id|>"}

	p.setDefaultsIfNeeded()
	errSanity := data.SanityCheck(p.System, p.Assistant, p.Users)
	if errSanity != nil {
		return "", errSanity
	}

	sb.WriteString("<|begin_of_text|>")
	for i, item := range data {
		cont := stringContainsAnyStrings(item.Type, reserved)
		if cont != "" {
			return "", fmt.Errorf("message%d contains reserved string %s in type", i, cont)
		}
		cont = stringContainsAnyStrings(item.Content, reserved)
		if cont != "" {
			return "", fmt.Errorf("message%d contains reserved string %s in content", i, cont)
		}

		//sb.WriteString(fmt.Sprintf("<|start_header_id|>%s<|end_header_id|>\n\n%s<|eot_id|>\n", item.Type, item.Content))

		sb.WriteString(fmt.Sprintf("<|start_header_id|>%s<|end_header_id|>\n\n%s", item.Type, item.Content))
		if i == len(data)-1 { //Idea is to have assistant (or dungeonmaster) as last entry
			if len(item.Content) != 0 {
				return "", fmt.Errorf("last content is not empty content: %s", item.Content)
			}
		} else {
			sb.WriteString("<|eot_id|>\n")
		}
	}
	return sb.String(), nil
}

func (p *Llama3InstructFormatter) Parse(s string) (LLMMessages, error) {
	s = strings.TrimSpace(s)
	s = strings.Replace(s, "<|begin_of_text|>", "", 1) //TODO need to test this?

	if strings.HasSuffix(s, "<|end_of_text|>") {
		s = strings.Replace(s, "<|end_of_text|>", "", 1) //TODO check end of text?
	}
	s = strings.Replace(s, "<|start_header_id|>", "", 1) //only one!

	pieces := strings.Split(s, "<|start_header_id|>")

	result := make([]LLMMessage, len(pieces))
	for i, piece := range pieces {
		txt := strings.TrimSpace(piece)
		if !strings.HasSuffix(txt, "<|eot_id|>") && i != len(pieces)-1 {
			return result, fmt.Errorf("missing <|eot_id|> on index %v", i)
		}
		txt = strings.Replace(txt, "<|eot_id|>", "", 1) //TODO check end of text?
		typeContent := strings.Split(txt, "<|end_header_id|>")
		if len(typeContent) != 2 {
			return result, fmt.Errorf("invalid format type content pair %v  : %#v", i, typeContent)
		}
		result[i].Type = strings.TrimSpace(typeContent[0])
		result[i].Content = strings.TrimSpace(typeContent[1])
	}
	return result, nil
}
