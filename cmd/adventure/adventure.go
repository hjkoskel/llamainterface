/*
Simple text based adventure game in golang

uses ollama's llamafile
*/
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"llamainterface"

	"github.com/fatih/color"
)

const SUMMARIZEPROMPT = "summarize what have happend so far, where player is located and details with current inventory and what deals or quests are still to be done and what persons are met or what clues are found"

//var promptFormatter PromptFormatter //Common formatter... TODO make switchable if using other models

const (
	LLAMAFILENAME = "/home/henri/Downloads/Meta-Llama-3-8B-Instruct.Q5_K_M.llamafile"
)

const (
	MAINTITLESCREENGRAPHICPROMPT = "movie poster with different kind of themes and heroes. Fantasy characters, astronauts, superheros, cowboys, ghosts, robots, dogs,grim reaper, cthulhu, star trek vulcan. Picture have large title text ADVENTURE"
	TITLEPICTUREFILE             = "title.png"
	MAINMENUGRAPHICPROMPT        = "ADVENTURE game main menu with option A:New Game B:Branch C:Continue Q:Quit"
	MENUPICTUREFILE              = "menu.png"
)

const (
	SYSTEM        = "system"
	PLAYER        = "player"
	DUNGEONMASTER = "dungeonmaster"
)

const (
	GAMESDIR            = "./games/"
	SAVEGAMEDIR         = "./savegames/"
	LATESTDUNGEONAMSTER = "latestDungeonMaster.txt"
)

const (
	PROMPTTYPE_SYS    PromptType = 0
	PROMPTTYPE_PLAYER PromptType = 1
	PROMPTTYPE_DM     PromptType = 2
)

type PromptType int

type PromptEntry struct {
	Text       string
	TokenCount int //Tokenizing is cheap and depends on model. Better NOT TO SAVE tokens
	//Tokens             []int //TODO use tokens when querying?
	Type               PromptType
	PictureDescription string
	PictureFileName    string
	Timestamp          time.Time //Nice to have information
}

/*
TODO HUOM!!! PROMPT ENTRY JUTUSSA EI SAA OLLA MUOTOILUJUTTUJA

	func (p *PromptEntry) ToMarkdown(promptFormatter PromptFormatter) string {
		switch p.Type {
		case PROMPTTYPE_SYS:
			return fmt.Sprintf("# System prompt\n%s", p.Text)
		case PROMPTTYPE_PLAYER:
			s := promptFormatter.ExtractText(PLAYER, DUNGEONMASTER, p.Text)
			return fmt.Sprintf("\n~~~\n%s\n~~~\n", s)
		case PROMPTTYPE_DM:
			return fmt.Sprintf("![%s](%s)", strings.TrimSpace(strings.ReplaceAll(p.PictureDescription, "\n", " ")), path.Base(p.PictureFileName)) + "\n" + p.Text + "\n"
		}
		return p.Text
	}
*/

/*
func (p *PromptEntry) ToMarkdown() string {
	switch p.Type {
	case PROMPTTYPE_SYS:
		return fmt.Sprintf("# System prompt\n%s", p.Text)
	case PROMPTTYPE_PLAYER:
		s := p.Text
		return fmt.Sprintf("\n~~~\n%s\n~~~\n", s)
	case PROMPTTYPE_DM:
		return fmt.Sprintf("![%s](%s)", strings.TrimSpace(strings.ReplaceAll(p.PictureDescription, "\n", " ")), path.Base(p.PictureFileName)) + "\n" + p.Text + "\n"
	}
	return p.Text
}*/

func (p *AdventureGame) WriteBlankFile(fname string) error {
	ref := AdventureGame{
		Introprompt: p.Introprompt,
		MaxTokens:   p.MaxTokens,
		//Artist:      GameArtistSettings{BaseText: p.Artist.BaseText},
		Artist:   p.Artist,
		GameName: p.GameName,
	}
	byt, errMarshal := json.MarshalIndent(ref, "", " ")
	if errMarshal != nil {
		return errMarshal
	}
	return os.WriteFile(fname, byt, 0666)
}

