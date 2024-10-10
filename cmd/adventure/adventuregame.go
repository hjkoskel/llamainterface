package main

import (
	"encoding/json"
	"fmt"
	"llamainterface"
	"os"
	"path"
	"strings"
	"time"
)

type AdventurePage struct {
	Text         string
	UserResponse string
	Summary      string //Summarizes what have happend since start

	TokenCount int //Tokenizing is cheap and depends on model. Better NOT TO SAVE tokens, omitempty

	PictureDescription string
	//PictureFileName    string //without path?  generate on fly
	Timestamp time.Time //Nice to have information, used for generating pictures
}

func (p *AdventurePage) ToMarkdown() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("![%s](%s)", strings.TrimSpace(strings.ReplaceAll(p.PictureDescription, "\n", " ")), path.Base(p.PictureFileName())))
	sb.WriteString("\n" + p.Text + "\n")

	if 0 < len(p.UserResponse) {
		sb.WriteString(fmt.Sprintf("\n~~~\n%s\n~~~\n---------------------------\n", p.UserResponse))
	}
	return sb.String()
}

// Just filename without path
func (p *AdventurePage) PictureFileName() string {
	return fmt.Sprintf("%v.png", p.Timestamp.Unix()) //Not suitable for serving huge number of games... or just use random ID?
}

func (p *AdventurePage) ToLLMMessages() []llamainterface.LLMMessage {
	result := []llamainterface.LLMMessage{
		llamainterface.LLMMessage{Type: DUNGEONMASTER, Content: p.Text},
	}
	if 0 < len(p.UserResponse) {
		result = append(result, llamainterface.LLMMessage{Type: PLAYER, Content: p.UserResponse})
	}
	return result
}

type AdventureGame struct {
	TitleGraphicPrompt string
	Introprompt        string
	MaxTokens          int //TODO changed by game engine if model have limitations. Or warning is given that token count is not enough

	StartTime time.Time //Used for identifying gamedir
	GameName  string    //Combination of fixed text and timestamp
	//Things added when saving game
	Artist GameArtistSettings //promptArtist.txt
	Pages  []AdventurePage

	Textmode bool

	//Internal state, not from json
	gameTitlepictureFilename string
	promptFormatter          llamainterface.PromptFormatter // PromptFormatter
	introPromptTokens        int
	llama                    *llamainterface.LLamaServer
}

const TIMEOUTTOKENIZE time.Duration = time.Second * 120
const TIMEOUTCOMPLETION time.Duration = time.Minute * 15

func (p *AdventureGame) getIntroPromptMessages() []llamainterface.LLMMessage {
	return llamainterface.LLMMessages{
		{Type: llamainterface.LLMMESSAGETYPE_SYSTEM, Content: p.Introprompt},
		{Type: PLAYER, Content: INITIALPROMPT}, //Hidden&hard coded or taken from game configuration?
	}
}

func (p *AdventureGame) GetLLMMEssages() ([]llamainterface.LLMMessage, error) {

	result := []llamainterface.LLMMessage{}

	errInternalCalc := p.CalcTokensMissing()
	if errInternalCalc != nil {
		return nil, errInternalCalc
	}

	tokensLeft := p.MaxTokens - p.introPromptTokens
	fmt.Printf("Tokens left initially %v\n", tokensLeft)
	if tokensLeft < 1 {
		return nil, fmt.Errorf("internal error, intro prompt have %v tokens, whole maxTokens is %v", p.introPromptTokens, p.MaxTokens)
	}
	if len(p.Pages) == 0 {
		return p.getIntroPromptMessages(), nil
	}
	fmt.Printf("Got %v pages\n", len(p.Pages))
	//Lets calculate where start and where get summary. TODO ADD SUMMARY FEATURE LATER. DROP AT FIRST!
	for i := len(p.Pages) - 1; 0 <= i; i-- {
		tokensLeft -= p.Pages[i].TokenCount
		if tokensLeft < 0 {
			return result, nil //Not taking this anymore  TODO TAKE SUMMARY IN ACCOUNT!?
		}

		newMessages := p.Pages[i].ToLLMMessages()
		fmt.Printf("\npage%v: %#v\n\n", i, newMessages)
		result = append(newMessages, result...)
	}
	result = append(p.getIntroPromptMessages(), result...)

	return result, nil
}

