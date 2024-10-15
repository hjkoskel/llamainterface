package main

import (
	"llamainterface"
)

/*
type Artist struct {
	Formatter llamainterface.PromptFormatter
	Settings  GameArtistSettings
	//Pars bindstablediff.TextGenPars
}
*/

type GameArtistSettings struct {
	SystemPrompt string
	Examples     []PromptExample
	//Add some stuff to prompt
	Prefix string
	Suffix string
}

func (p *GameArtistSettings) GetLLMMEssages(prompt string) ([]llamainterface.LLMMessage, error) {
	data := []llamainterface.LLMMessage{
		{Type: llamainterface.LLMMESSAGETYPE_SYSTEM, Content: p.SystemPrompt},
	}

	for _, ex := range p.Examples {
		data = append(data, llamainterface.LLMMessage{Type: PLAYER, Content: ex.Input})
		data = append(data, llamainterface.LLMMessage{Type: DUNGEONMASTER, Content: ex.Result})
	}
	data = append(data, llamainterface.LLMMessage{Type: PLAYER, Content: prompt})
	data = append(data, llamainterface.LLMMessage{Type: DUNGEONMASTER, Content: ""}) //Waiting for that completion fills up this

	return data, nil //TODO error check?
}