const MDFILENOTIFICATION = "This is game printout made by [LLM adventure game prototype, https://github.com/hjkoskel/llamainterface/tree/main/cmd/adventure](https://github.com/hjkoskel/llamainterface/tree/main/cmd/adventure)"

// Writes game to disk markdown?
func (p *AdventureGame) SaveGame() error {
	var sb strings.Builder
	sb.WriteString(MDFILENOTIFICATION + "\n\n# " + p.GameName + "\n" + p.Introprompt + "\n")

	for i, page := range p.Pages {
		sb.WriteString(fmt.Sprintf("\n# %v\n", i+1))
		sb.WriteString(page.ToMarkdown())
	}

	d := p.GetSaveDir()
	os.MkdirAll(d, 0777)

	errWriteMarkdown := os.WriteFile(path.Join(d, p.GameName+".md"), []byte(sb.String()), 0666)
	if errWriteMarkdown != nil {
		return fmt.Errorf("error writing write markdown %s", errWriteMarkdown)
	}

	byt, errMarsh := json.MarshalIndent(*p, "", " ")
	if errMarsh != nil {
		return errMarsh
	}
	return os.WriteFile(path.Join(d, p.GameName+".json"), byt, 0666)
}

func (p *AdventureGame) GenerateLatestPicture(temperature float64, imGen *ImageGenerator) error {
	if p.Textmode {
		return nil
	}
	if len(p.Pages) == 0 {
		return fmt.Errorf("no pages")
	}
	last := p.Pages[len(p.Pages)-1]
	tImageCreateStart := time.Now() //aftertought TODO REFACTOR
	errCreate := CreatePngIfNotFound(*imGen, path.Join(p.GetSaveDir(), last.PictureFileName()), last.PictureDescription, "")
	p.Pages[len(p.Pages)-1].GenerationTimes.Picture = int(time.Since(tImageCreateStart).Milliseconds())
	return errCreate
}

func (p *AdventureGame) CalcImagePrompt(inputText string, temperature float64) (string, error) {
	//---- TODO add separate...
	imGenPrompt, errImGenPrompt := p.Artist.GetLLMMEssages(inputText)
	if errImGenPrompt != nil {
		return "", fmt.Errorf("failed GetLLMessages for artist err:%s", errImGenPrompt)
	}

	//Lets generate text prompt. Even when playing text mode, better to have at least for debug and prompt text development
	queryImPromptGen := gameBasicQuery(p.promptFormatter)
	queryImPromptGen.Temperature = temperature
	var errImGenMessage error

	queryImPromptGen.Prompt, errImGenMessage = p.promptFormatter.ToPrompt(imGenPrompt)
	if errImGenMessage != nil {
		return "", errImGenMessage
	}

	postImgResp, errImgPost := p.llama.PostCompletion(queryImPromptGen, nil, TIMEOUTCOMPLETION)
	if errImgPost != nil {
		return "", errImgPost
	}

	return p.promptFormatter.Cleanup(postImgResp.Content), nil
}

func (p *AdventureGame) CalcAllImagePrompts(temperature float64) error {
	for i, page := range p.Pages {
		tCalcStart := time.Now() //Added as aftertought TODO refactor
		picDesc, errCalcImagePrompt := p.CalcImagePrompt(page.Text, temperature)
		if errCalcImagePrompt != nil {
			return errCalcImagePrompt
		}
		p.Pages[i].GenerationTimes.PicturePrompt = int(time.Since(tCalcStart).Milliseconds())
		p.Pages[i].PictureDescription = picDesc
	}
	return nil
}

