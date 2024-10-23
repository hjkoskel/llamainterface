package main

import (
	"llamainterface"
	"strings"
)

// GameArtistSettings.  DEPRECATED! using just prefix and suffix for promt. Allows to inject fixed theme on game
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

const (
	PROMPT_GETSCENERY = "What kind of things, furnitures, textures, places , persons, objects and entities could are on scenery picture if it is was made from current situation. Describe what people look and what they are wearing. Also list where player is and what kind of lighting there is. Only respond with valid JSON array of object with type,name and description. Do not write an introduction or summary"
)

/*
for example get scenery produces
*/

type SceneryItem struct { //Hope that LLM coud provide list...
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (p *SceneryItem) AllAreSet() bool {
	return 0 < len(p.Type) && 0 < len(p.Name) && 0 < len(p.Description)
}

type SceneryOutput []SceneryItem //Hope that AI can generate valid JSON

type KeyValueStringPair struct {
	Key   string
	Value string
}

func findAllIndexes(str, substr string) []int {
	var indexes []int
	start := 0
	for {
		index := strings.Index(str[start:], substr)
		if index == -1 {
			break
		}
		indexes = append(indexes, start+index)
		start += index + len(substr)
	}
	return indexes
}

type KeyValueStringPairArr []KeyValueStringPair

func GetKeyValuePairs(code string) KeyValueStringPairArr { //For inva
	code = strings.ReplaceAll(code, "\": \"", "\":\"")
	code = strings.ReplaceAll(code, "\" : \"", "\":\"")
	code = strings.ReplaceAll(code, "\" :\"", "\":\"")

	centerPositions := findAllIndexes(code, "\":\"")

	result := []KeyValueStringPair{}
	for _, pos := range centerPositions {
		firstPart := code[0:pos]
		firstFirstIndex := strings.LastIndex(firstPart, "\"")
		if firstFirstIndex < 0 {
			continue
		}
		firstPart = firstPart[firstFirstIndex+1 : len(firstPart)]
		//last part
		lastPart := code[pos+3 : len(code)-1]
		lastlastIndex := strings.Index(lastPart, "\"")
		if lastlastIndex < 0 {
			continue
		}
		lastPart = lastPart[0:lastlastIndex]

		result = append(result, KeyValueStringPair{Key: firstPart, Value: lastPart})
	}

	return result

}

func (p *KeyValueStringPairArr) ToSceneryOutput() SceneryOutput {
	var result []SceneryItem

	var element SceneryItem
	for _, m := range *p {
		switch strings.ToUpper(m.Key) {
		case "DESCRIPTION":
			element.Description = m.Value
		case "NAME":
			element.Name = m.Value
		case "TYPE":
			element.Type = m.Value
		}
		if element.AllAreSet() {
			result = append(result, element)
			element = SceneryItem{}
		}
	}
	return result
}

// Convert so it is possible to use as artist prompt
func (a SceneryOutput) String() string {
	//Lets group and do not repeat type
	byType := make(map[string][]SceneryItem)
	for _, item := range a {
		key := item.Description
		if key == "" {
			key = "-" //fix if type is not defined
		}
		arr, haz := byType[key]
		if !haz {
			arr = make([]SceneryItem, 0)
		}
		byType[key] = append(arr, item)
	}
	//Now try combine string
	pieces := []string{}

	for typename, arr := range byType {
		if len(arr) == 1 {
			pieces = append(pieces, typename)
		} else {
			pieces = append(pieces, typename+"s")
		}
		for n, item := range arr {
			pieces = append(pieces, item.Name)
			pieces = append(pieces, item.Description)
			if n < len(arr)-1 {
				pieces = append(pieces, "and")
			}
		}
	}
	return strings.Join(pieces, " ")
}
