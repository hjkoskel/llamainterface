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

const (
	LLAMAFILENAME = "/home/henri/Downloads/Meta-Llama-3-8B-Instruct.Q5_K_M.llamafile"
)

const (
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
	Text               string
	Tokens             []int //TODO use tokens when querying?
	Type               PromptType
	PictureDescription string
	PictureFileName    string
}

func (p *PromptEntry) ToMarkdown() string {
	switch p.Type {
	case PROMPTTYPE_SYS:
		return fmt.Sprintf("# System prompt\n%s", p.Text)
	case PROMPTTYPE_PLAYER:
		s := strings.ReplaceAll(p.Text, "<|start_header_id|>player<|end_header_id|>", "")
		s = strings.ReplaceAll(s, "<|eot_id|><|start_header_id|>dungeonmaster<|end_header_id|>", "")
		s = strings.TrimSpace(s)
		return fmt.Sprintf("\n~~~\n%s\n~~~\n", s)
	case PROMPTTYPE_DM:
		return fmt.Sprintf("![%s](%s)", strings.TrimSpace(strings.ReplaceAll(p.PictureDescription, "\n", " ")), path.Base(p.PictureFileName)) + "\n" + p.Text + "\n"
	}
	return p.Text
}

type GameArtistSettings struct {
	BaseText string
}

type AdventureGame struct {
	llama       llamainterface.LLamaServer
	Introprompt string

	MaxTokens int

	Artist GameArtistSettings //promptArtist.txt

	//totalPrompt   string
	PromptEntries []PromptEntry //Initial prompt is first, then DM, player,DM,player

	StartTime time.Time //Used for identifying gamedir
	GameName  string    //Combination of fixed text and timestamp
}