func (p *AdventureGame) ExperimentRunPromptOnLatest(textInput string, temperature float64) (string, string, error) {
	if len(p.Pages) == 0 {
		return "", "", fmt.Errorf("no pages")
	}

	oldValue := p.Pages[len(p.Pages)-1].UserResponse //empty...
	p.Pages[len(p.Pages)-1].UserResponse = textInput

	messages, errMessages := p.GetLLMMEssages() //Gets messages, token count not going over
	if errMessages != nil {
		return "", "", fmt.Errorf("error getting messages %s", errMessages)
	}

	p.Pages[len(p.Pages)-1].UserResponse = oldValue //recover

	promptMessages := append(messages, llamainterface.LLMMessage{Type: DUNGEONMASTER})

	query := gameBasicQuery(p.promptFormatter)
	query.Temperature = temperature

	var errToPrompt error

	query.Prompt, errToPrompt = p.promptFormatter.ToPrompt(promptMessages)
	if errToPrompt != nil {
		return query.Prompt, "", fmt.Errorf("to promt fail %s", errToPrompt)
	}

	postResp, errPost := p.llama.PostCompletion(query, nil, TIMEOUTCOMPLETION)
	if errPost != nil {
		return query.Prompt, "", errPost
	}
	return query.Prompt, p.promptFormatter.Cleanup(postResp.Content), nil
}

func cleanupForJSONlist(s string) string {
	s = strings.ReplaceAll(s, "`", "")
	startIndex := strings.Index(s, "[")
	endIndex := strings.LastIndex(s, "]")

	if startIndex < 0 || endIndex < 0 {
		return s
	}
	if endIndex <= startIndex {
		return s
	}
	return s[startIndex : endIndex+1]
}

