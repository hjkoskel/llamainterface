package main

import (
	"fmt"
	"llamainterface"
	"path"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func localUIStart(fontFilename string) (*GraphicalUI, error) {
	/**** Check is there title picture for menu *****/
	ui, uiErr := InitGraphicalUI("ADVENTURE", fontFilename)
	if uiErr != nil {
		return nil, fmt.Errorf("UI init err %s\n", uiErr)
	}

	errImg := ui.SplashScreen(TITLEPICTUREFILE)
	if errImg != nil {
		return nil, fmt.Errorf("Failed loading title...%s re-trying generate title\n\n", errImg)
	}
	ui.Render()
	ui.SetGenerating("Starting up LLM...")
	ui.Render()

	return ui, nil
}

func localUIGetGame(ui *GraphicalUI, imGen ImageGenerator, llama *llamainterface.LLamaServer, translator *Translator, ttsEngine *TTSInterf) (AdventureGame, error) {
	var game AdventureGame
	//Lets show main menu
	menuSel, errMenuSel := ui.RunMainMenu()
	if errMenuSel != nil {
		return AdventureGame{}, fmt.Errorf("menu selection fail %s", errMenuSel)
	}

	var gameJsonList []string
	var errGameList error

	switch menuSel {
	case MAINMENUSELECT_NEWGAME:
		gameJsonList, errGameList = ListNewGamesJson(GAMESDIR, llama)
	case MAINMENUSELECT_CONTINUE, MAINMENUSELECT_BRANCHOLD:
		gameJsonList, errGameList = ListSavedGamesJson(SAVEGAMEDIR, llama)

	}
	if errGameList != nil {
		return AdventureGame{}, fmt.Errorf("Error game list %s\n", errGameList)
	}

	fmt.Printf("--- List of game json files ---\n")
	for i, s := range gameJsonList {
		fmt.Printf(" %v: %s\n", i, s)
	}

	cat, catErr := CreateToCatalogue(gameJsonList, llama, translator, ttsEngine)

	if catErr != nil {
		return AdventureGame{}, fmt.Errorf("error catalogueing %s", catErr)
	}

	pickResult, errPick := ui.PickFromCatalogue(cat)
	if errPick != nil {
		return AdventureGame{}, fmt.Errorf("localUIGetGame: error picking %s\n", errPick)
	}
	game, errGameLoad := loadAdventure(pickResult.GameFileName, llama, &textPromptFormatter, translator, ttsEngine)
	if errGameLoad != nil {
		return AdventureGame{}, errGameLoad
	}

	if menuSel == MAINMENUSELECT_BRANCHOLD {
		game.StartTime = time.UnixMilli(0)
	}

	return game, nil
}

const INITIALPROMPT = "introduce game"

/*
Local interactive game flow  (vs stateless server)
*/
func localUIFlow(ui *GraphicalUI, game AdventureGame, temperature float64, imageTemperature float64, imGen *ImageGenerator) error { //Run game on localhost screen with flow. Not web ui with stateless
	if len(game.Pages) == 0 { //New game, create first page
		ui.SetGenerating("Running LLM first time....")
		ui.Render()
		errInitial := game.UserInteraction(temperature, imageTemperature, "")
		if errInitial != nil {
			return errInitial
		}
	}
	last := game.Pages[len(game.Pages)-1]

	ui.SetGenerating("Generating audio...")
	ui.Render()
	errGenTTS := game.GenerateLatestTTS(game.ttsengine)
	if errGenTTS != nil {
		return errGenTTS
	}

	if !game.Textmode {
		ui.SetPage(path.Join(game.GetSaveDir(), last.PictureFileName()), last) //menucpicture is the last or menupicture if no latest pic
		ui.SetGenerating("Generating picture...")
		ui.Render()
		imgGenErr := game.GenerateLatestPicture(imageTemperature, imGen)
		if imgGenErr != nil {
			return imgGenErr
		}
	}
	ui.SetPage(path.Join(game.GetSaveDir(), last.PictureFileName()), last) //menucpicture is the last or menupicture if no latest pic
	ui.Render()

	var snd rl.Sound
	if game.ttsengine != nil {
		snd, _ = ui.PlayAudioFile(path.Join(game.GetSaveDir(), last.TTSTextFileName(game.translator.Lang)))
	}
	for {
		fmt.Printf("--get user input--\n")
		userInput, errUser := ui.GetPrompt()
		if errUser != nil {
			fmt.Printf("\n\nExit by %s\n\n", errUser)
			break
		}
		rl.UnloadSound(snd)
		fmt.Printf("---process user input:%s ---\n", userInput)
		ui.SetGenerating("Running LLM....")
		ui.Render()

		errRun := game.UserInteraction(temperature, imageTemperature, userInput)
		if errRun != nil {
			return fmt.Errorf("ERROR: Failed running game: %s\n", errRun)
		}
		last := game.Pages[len(game.Pages)-1]

		ui.SetGenerating("Generating audio...")
		ui.Render()
		errGenTTS := game.GenerateLatestTTS(game.ttsengine)
		if errGenTTS != nil {
			return errGenTTS
		}
		snd, _ = ui.PlayAudioFile(path.Join(game.GetSaveDir(), last.TTSTextFileName(game.translator.Lang)))

		if !game.Textmode {
			ui.SetPage(path.Join(game.GetSaveDir(), last.PictureFileName()), last)
			ui.SetGenerating("Generating picture...")
			ui.Render()
			imgGenErr := game.GenerateLatestPicture(imageTemperature, imGen)
			if imgGenErr != nil {
				return imgGenErr
			}
		}
		last = game.Pages[len(game.Pages)-1]
		ui.SetPage(path.Join(game.GetSaveDir(), last.PictureFileName()), last)
		ui.SetGenerating("Generating picture...")
		ui.Render()

	}
	return nil
}
