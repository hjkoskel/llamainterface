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
	"github.com/hjkoskel/bindstablediff"
)

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

// Writes game to disk markdown?
func (p *AdventureGame) SaveGame() error {
	var sb strings.Builder

	for _, page := range p.Pages {
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
	if len(p.Pages) == 0 {
		return fmt.Errorf("no pages")
	}
	last := p.Pages[len(p.Pages)-1]
	return CreatePngIfNotFound(*imGen, path.Join(p.GetSaveDir(), last.PictureFileName()), last.PictureDescription)
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
	messages = append(messages, llamainterface.LLMMessage{Type: DUNGEONMASTER})
	var errToPrompt error

	query.Prompt, errToPrompt = p.promptFormatter.ToPrompt(messages)
	if errToPrompt != nil {
		return errToPrompt
	}

	color.Blue(fmt.Sprintf("\nUSE PROMPT: %s\n", query.Prompt))

	postResp, errPost := p.llama.PostCompletion(query, nil, TIMEOUTCOMPLETION)
	if errPost != nil {
		return errPost
	}
	respText := postResp.Content

	color.Magenta(fmt.Sprintf("PROMPT RESPONSE IS %s\n", respText))
	msg := p.promptFormatter.Cleanup(respText)
	fmt.Printf("---after parsing---\n")
	/*
		respMessages, errRespParse := p.promptFormatter.Parse(respText)
		if errRespParse != nil {
			return errRespParse
		}
		fmt.Printf("Messages %v on response %#v\n", len(respMessages), respMessages)
		msg := respMessages[len(respMessages)-1]
		if msg.Type != DUNGEONMASTER {
			return fmt.Errorf("Internal error parsed response is not dungeon master %#v", respMessages)
		}*/

	nextPage := AdventurePage{
		Text:               msg,
		UserResponse:       "",
		Summary:            "",
		TokenCount:         0,  //TODO CALC LATER?
		PictureDescription: "", //TODO GET WITH LLM
		//PictureFileName    string //without path?  generate on fly
		Timestamp: time.Now(),
	}

	imGenPrompt, errImGenPrompt := p.Artist.GetLLMMEssages(nextPage.Text)
	if errImGenPrompt != nil {
		return fmt.Errorf("failed GetLLMessages for artist err:%s", errImGenPrompt)
	}

	//Lets generate text prompt. It is fast
	queryImPromptGen := gameBasicQuery(p.promptFormatter)
	queryImPromptGen.Temperature = temperature
	var errImGenMessage error

	queryImPromptGen.Prompt, errImGenMessage = p.promptFormatter.ToPrompt(imGenPrompt)
	if errImGenMessage != nil {
		return errImGenMessage
	}

	color.Cyan("\n----IMGEN prompt----\n%s\n----------\n", queryImPromptGen.Prompt)

	postImgResp, errImgPost := p.llama.PostCompletion(queryImPromptGen, nil, TIMEOUTCOMPLETION)
	if errImgPost != nil {
		return errImgPost
	}
	nextPage.PictureDescription = postImgResp.Content

	color.Cyan("\n----ART PROMPT----\n%s\n----------\n", nextPage.PictureDescription)

	p.Pages = append(p.Pages, nextPage)

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

func main() {
	pLoadFile := flag.String("load", "", "load existing game from json and continue")
	pBrancToNewGame := flag.Bool("b", false, "branch new game from loadfile")

	//Some simple settings
	pTempTextPrompt := flag.Float64("tp", 0.7, "temperature of text prompt")
	pTempImageTextPrompt := flag.Float64("tip", 0.7, "temperature of imagePrompt prompt")

	pTextMode := flag.Bool("tm", false, "use text mode, do not generate new in game pictures") //For faster gaming

	//TODO start from json  with -load at all times. These were just for bootstrapping
	//pGameFile := flag.String("f", "firetopadventure.txt", "uses this file as system prompt. This is what you want to play")
	//pArtistBaseTextFile := flag.String("ab", "promptArtist.txt", "prompt for generating image prompts from dungeon master text")
	//pMaxTokens := flag.Int("maxtokens", 1024, "number of max tokens for LLM model")

	//pLlamafile := flag.String("l", LLAMAFILENAME, ".llamafile filename, starts that file as server")
	pLlamafile := flag.String("l", "/usr/local/bin/llamafile", ".llamafile filename, starts that file as server. Or llama-server file (then model is needed)")
	//pLlamafile := flag.String("l", "", ".llamafile filename, starts that file as server")

	pLlamafileModel := flag.String("lm", "", "use alternative .gguf model file on this llamafile instead")
	//pLlamaport := flag.Int("lp", 0, "llamafile port, use 0 if let kernel decide port")
	pServerHost := flag.String("h", "127.0.0.1", "llama.cpp server host")
	//pServerPort := flag.Int("p", 8080, "llama.cpp server port")
	pllmPort := flag.Int("p", 8080, "llama.cpp or lllamafile server port. 0 for kernel decides for llamafile")

	pUiPort := flag.Int("uip", 2222, "web ui port")

	pDiffusionModelFile := flag.String("dmf", "/home/henri/aimallit/stable-diffusionMuunnetut/sd-v1-4-ggml-model-f16.bin", "diffusion model file")
	//pDiffusionModelFile := flag.String("dmf", "/home/henri/aimallit/stable-diffusionMuunnetut/HassanBlend1.5-ggml-model-q4_1.bin", "diffusion model file")

	pFluxHost := flag.String("fh", "127.0.0.1", "hostname of flux.1 server")
	pFluxPort := flag.Int("fp", 8800, "flux.1 server port")

	//For generating new adventures
	pCreateBlank := flag.String("blank", "", "Create blank game file as example from this for creating new adventure games")
	flag.Parse()

	if len(*pLlamafile) == 0 && *pllmPort == 0 {
		fmt.Printf("llama port must be defined when not starting llamafile from executable")
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

	//webUiRouter, errWebUiRouter := InitRouter()

	/*******************************************
	* Initialize model that generates pictures *
	********************************************/
	var imGen ImageGenerator
	if 0 < len(*pFluxHost) {
		var imGenErr error
		imGen, imGenErr = InitFluxImageGen(*pFluxHost, *pFluxPort)
		if imGenErr != nil {
			fmt.Printf("error initializing flux.1: %s\n", imGenErr)
			return
		}
	} else {
		var imGenErr error
		imGen, imGenErr = InitDiffusionImageGen(*pDiffusionModelFile, 15, -1, bindstablediff.DEFAULT)
		if imGenErr != nil {
			fmt.Printf("error initializing stable diffusion: %s\n", imGenErr)
			return
		}
	}

	/********************************
	* Generate missing media assets *
	*********************************/
	errCreateTitle := CreatePngIfNotFound(imGen, TITLEPICTUREFILE, MAINTITLESCREENGRAPHICPROMPT)
	if errCreateTitle != nil {
		fmt.Printf("\nerror while creating title picture %s\n", errCreateTitle)
		return
	}
	//Generate menu
	errCreateMenu := CreatePngIfNotFound(imGen, MENUPICTUREFILE, MAINMENUGRAPHICPROMPT)
	if errCreateMenu != nil {
		fmt.Printf("error creating menu pictur file %s\n", errCreateMenu)
	}
	//generate pictures for newgame

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
		errPrepare := CreatePngIfNotFound(imGen, catItem.TitleImageFileName, catItem.TitleDescription)
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
	ui, errUi := localUIStart()
	if errUi != nil {
		fmt.Printf("%s\n", errUi)
		return
	}

	/****************
	* Get local game
	*****************/
	var game AdventureGame
	game.promptFormatter = &llamainterface.Llama3InstructFormatter{System: "system", Users: []string{"player"}, Assistant: "dungeonmaster"} //TODO choose and support other formats

	if 0 < len(*pLoadFile) {
		// TODO PÄÄOHJELMAAN JOS KOMENTORIVI ARGUMENTILLA

		fmt.Printf("\n--- Loading:%s ---\n\n", *pLoadFile)
		var errGameInit error
		game, errGameInit = loadAdventure(*pLoadFile, llama)
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

	ui.SetGenerating("Starting up, wait....")
	ui.Render()

	localUiRunErr := localUIFlow(ui, game, *pTempTextPrompt, *pTempImageTextPrompt, &imGen)
	if localUiRunErr == nil {
		return
	}
	fmt.Printf("\n\n\nFAIL %s\n", localUiRunErr)
}
