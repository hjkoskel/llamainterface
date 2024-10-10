package main

import (
	"fmt"
	"llamainterface"
	"path"
	"time"
)

func localUIStart() (*GraphicalUI, error) {
	/**** Check is there title picture for menu *****/
	ui, uiErr := InitGraphicalUI("ADVENTURE")
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

func localUIGetGame(ui *GraphicalUI, imGen ImageGenerator, llama *llamainterface.LLamaServer) (AdventureGame, error) {
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

	cat, catErr := CreateToCatalogue(gameJsonList, llama)

	if catErr != nil {
		return AdventureGame{}, fmt.Errorf("error catalogueing %s", catErr)
	}

	//TODO GENERATE MISSING GRAPHICS WHEN BUILDING CATALOGUE
	/*
		index := 0
		for gamename, catItem := range cat {
			index++
			ui.generatingText = fmt.Sprintf("Generating start image %v/%v : %s", index, len(cat), catItem.Game.GameName)
			ui.Render()
			errPrepare := catItem.PrepareImage(imGen)
			if errPrepare != nil {
				fmt.Printf("failed preparing %s\n", errPrepare)
			}
			catItem.MenuImage = rl.LoadTexture(catItem.TitleImageFileName)
			if catItem.MenuImage.Width == 0 || catItem.MenuImage.Height == 0 {
				return AdventureGame{}, false, fmt.Errorf("error loading menu picture %s", catItem.MenuImage)
			}
			cat[gamename] = catItem
		}*/

	pickResult, errPick := ui.PickFromCatalogue(cat)
	if errPick != nil {
		return AdventureGame{}, fmt.Errorf("localUIGetGame: error picking %s\n", errPick)
	}
	game, errGameLoad := loadAdventure(pickResult.GameFileName, llama)
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

	for {
		fmt.Printf("--get user input--\n")
		userInput, errUser := ui.GetPrompt()
		if errUser != nil {
			fmt.Printf("\n\nExit by %s\n\n", errUser)
			break
		}
		fmt.Printf("---process user input:%s ---\n", userInput)
		ui.SetGenerating("Running LLM....")
		ui.Render()

		errRun := game.UserInteraction(temperature, imageTemperature, userInput)
		if errRun != nil {
			return fmt.Errorf("ERROR: Failed running game: %s\n", errRun)
		}
		last := game.Pages[len(game.Pages)-1]

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