func (p *AdventureGame) WriteBlankFile(fname string) error {
	ref := AdventureGame{
		Introprompt: p.Introprompt,
		MaxTokens:   p.MaxTokens,
		Artist:      GameArtistSettings{BaseText: p.Artist.BaseText},
		GameName:    p.GameName,
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
	for _, entry := range p.PromptEntries {
		sb.WriteString(entry.ToMarkdown())
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

func (p *AdventureGame) GetTotalPrompt() string {
	//TODO calc tokens, pick first and then as much as possible...
	result := ""
	first := p.PromptEntries[0].Text

	count := len(p.PromptEntries[0].Tokens)

	for i := len(p.PromptEntries) - 1; 0 < i; i-- {
		fmt.Printf("i=%v len=%v\n", i, len(p.PromptEntries))
		count += len(p.PromptEntries[i].Tokens)
		fmt.Printf("count=%d maxtokens=%d\n", count, p.MaxTokens)
		if p.MaxTokens < count {
			return first + result
		}
		result = p.PromptEntries[i].Text + result
	}
	return first + result
}

func (p *AdventureGame) GetTotalTokens() []int { // llama.cpp server
	//TODO calc tokens, pick first and then as much as possible...
	result := []int{}
	first := p.PromptEntries[0].Tokens

	count := len(p.PromptEntries[0].Tokens)

	for i := len(p.PromptEntries) - 1; 0 < i; i-- {
		count += len(p.PromptEntries[i].Tokens)
		if p.MaxTokens < count {
			return append(first, result...)
		}
		result = append(p.PromptEntries[i].Tokens, result...)
	}
	return append(first, result...)
}

func (p *AdventureGame) latestEntryIndexForDM() int {
	for i := len(p.PromptEntries) - 1; 0 < i; i-- {
		if p.PromptEntries[i].Type == PROMPTTYPE_DM {
			return i
		}
	}
	return -1
}

// Stores latest text output on disk for editing. Way to hack if game is not progressing to right direction
func (p *AdventureGame) StoreLatestTextOutput() error {
	index := p.latestEntryIndexForDM()
	if index < 0 {
		return fmt.Errorf("no dungeon master outputs found in messages (len:%d)", len(p.PromptEntries))
	}
	return os.WriteFile(LATESTDUNGEONAMSTER, []byte(p.PromptEntries[index].Text), 0666)
}

// If user have edited output.. like fixing errors, pull that from file. Missing file is not error. Then just dont change
func (p *AdventureGame) CheckOutputChanges() (bool, error) {
	byt, readErr := os.ReadFile(LATESTDUNGEONAMSTER)
	if readErr != nil {
		fmt.Printf("was not able to load %s, continue without changes\n") //Is not error if user deleted
		return false, nil
	}

	index := p.latestEntryIndexForDM()
	if index < 0 {
		return false, fmt.Errorf("no dungeon master outputs found in messages (len:%d)", len(p.PromptEntries))
	}

	p.PromptEntries[index].Text = string(byt)
	return true, nil
}

// loadAdventure Continues existing adventure
func loadAdventure(loadGameFile string, llama llamainterface.LLamaServer) (AdventureGame, error) {
	byt, readErr := os.ReadFile(loadGameFile)
	if readErr != nil {
		return AdventureGame{}, fmt.Errorf("error reading savegame %s  err:%s", loadGameFile, readErr)
	}
	var result AdventureGame
	errParse := json.Unmarshal(byt, &result)
	result.llama = llama
	if result.StartTime.Unix() < 1000 {
		result.StartTime = time.Now()
	}
	return result, errParse
}

func initAdventure(gameName string, llama llamainterface.LLamaServer, systemprompt string, maxTokens int) (AdventureGame, error) {
	query := llamainterface.DefaultQueryCompletion()
	query.Prompt = fmt.Sprintf("<|start_header_id|>system<|end_header_id|>\n\n%s<|eot_id|>", systemprompt)

	tokens, errTokenize := llama.PostTokenize(query.Prompt, time.Minute*5)
	if errTokenize != nil {
		return AdventureGame{}, fmt.Errorf("system prompt tokenization error %s", errTokenize)
	}

	//gameNameAndTime := fmt.Sprintf("%s%d", gameName, time.Now().Unix())
	//os.MkdirAll(path.Join(SAVEGAMEDIR, gameNameAndTime), 0777)

	return AdventureGame{
		GameName:      gameName,
		StartTime:     time.Now(),
		llama:         llama,
		PromptEntries: []PromptEntry{PromptEntry{Text: query.Prompt, Tokens: tokens}},
		Introprompt:   query.Prompt,
		MaxTokens:     maxTokens}, nil
}

func (p *AdventureGame) GetSaveDir() string {
	return path.Join(SAVEGAMEDIR, fmt.Sprintf("%s%d", p.GameName, p.StartTime.Unix()))
}

func (p *AdventureGame) RunQuery(query llamainterface.QueryCompletion) (string, error) {
	timeoutLlama := time.Minute * 30

	fmt.Printf("---------- START QUERY---------------\n")
	result, errPost := p.llama.PostCompletion(query, nil, timeoutLlama)
	if errPost != nil {
		return "", errPost
	}
	fmt.Printf("-----------END QUERY ------------\n")
	fmt.Printf("\n\nGot Result object %#v\n", result)

	return result.Content, nil //Assuming that prompt ends with <|start_header_id|>dungeonmaster<|end_header_id|>  so this have to be just removed
}

func (p *AdventureGame) tokenizePrompt(prompt string) (PromptEntry, error) {
	tokens, errTokenize := p.llama.PostTokenize(strings.ReplaceAll(prompt, "\n", "\\n"), time.Minute*5)
	if errTokenize != nil {
		return PromptEntry{}, fmt.Errorf("prompt tokenization error %s", errTokenize)
	}
	return PromptEntry{Text: prompt, Tokens: tokens}, nil
}

// Write to disk? return filename?
func (p *AdventureGame) GeneratePicture(longText string) (string, string, error) {
	os.Mkdir(p.GetSaveDir(), 0777)
	outputPngName := path.Join(p.GetSaveDir(), fmt.Sprintf("out%d.png", time.Now().Unix()))

	query := gameBasicQuery()
	query.Prompt = artist.ArtPromptText(longText)

	fmt.Printf("\n----ART PROMPT----\n%s\n----------\n", query.Prompt)

	respText, errQuery := p.RunQuery(query) //----------RUN HERE!!!!!!!!!!!
	if errQuery != nil {
		return respText, "", errQuery
	}

	respText = strings.ReplaceAll(respText, "-", "")
	respText = strings.ReplaceAll(respText, "*", "") //Bullet list?

	img, imgErr := artist.CreatePic(respText)
	if imgErr != nil {
		return respText, "", imgErr
	}
	errSave := SavePng(outputPngName, img)

	p.PromptEntries[len(p.PromptEntries)-1].PictureDescription = respText
	p.PromptEntries[len(p.PromptEntries)-1].PictureFileName = outputPngName

	return respText, outputPngName, errSave
}

func (p *AdventureGame) UserInteraction(input string) (string, error) {
	query := gameBasicQuery()

	newPromptText := fmt.Sprintf("\n<|start_header_id|>player<|end_header_id|>\n\n%s<|eot_id|><|start_header_id|>dungeonmaster<|end_header_id|>", input)

	/*query.Prompt = fmt.Sprintf("%s\n<|start_header_id|>player<|end_header_id|>%s<|eot_id|><|start_header_id|>dungeonmaster<|end_header_id|>",
	p.GetTotalPrompt(), input)*/

	newToken, errNewToken := p.tokenizePrompt(newPromptText)
	if errNewToken != nil {
		return "", fmt.Errorf("system prompt token err %s", errNewToken)
	}
	newToken.Type = PROMPTTYPE_PLAYER
	p.PromptEntries = append(p.PromptEntries, newToken)

	//color.Red(fmt.Sprintf("\n\nNYT HELVETTI PROMPT ENTRYT OVAT %#v\n\n", p.promptEntries))

	query.Prompt = p.GetTotalPrompt()
	color.Blue(fmt.Sprintf("\nUSE PROMPT: %s\n", query.Prompt))
	//query.PromptTokens = p.GetTotalTokens()

	respText, errQuery := p.RunQuery(query) //----------RUN HERE!!!!!!!!!!!
	if errQuery != nil {
		return respText, errQuery
	}

	newToken, errNewToken = p.tokenizePrompt(respText)
	if errNewToken != nil {
		return "", errNewToken
	}
	newToken.Type = PROMPTTYPE_DM
	p.PromptEntries = append(p.PromptEntries, newToken)

	return strings.Replace(respText, "<|eot_id|>", "", -1), nil
}

func gameBasicQuery() llamainterface.QueryCompletion {
	query := llamainterface.DefaultQueryCompletion()

	query.Stream = false
	query.N_predict = 1024
	query.Temperature = 0.7
	query.Stop = []string{"<|eot_id|>"}
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

var artist Artist

func main() {
	pLoadFile := flag.String("load", "", "load existing game from json and continue")
	pBrancToNewGame := flag.Bool("b", false, "branch new game from loadfile")

	//TODO start from json  with -load at all times. These were just for bootstrapping
	//pGameFile := flag.String("f", "firetopadventure.txt", "uses this file as system prompt. This is what you want to play")
	//pArtistBaseTextFile := flag.String("ab", "promptArtist.txt", "prompt for generating image prompts from dungeon master text")
	//pMaxTokens := flag.Int("maxtokens", 1024, "number of max tokens for LLM model")

	pLlamafile := flag.String("l", LLAMAFILENAME, ".llamafile filename, starts that file as server")
	pLlamaport := flag.Int("lp", 0, "llamafile port, use 0 if let kernel decide port")
	pServerHost := flag.String("h", "127.0.0.1", "llama.cpp server host")
	pServerPort := flag.Int("p", 8080, "llama.cpp server port")

	pDiffusionModelFile := flag.String("dmf", "/home/henri/aimallit/stable-diffusionMuunnetut/sd-v1-4-ggml-model-f16.bin", "diffusion model file")

	pFluxHost := flag.String("fh", "127.0.0.1", "hostname of flux.1 server")
	pFluxPort := flag.Int("fp", 8800, "flux.1 server port")

	//For generating new adventures
	pCreateBlank := flag.String("blank", "", "Create blank game file as example from this for creating new adventure games")

	flag.Parse()

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

	ctxLllama, _ := context.WithCancel(context.Background())

	var llamacmdFlags llamainterface.ServerCommandLineFlags
	llamacmdFlags.ToDefaults()

	var llama llamainterface.LLamaServer
	if 0 < len(*pLlamafile) {
		var errLlamaInit error
		if *pLlamaport == 0 {
			errGetPort := llamacmdFlags.GetPort() //Pick random internal port
			if errGetPort != nil {
				fmt.Printf("Internal error for getting TCP port %s\n", errGetPort)
				return
			}
		} else {
			llamacmdFlags.ListenPort = *pLlamaport
		}
		//llamacmdFlags.NGPULayers = 9999
		llama, errLlamaInit = llamainterface.InitLlamafileServer(ctxLllama, *pLlamafile, llamacmdFlags)
		if errLlamaInit != nil {
			fmt.Printf("llama init fail %s\n", errLlamaInit)
			return
		}
	} else {
		var errSrv error
		llama, errSrv = llamainterface.InitLLamaServer(*pServerHost, *pServerPort)
		if errSrv != nil {
			fmt.Printf("err %s\n", errSrv)
			return
		}
	}

	healthStatus, _ := llama.GetHealth()
	for healthStatus != "ok" {
		var errhealth error
		healthStatus, errhealth = llama.GetHealth()
		if errhealth != nil {
			fmt.Printf("llama err...%s\n", errhealth.Error())
		} else {
			fmt.Printf("waiting llama %s\n", healthStatus)
		}
		time.Sleep(time.Second)
	}

	var game AdventureGame
	if len(*pLoadFile) == 0 { //Start new?
		fmt.Printf("Please load with -load blank or saved adventure")
		return
		/*
			fmt.Printf("\n--- Starting new game ---\n\n")
			//"You are dungeonmaster who describes scene where player character is at the moment. Player asks details about surroundings and dungeonmaster reports those. Also dungeon master tells what happens when player wants to do specific thing. Player approaches firetop mountain and sees entrance of maze at side of hill of mountain"
			gameFileText, errGameFile := os.ReadFile(*pGameFile)
			if errGameFile != nil {
				fmt.Printf("error reading game file %s err %s\n", *pGameFile, errGameFile)
				return
			}

			artistBaseText, errArtistBaseText := os.ReadFile(*pArtistBaseTextFile)
			if errArtistBaseText != nil {
				fmt.Printf("Error reading artist prompt file %s err:%s\n", *pArtistBaseTextFile, errArtistBaseText)
				return
			}

			var errGameInit error
			game, errGameInit = initAdventure(
				"firetopadventure",
				llama,
				string(gameFileText),
				*pMaxTokens)
			game.Artist.BaseText = string(artistBaseText)

			if errGameInit != nil {
				fmt.Printf("err game init %s\n", errGameInit)
				return
			}
		*/
	} else {
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
	}

	if 0 < len(*pCreateBlank) {
		errWriteBlank := game.WriteBlankFile(*pCreateBlank)
		if errWriteBlank != nil {
			fmt.Printf("error writing blank file %s  err:%s", *pCreateBlank, errWriteBlank)
			return
		}
		fmt.Printf("\n\n--Exported %s\n", *pCreateBlank)
		return
	}

	if len(game.PromptEntries) == 0 { //NEED TO INITIALIZE.  TODO PUT INTO LOAD METHOD!
		query := llamainterface.DefaultQueryCompletion()
		query.Prompt = fmt.Sprintf("<|start_header_id|>system<|end_header_id|>\n\n%s<|eot_id|>", game.Introprompt)

		tokens, errTokenize := llama.PostTokenize(query.Prompt, time.Minute*5)
		if errTokenize != nil {
			fmt.Errorf("system prompt tokenization error %s", errTokenize)
			return
		}
		game.PromptEntries = []PromptEntry{PromptEntry{Text: query.Prompt, Tokens: tokens}}
	}

	ui, uiErr := InitGraphicalUI(fmt.Sprintf("%s [%s]", game.GameName, game.StartTime))
	if uiErr != nil {
		fmt.Printf("UI init err %s\n", uiErr)
		return
	}
	ui.SetGenerating("Starting up, wait....")
	var artistInitErr error
	artist, artistInitErr = InitArtist(*pDiffusionModelFile, game.Artist.BaseText, "", "", imGen)
	if artistInitErr != nil {
		fmt.Printf("error initializing artist %s\n", artistInitErr)
		return
	}

	//Simple read input
	var errUser error
	userInput := "introduce game" //Starting query, not printing on ui, part of system prompt getting started
	if 0 < len(*pLoadFile) {
		last := game.PromptEntries[len(game.PromptEntries)-1]
		if last.Type == PROMPTTYPE_DM {
			userInput = "" //Nope... loaded game wait input
		}
		//get pic and latest text
		ui.SetDungeonMasterText(last.Text)
		ui.SetImagePromptText(last.PictureDescription)
		//latestPic, errLatestPic := LoadPng(last.PictureFileName)
		//if errLatestPic != nil {
		//	fmt.Printf("not possible to load pic %s  err:%s now just text\n", last.PictureFileName, errLatestPic)
		//} else {
		//	ui.SetPicture(lat
		//}
		ui.SetPicture(last.PictureFileName)
	}

	for {
		if 0 < len(userInput) {
			fmt.Printf("---process user input:%s ---\n", userInput)
			ui.SetGenerating("Running LLM....")
			ui.Render()

			dungeonMastersays, errRun := game.UserInteraction(userInput)
			if errRun != nil {
				fmt.Printf("ERROR: Failed running game %s\n", errRun)
				break
			} else {
				game.StoreLatestTextOutput()
				//ui.PrintDungeonMaster(dungeonMastersays)
				ui.SetDungeonMasterText(dungeonMastersays)
				renderErr := ui.Render()
				if renderErr != nil {
					fmt.Printf("error rendering UI %s\n", renderErr)
					return
				}
			}

			ui.SetGenerating("Generating picture...")
			ui.Render()
			picDescription, picFileName, errPic := game.GeneratePicture(dungeonMastersays)
			if errPic != nil {
				fmt.Printf("\n\nERROR generating picture:%s\n", errPic)
				return
			}
			fmt.Printf("picture filename:%s\n", picFileName)
			color.Yellow(picDescription)
			fmt.Printf("\n\n")
			ui.SetPicture(picFileName)

			ui.SetImagePromptText(picDescription)
			renderErr := ui.Render()
			if renderErr != nil {
				fmt.Printf("error rendering UI %s\n", renderErr)
				return
			}

			errSave := game.SaveGame()
			if errSave != nil {
				fmt.Printf("error saving game %s\n", errSave)
				return
			}
		}
		//userInput, errUser = ui.GetUserInput()
		fmt.Printf("--get user input--\n")
		userInput, errUser = ui.GetPrompt()
		if errUser != nil {
			fmt.Printf("\n\nExit by %s\n\n", errUser)
			break
		}
		userInput = strings.TrimSpace(userInput)
		//ui.PrintUser(userInput)

		updated, errUpdate := game.CheckOutputChanges()
		if errUpdate != nil {
			fmt.Printf("\n\nERR %s\n", errUpdate)
			return
		}
		if updated {
			fmt.Printf("\n\nUSER UPDATED PROMPT!!\n")
		}
	}

}
