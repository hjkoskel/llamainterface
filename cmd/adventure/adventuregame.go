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

type CalcStatistics struct { //Track performance of gaming system
	MainLLM       int //all are in milliseconds for json marshaling
	Summary       int
	PicturePrompt int
	Picture       int
	TextTTS       int
} //Instead of map, use struct. More controlled

// Contains strings for looking up...
type TranslatedPageData struct {
	Text         map[LanguageName]string
	UserResponse map[LanguageName]string
}

func (p *TranslatedPageData) ListLanguages() []LanguageName {
	r := make(map[LanguageName]bool)

	for k, _ := range p.Text {
		r[k] = true
	}
	for k, _ := range p.UserResponse {
		r[k] = true
	}

	result := []LanguageName{}
	for name, _ := range r {
		result = append(result, name)
	}
	return result
}

type AdventurePage struct {
	Text         string
	UserResponse string
	Summary      string //Summarizes what have happend since start

	SceneSummary string //SceneryOutput //Image prompt is based on this.  Store as text so it is possible to make "json cleaner" methods!

	TokenCount int //Tokenizing is cheap and depends on model. Better NOT TO SAVE tokens, omitempty

	PictureDescription string
	//PictureFileName    string //without path?  generate on fly
	Timestamp time.Time //Nice to have information, used for generating pictures

	GenerationTimes CalcStatistics

	Localization TranslatedPageData
}

func (p *AdventurePage) GetText(lang LanguageName) string { // if localized not found, return default
	if p.Localization.Text == nil {
		return p.Text
	}
	txt, haz := p.Localization.Text[lang]
	if !haz {
		return p.Text
	}
	return txt
}

func (p *AdventurePage) GetUserResponse(lang LanguageName) string {
	if p.Localization.UserResponse == nil {
		return p.UserResponse
	}
	txt, haz := p.Localization.UserResponse[lang]
	if !haz {
		return p.UserResponse
	}
	return txt
}

func (p *AdventurePage) ToMarkdown(lang LanguageName) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("![%s](%s)", strings.TrimSpace(strings.ReplaceAll(p.PictureDescription, "\n", " ")), path.Base(p.PictureFileName())))
	sb.WriteString("\n" + p.GetText(lang) + "\n")

	if 0 < len(p.UserResponse) {
		sb.WriteString(fmt.Sprintf("\n~~~\n%s\n~~~\n---------------------------\n", p.GetUserResponse(lang)))
	}
	return sb.String()
}

// Just filename without path
func (p *AdventurePage) PictureFileName() string {
	return fmt.Sprintf("%v.png", p.Timestamp.Unix()) //Not suitable for serving huge number of games... or just use random ID?
}

func (p *AdventurePage) TTSTextFileName(lang LanguageName) string {
	if len(lang) == 0 || lang == LANG_eng_Latn {
		return fmt.Sprintf("%v.wav", p.Timestamp.Unix())
	}
	return fmt.Sprintf("%v_%s.wav", p.Timestamp.Unix(), lang) //Not suitable for serving huge number of games... or just use random ID?
}

func (p *AdventurePage) TTSPromptFileName(lang LanguageName) string {
	if len(lang) == 0 || lang == LANG_eng_Latn {
		return fmt.Sprintf("%v_prompt.wav", p.Timestamp.Unix())
	}
	return fmt.Sprintf("%v_prompt_%s.wav", p.Timestamp.Unix(), lang) //Not suitable for serving huge number of games... or just use random ID?
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

	translator *Translator
	ttsengine  *TTSInterf
}

const TIMEOUTTOKENIZE time.Duration = time.Second * 120
const TIMEOUTCOMPLETION time.Duration = time.Minute * 15

func (p *AdventureGame) ResetPicturePrompts() error {
	for pagenumber, _ := range p.Pages {
		p.Pages[pagenumber].PictureDescription = ""
	}
	return p.SaveGame()
}

func (p *AdventureGame) ListMissingPicturePrompts() []int { //For calculating ammount
	result := []int{}
	for i, page := range p.Pages {
		if len(page.PictureDescription) == 0 {
			result = append(result, i)
		}
	}
	return result
}

func (p *AdventureGame) RemovePicturesNotInPrompts() {
	for _, page := range p.Pages {
		fname := path.Join(p.GetSaveDir(), page.PictureFileName())
		if page.PictureDescription == "" {
			os.Remove(fname)
		}
	}
}

// List filename and prompt from all pages.
func (p *AdventureGame) ListMissingPictures() map[string]string {
	result := make(map[string]string)

	for _, page := range p.Pages {
		fname := path.Join(p.GetSaveDir(), page.PictureFileName())
		if page.PictureDescription == "" {
			os.Remove(fname)
		}

		_, errLoad := LoadPng(fname)
		if errLoad != nil { //Got error
			result[fname] = page.PictureDescription
		}
	}
	return result
}

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
func loadAdventure(loadGameFile string, llama *llamainterface.LLamaServer, promptformatter *llamainterface.PromptFormatter, translator *Translator, ttsEngine *TTSInterf) (AdventureGame, error) {
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
	result.promptFormatter = *promptformatter // &llamainterface.Llama3InstructFormatter{System: "system", Users: []string{"player"}, Assistant: "dungeonmaster"}
	result.llama = llama
	result.translator = translator
	result.ttsengine = ttsEngine
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
	for i := len(p.Pages) - 1; 0 <= i; i-- {
		fname := path.Join(p.GetSaveDir(), p.Pages[i].PictureFileName())
		_, errLoad := LoadPng(fname)
		if errLoad == nil {
			return fname
		}
	}
	return p.gameTitlepictureFilename
}

func (p *AdventureGame) GenerateAllSpeeches() error {
	for pagenumber, page := range p.Pages {
		fmt.Printf("Running TTS on page %v/%v\n", pagenumber, len(p.Pages))
		txt := page.GetText(p.translator.Lang)
		a := *p.ttsengine
		errRun := a.Run(txt, path.Join(p.GetSaveDir(), page.TTSTextFileName(p.translator.Lang)))
		if errRun != nil {
			return errRun
		}
	}
	return nil
}

func (p *AdventureGame) TranslateMissing() error {
	if len(p.translator.Lang) == 0 {
		return nil
	}
	//First texts?
	for pagenumber, page := range p.Pages {
		if page.Localization.Text != nil {
			_, haz := page.Localization.Text[p.translator.Lang]
			if haz {
				continue
			}
		} else {
			p.Pages[pagenumber].Localization.Text = make(map[LanguageName]string)
		}
		//Need to translate
		translated, errTranslator := p.translator.FromEnglish(page.Text)
		if errTranslator != nil {
			return errTranslator
		}
		p.Pages[pagenumber].Localization.Text[p.translator.Lang] = translated
	}

	for pagenumber, page := range p.Pages {
		if page.Localization.UserResponse != nil {
			_, haz := page.Localization.UserResponse[p.translator.Lang]
			if haz {
				continue
			}
		} else {
			p.Pages[pagenumber].Localization.UserResponse = make(map[LanguageName]string)
		}
		//Need to translate
		translated, errTranslator := p.translator.FromEnglish(page.UserResponse)
		if errTranslator != nil {
			return errTranslator
		}
		p.Pages[pagenumber].Localization.UserResponse[p.translator.Lang] = translated
	}

	return nil
}
