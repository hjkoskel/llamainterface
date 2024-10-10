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

/*
func InitArtist(artSetting GameArtistSettings, formatter llamainterface.PromptFormatter) (Artist, error) {
	return Artist{
		Settings:  artSetting,
		Formatter: formatter}, nil
}

func (p *Artist) ArtPromptText(longtext string) string {
	data := []llamainterface.LLMMessage{
		{Type: llamainterface.LLMMESSAGETYPE_SYSTEM, Content: p.Settings.SystemPrompt},
	}

	for _, ex := range p.Settings.Examples {
		data = append(data, llamainterface.LLMMessage{Type: llamainterface.LLMMESSAGETYPE_USER, Content: ex.Input})
		data = append(data, llamainterface.LLMMessage{Type: llamainterface.LLMMESSAGETYPE_ASSISTANT, Content: ex.Result})
	}

	s, _ := p.Formatter.ToPrompt(data)
	return s
}
*/

/*
func (p *Artist) CreatePic(prompt string) (image.Image, error) {
	return p.Gen.CreatePic(strings.TrimSpace(p.Settings.Prefix + " " + strings.TrimSpace(prompt) + " " + p.Settings.Suffix))
}
*/