/*
One mega function, advance game state.
*/
func (p *AdventureGame) UserInteraction(temperature float64, temperaturePicture float64, input string) error {
	query := gameBasicQuery(p.promptFormatter)
	query.Temperature = temperature

	//Last entry?
	/*
		if 0 < len(p.Pages[len(p.Pages)-1].UserResponse) {
			return fmt.Errorf("Internal error game already have input %s on page %d", len(p.Pages)-1, p.Pages[len(p.Pages)-1].UserResponse)
		}*/

	input = strings.TrimSpace(input)
	if len(input) == 0 && 0 < len(p.Pages) {
		return fmt.Errorf("empty input to existing game")
	}

	if 0 < len(p.Pages) {
		p.Pages[len(p.Pages)-1].UserResponse = input
	}

	fmt.Printf("------- PAGES ARE -------\n%#v\n--------------\n", p.Pages)

	messages, errMessages := p.GetLLMMEssages() //Gets messages, token count not going over
	if errMessages != nil {
		return fmt.Errorf("error getting messages %s", errMessages)
	}

	fmt.Printf("\n\nUSER INTERACTION MESSAGES : %#v\n\n", messages)
	promptMessages := append(messages, llamainterface.LLMMessage{Type: DUNGEONMASTER})
	var errToPrompt error

	query.Prompt, errToPrompt = p.promptFormatter.ToPrompt(promptMessages)
	if errToPrompt != nil {
		return errToPrompt
	}

	color.Blue(fmt.Sprintf("\nUSE PROMPT: %s\n", query.Prompt))

	tCompletionStart := time.Now()
	postResp, errPost := p.llama.PostCompletion(query, nil, TIMEOUTCOMPLETION)
	if errPost != nil {
		return errPost
	}
	respText := postResp.Content

	color.Magenta(fmt.Sprintf("PROMPT RESPONSE IS %s\n", respText))
	msg := p.promptFormatter.Cleanup(respText)
	fmt.Printf("---after parsing---\n")

	nextPage := AdventurePage{
		Text:               msg,
		UserResponse:       "",
		Summary:            "",
		TokenCount:         0,  //TODO CALC LATER?
		PictureDescription: "", //TODO GET WITH LLM
		//PictureFileName    string //without path?  generate on fly
		Timestamp: time.Now(),
	}
	nextPage.GenerationTimes.MainLLM = int(time.Since(tCompletionStart).Milliseconds())

	//---- TODO add separate...
	/*
		TODO DEPRECATE CalcImagePrompt

		var errCalcImagePrompt error
		tStartImageCalc := time.Now() //TODO added as aftertough Think better, refactor
		nextPage.PictureDescription, errCalcImagePrompt = p.CalcImagePrompt(nextPage.Text, temperature)
		if errCalcImagePrompt != nil {
			return errCalcImagePrompt
		}
		nextPage.GenerationTimes.PicturePrompt = int(time.Since(tStartImageCalc).Milliseconds())
		color.Cyan("\n----ART PROMPT----\n%s\n----------\n", nextPage.PictureDescription)
	*/

	//Create summary
	querySummary := gameBasicQuery(p.promptFormatter)
	querySummary.Temperature = temperature
	var errGenSummaryPrompt error

	summarizeMessages := append(messages,
		llamainterface.LLMMessage{Type: DUNGEONMASTER})
	if len(messages) <= 2 {
		summarizeMessages = append(messages,
			llamainterface.LLMMessage{Type: PLAYER, Content: SUMMARIZEPROMPT},
			llamainterface.LLMMessage{Type: DUNGEONMASTER})
	} else {
		summarizeMessages[len(summarizeMessages)-2].Content = SUMMARIZEPROMPT //Instead of user
	}

	querySummary.Prompt, errGenSummaryPrompt = p.promptFormatter.ToPrompt(summarizeMessages)
	if errGenSummaryPrompt != nil {
		return errGenSummaryPrompt
	}

	color.Green("\n---SUMMARY PROMPT----\n%s\n--------------\n\n", querySummary.Prompt)

	tSummaryCompletionStart := time.Now()
	postSummaryResp, errSummaryPost := p.llama.PostCompletion(querySummary, nil, TIMEOUTCOMPLETION)
	if errSummaryPost != nil {
		return errSummaryPost
	}
	nextPage.Summary = p.promptFormatter.Cleanup(postSummaryResp.Content)
	nextPage.GenerationTimes.Summary = int(time.Since(tSummaryCompletionStart).Milliseconds())
	p.Pages = append(p.Pages, nextPage)

	var errCalcImagePrompt error
	tStartImageCalc := time.Now() //TODO added as aftertough Think better, refactor

	//HACK! todo refactor this art thing to somewhere else!
	_, responseSceneJson, errRespScene := p.ExperimentRunPromptOnLatest(PROMPT_GETSCENERY, temperature)
	if errRespScene != nil {
		return errRespScene
	}

	//responseSceneJson = cleanupForJSONlist(responseSceneJson)
	nextPage.SceneSummary = responseSceneJson

	color.Yellow("\n----GOT SCENERY----\n%s\n----------\n", nextPage.SceneSummary)

	sceneKeyValues := GetKeyValuePairs(nextPage.SceneSummary)

	scene := sceneKeyValues.ToSceneryOutput() //Parse even hallunizated json

	if len(scene) == 0 {
		fmt.Printf("Scene parsing FAIL, fallback")
		//Try something! Fallback to old
		nextPage.PictureDescription, errCalcImagePrompt = p.CalcImagePrompt(nextPage.Text, temperature)
		if errCalcImagePrompt != nil {
			return errCalcImagePrompt
		}
	} else {
		nextPage.PictureDescription = strings.TrimSpace(p.Artist.Prefix + " " + scene.String() + " " + p.Artist.Suffix)
	}
	nextPage.GenerationTimes.PicturePrompt = int(time.Since(tStartImageCalc).Milliseconds())

	color.Cyan("\n----ART PROMPT----\n%s\n----------\n", nextPage.PictureDescription)
	p.Pages[len(p.Pages)-1] = nextPage //RE set  TODO HACK!  REFACTOR when know that this is good idea to generate prompts for art

	return p.SaveGame()
}

func gameBasicQuery(promptFormatter llamainterface.PromptFormatter) llamainterface.QueryCompletion {
	query := llamainterface.DefaultQueryCompletion()

	query.Stream = false
	query.N_predict = 1024
	query.Temperature = 0.7
	query.Stop = promptFormatter.QueryStop() //[]string{"<|eot_id|>"}
	//query.Stop = []string{"</s>", "dungeonmaster:", "player:"}
	query.Repeat_last_n = 256
	query.Repeat_penalty = 1.18
	query.Top_k = 40
	query.Top_p = 0.95
	//MISSING? query.Min_p = 0.05
	query.Tfs_z = 1
	query.Typical_p = 1
	query.Presence_penalty = 0
	query.Frequency_penalty = 0
	query.Mirostat = 0
	query.Mirostat_tau = 5
	query.Mirostat_eta = 0.1
	//query.grammar="",
	query.N_probs = 0

	return query
}

//var artist Artist

