package llamainterface

import (
	"encoding/json"
	"time"
)

type TokenizationJob struct {
	Content string `json:"content"`
}

type DetokenizationJob struct {
	Tokens []int `json:"tokens"`
}

func (p *LLamaServer) PostTokenize(s string, timeout time.Duration) ([]int, error) {
	job := TokenizationJob{Content: s}

	byt, _ := json.Marshal(job)
	final, errPost := p.PostQuery("/tokenize", byt, nil, timeout)
	if errPost != nil {
		return []int{}, errPost
	}

	var result DetokenizationJob
	errParse := json.Unmarshal([]byte(final), &result)
	if errParse != nil {
		return []int{}, errParse
	}
	return result.Tokens, nil
}

func (p *LLamaServer) PostDetokenize(arr []int, timeout time.Duration) (string, error) {
	job := DetokenizationJob{Tokens: arr}
	byt, _ := json.Marshal(job)
	final, errPost := p.PostQuery("/detokenize", byt, nil, timeout)
	if errPost != nil {
		return "", errPost
	}

	var result TokenizationJob
	errParse := json.Unmarshal([]byte(final), &result)
	if errParse != nil {
		return "", errParse
	}
	return result.Content, nil
}