// Calculates tokens with 0 entry
func (p *AdventureGame) CalcTokensMissing() error {
	for i, page := range p.Pages {
		if page.TokenCount == 0 {
			messages := page.ToLLMMessages()
			messages = append(messages, llamainterface.LLMMessage{Type: DUNGEONMASTER}) //Waiting for fill
			fmt.Printf("\nCalc tokens on messages %#v\n\n", messages)

			prompt, formatErr := p.promptFormatter.ToPrompt(messages)
			if formatErr != nil {
				return fmt.Errorf("prompt format err %s on page %v", formatErr, i)
			}
			tokenList, errToken := p.llama.PostTokenize(prompt, TIMEOUTTOKENIZE)
			if errToken != nil {
				return fmt.Errorf("error tokenizing %s", errToken)
			}
			page.TokenCount = len(tokenList)
			p.Pages[i] = page
		}
	}
	return nil
}

// loadAdventure Continues existing adventure... OR starts from empty?
func loadAdventure(loadGameFile string, llama *llamainterface.LLamaServer) (AdventureGame, error) {
	byt, readErr := os.ReadFile(loadGameFile)
	if readErr != nil {
		return AdventureGame{}, fmt.Errorf("error reading savegame %s  err:%s", loadGameFile, readErr)
	}
	var result AdventureGame
	errParse := json.Unmarshal(byt, &result)
	if errParse != nil {
		return result, fmt.Errorf("UNMARSHAL ERROR %s", errParse)
	}
	if result.StartTime.Unix() < 1000 {
		result.StartTime = time.Now()
	}
	result.promptFormatter = &llamainterface.Llama3InstructFormatter{System: "system", Users: []string{"player"}, Assistant: "dungeonmaster"} // Llama31Formatter{} //TODO:Not need to other formatters?
	result.llama = llama
	errTok := result.CalcTokensMissing() //Assuming that does not take long. Tokenization is fast
	if errTok != nil {
		return result, fmt.Errorf("error tokens missing %s", errTok)
	}
	promptMessages := result.getIntroPromptMessages()
	promptMessages = append(promptMessages, llamainterface.LLMMessage{Type: DUNGEONMASTER}) //Waiting for fill
	fmt.Printf("intro prompt messages are %#v\n", promptMessages)

	introPrompt, errIntroPrompt := result.promptFormatter.ToPrompt(promptMessages) //Intro prompt and initial prompt
	if errIntroPrompt != nil {
		return result, fmt.Errorf("err while converting to prompt %s", errIntroPrompt)
	}
	introTokens, errTokenizeIntro := result.llama.PostTokenize(introPrompt, TIMEOUTTOKENIZE)
	if errTokenizeIntro != nil {
		return result, fmt.Errorf("err while tokinizing %s", errTokenizeIntro)
	}
	result.introPromptTokens = len(introTokens)

	result.gameTitlepictureFilename = strings.Replace(loadGameFile, ".json", ".png", 1)
	if 0 < len(result.Pages) {
		result.gameTitlepictureFilename = path.Join(GAMESDIR, result.GameName+".png")
	}

	fmt.Printf("LOAD OK\n")
	return result, nil
}

func (p *AdventureGame) GameId() string {
	return fmt.Sprintf("%s%d", p.GameName, p.StartTime.Unix())
}

func (p *AdventureGame) GetSaveDir() string {
	return path.Join(SAVEGAMEDIR, p.GameId())
}

func (p *AdventureGame) GetMenuPictureFile() string {
	/*	if len(p.Pages) == 0 {
			return p.gameTitlepictureFilename
		}
	*/
	for i := len(p.Pages) - 1; 0 <= i; i-- {
		fname := path.Join(p.GetSaveDir(), p.Pages[i].PictureFileName())
		_, errLoad := LoadPng(fname)
		if errLoad == nil {
			return fname
		}
	}

	return p.gameTitlepictureFilename
}