func InitLLM(llamafilename string, llamaport int, llamamodelfile string, serverHost string) (*llamainterface.LLamaServer, error) {
	ctxLllama, _ := context.WithCancel(context.Background())
	var llamacmdFlags llamainterface.ServerCommandLineFlags
	llamacmdFlags.ToDefaults()

	var llama llamainterface.LLamaServer
	if 0 < len(llamafilename) {
		llamacmdFlags.ListenPort = llamaport

		if 0 < len(llamamodelfile) {
			llamacmdFlags.ModelFilename = llamamodelfile
		}
		var errLlamaInit error
		llama, errLlamaInit = llamainterface.InitLlamafileServer(ctxLllama, llamafilename, llamacmdFlags)
		if errLlamaInit != nil {
			return nil, fmt.Errorf("llama init fail %s\n", errLlamaInit)
		}
	} else {
		var errSrv error
		llama, errSrv = llamainterface.InitLLamaServer(serverHost, llamaport)
		if errSrv != nil {
			return nil, fmt.Errorf("err %s\n", errSrv)
		}
	}

	healthStatus, _ := llama.GetHealth()
	waitStarted := time.Now()
	for healthStatus != "ok" {
		var errhealth error
		healthStatus, errhealth = llama.GetHealth()
		if errhealth != nil {
			fmt.Printf("llama err...%s\n", errhealth.Error())
		} else {
			fmt.Printf("waiting llama %s\n", healthStatus)
		}
		if time.Second*600 < time.Since(waitStarted) {
			return nil, fmt.Errorf("timeout waiting llama health going to ok")
		}

		time.Sleep(time.Second)
	}
	return &llama, nil
}

func writePromptResults(promptListFileName string, outputfileName string, g *AdventureGame, temperature float64) error {
	byt, errRead := os.ReadFile(promptListFileName)
	if errRead != nil {
		return errRead
	}
	rows := strings.Split(string(byt), "\n")

	var sb strings.Builder //Build markdown
	for runIndex, row := range rows {
		row = strings.TrimSpace(row)
		if len(row) == 0 {
			continue
		}
		if strings.HasPrefix(row, "#") { //Comment out
			continue
		}
		completeQuery, response, errResp := g.ExperimentRunPromptOnLatest(row, temperature)
		if errResp != nil {
			return errResp
		}
		sb.WriteString(fmt.Sprintf("# query%v\n", runIndex))
		sb.WriteString(row + "\n")
		sb.WriteString("~~~ \n")
		sb.WriteString(completeQuery)
		sb.WriteString("~~~ \n")
		sb.WriteString(response + "\n")
	}
	return os.WriteFile(outputfileName, []byte(sb.String()), 0666)
}

var textPromptFormatter llamainterface.PromptFormatter

func main() {
	pLoadFile := flag.String("load", "", "load existing game from json and continue")
	pBrancToNewGame := flag.Bool("b", false, "branch new game from loadfile")

	//Some simple settings
	pTempTextPrompt := flag.Float64("tp", 0.7, "temperature of text prompt")
	pTempImageTextPrompt := flag.Float64("tip", 0.7, "temperature of imagePrompt prompt")

	pTextMode := flag.Bool("tm", false, "use text mode, do not generate new in game pictures") //For faster gaming
	pGenMissingImg := flag.Bool("gmi", false, "generate missing images on load or re-run creare picturepicture prompts when prompt is empty")
	pResetPicturePrompts := flag.Bool("rpp", false, "reset picture prompts from specific game")

	//LLM model
	pLlamafile := flag.String("l", "/usr/local/bin/llamafile", ".llamafile filename, starts that file as server. Or llama-server file (then model is needed)")
	pLlamafileModel := flag.String("lm", "", "use alternative .gguf model file on this llamafile instead")
	pServerHost := flag.String("h", "", "llama.cpp server host")
	pllmPort := flag.Int("p", 8080, "llama.cpp or lllamafile server port. 0 for kernel decides for llamafile")

	pPromptFormatterName := flag.String("pf", llamainterface.PROMPTFORMAT_LLAMA31, fmt.Sprintf("formatter name:%v", llamainterface.AllowedPromptFormatterNames()))
	//UI settings
	pUiPort := flag.Int("uip", 0, "web ui port")

	pImageGenEndpointUrl := flag.String("igurl", "http://127.0.0.1:8800/txt2img", "image generator URL")
	pImagGenWidth := flag.Int("igw", 1024, "image generation X-resolution")
	pImagGenHeight := flag.Int("igh", 1024, "image generation Y-resolution")
	pImagGenSampleSteps := flag.Int("igsteps", 20, "image generation steps, affects speed vs quality")
	pImagGenSampler := flag.String("igsampler", "euler", fmt.Sprintf("choose sampling method [%s]", strings.Join(ListAllSamplingMethods(), ",")))

	//For generating new adventures
	pCreateBlank := flag.String("blank", "", "Create blank game file as example from this for creating new adventure games")
	pRunOnLatest := flag.String("rol", "", "filename of text file, one query per row. Run on latest and produce out_filename as result. Used for development")

	flag.Parse()

	if len(*pLlamafile) == 0 && *pllmPort == 0 {
		fmt.Printf("llama port must be defined when not starting llamafile from executable")
		return
	}

	var errFormatterName error
	textPromptFormatter, errFormatterName = llamainterface.ParsePromptFormatter(*pPromptFormatterName, SYSTEM, []string{PLAYER}, DUNGEONMASTER)
	if errFormatterName != nil {
		fmt.Printf("formatter name error:%s\n", errFormatterName)
		return
	}

	if *pllmPort == 0 {
		llamaport, errGetPort := llamainterface.GetFreePort()
		if errGetPort != nil {
			fmt.Printf("getting free port failed %s\n", errGetPort)
			return
		}
		pllmPort = &llamaport
	}

	/*******************************************
	* Initialize model that generates pictures *
	********************************************/
	var imGen ImageGenerator

	imageGenPars := DefaultStableDiffusionCppParameters()

	imageGenPars.Width = *pImagGenWidth //TODO VALIDATE valid resolution
	imageGenPars.Height = *pImagGenHeight
	imageGenPars.SampleSteps = *pImagGenSampleSteps

	errMethod := imageGenPars.SetSampleMethod(*pImagGenSampler)
	if errMethod != nil {
		fmt.Printf("invalid sampling method %s\n", errMethod)
		return
	}

	fluxgen := FluxImageGenerator{Url: *pImageGenEndpointUrl, Parameters: imageGenPars}
	imGen = &fluxgen

	/********************************
	* Generate missing media assets *
	*********************************/
	errCreateTitle := CreatePngIfNotFound(imGen, TITLEPICTUREFILE, MAINTITLESCREENGRAPHICPROMPT, "")
	if errCreateTitle != nil {
		fmt.Printf("\nerror while creating title picture %s\n", errCreateTitle)
		return
	}
	//Generate menu
	errCreateMenu := CreatePngIfNotFound(imGen, MENUPICTUREFILE, MAINMENUGRAPHICPROMPT, "")
	if errCreateMenu != nil {
		fmt.Printf("error creating menu pictur file %s\n", errCreateMenu)
	}
	/************************
	** Initialize LLM model
	*************************/

	llama, errllama := InitLLM(*pLlamafile, *pllmPort, *pLlamafileModel, *pServerHost)
	if errllama != nil {
		fmt.Printf("error starting llm %s\n", errllama)
		return
	}

	jsonNewGameList, errJsonNewGameList := ListNewGamesJson(GAMESDIR, llama)
	if errJsonNewGameList != nil {
		fmt.Printf("ERROR LISTING NEW GAMES %s\n", errJsonNewGameList)
		return
	}

	cat, catErr := CreateToCatalogue(jsonNewGameList, llama)

	if catErr != nil {
		fmt.Printf("FAILED LISTING GAME CATALOG %s\n", catErr)
		return
	}
	for _, catItem := range cat {
		fmt.Printf("Generating start image %s\n", catItem.TitleImageFileName)
		errPrepare := CreatePngIfNotFound(imGen, catItem.TitleImageFileName, catItem.TitleDescription, "")
		if errPrepare != nil {
			fmt.Printf("failed preparing %s\n", errPrepare)
			continue
		}
	}

	/*******
	* start web UI if wanted in server mode
	*******/
	if 0 < *pUiPort {
		errWebRun := RunAsWebServer(*pUiPort, &imGen, llama, *pTempTextPrompt, *pTempImageTextPrompt)
		fmt.Printf("WEB UI FAILED WITH ERROR %s\n", errWebRun)
		return
	}

	/*****************
	* Start local ui *
	******************/
	fmt.Printf("\n\n\n--STARTING UI--\n\n")
	var ui *GraphicalUI
	if len(*pRunOnLatest) == 0 || len(*pLoadFile) == 0 {
		var errUi error
		ui, errUi = localUIStart()
		if errUi != nil {
			fmt.Printf("%s\n", errUi)
			return
		}
	}

	/****************
	* Get local game
	*****************/
	var game AdventureGame

	game.promptFormatter = textPromptFormatter

	if 0 < len(*pLoadFile) {
		// TODO PÄÄOHJELMAAN JOS KOMENTORIVI ARGUMENTILLA

		fmt.Printf("\n--- Loading:%s ---\n\n", *pLoadFile)
		var errGameInit error
		game, errGameInit = loadAdventure(*pLoadFile, llama, &textPromptFormatter)
		if errGameInit != nil {
			fmt.Printf("err game load from %s err:%s\n", *pLoadFile, errGameInit)
			return
		}
		if *pBrancToNewGame {
			game.StartTime = time.UnixMilli(0)
		}
	} else {
		var errGetGame error
		game, errGetGame = localUIGetGame(ui, imGen, llama)
		if errGetGame != nil {
			fmt.Printf("%s\n", errGetGame)
			return
		}
		//Not sure is this needed?
		if 0 < len(*pCreateBlank) {
			errWriteBlank := game.WriteBlankFile(*pCreateBlank)
			if errWriteBlank != nil {
				fmt.Printf("error writing blank file %s  err:%s", *pCreateBlank, errWriteBlank)
				return
			}
			fmt.Printf("\n\n--Exported %s\n", *pCreateBlank)
			return
		}
	}

	game.llama = llama
	game.Textmode = *pTextMode

	if 0 < len(*pRunOnLatest) {
		fOutName := "out_" + *pRunOnLatest + ".md"
		errWrite := writePromptResults(*pRunOnLatest, fOutName, &game, *pTempTextPrompt)
		fmt.Printf("\n\n---- RUNNED ON LATEST (file:%s) ----\n", *pRunOnLatest)
		if errWrite != nil {
			fmt.Printf("FAIL %s\n", errWrite)
			return
		}
		fmt.Printf("Wrote responses to %s\n\n", fOutName)
		return
	}

	if *pResetPicturePrompts {
		errReset := game.ResetPicturePrompts()
		if errReset != nil {
			fmt.Printf("failed resetting picture prompts err:%s\n", errReset)
		}
		game.RemovePicturesNotInPrompts()

		errImagePromptError := game.CalcAllImagePrompts(*pTempTextPrompt)
		if errImagePromptError != nil {
			fmt.Printf("error calculating image prompts %s\n", errImagePromptError)
			return
		}
		game.SaveGame()
	}

	if *pGenMissingImg {
		missingPics := game.ListMissingPictures()
		doneCounter := 0
		for pictureFilename, imageprompt := range missingPics {
			ui.SetGenerating(fmt.Sprintf("Generating missing pictures %v/%v : %s", doneCounter, len(missingPics), pictureFilename))
			ui.Render()
			errPrepare := CreatePngIfNotFound(imGen, pictureFilename, imageprompt, "")
			if errPrepare != nil {
				fmt.Printf("error generating picture %s\n", errPrepare)
				return
			}
			doneCounter++
		}
	}

	ui.SetGenerating("Starting up, wait....")
	ui.Render()

	localUiRunErr := localUIFlow(ui, game, *pTempTextPrompt, *pTempImageTextPrompt, &imGen)
	if localUiRunErr == nil {
		return
	}
	fmt.Printf("\n\n\nFAIL %s\n", localUiRunErr)
}
