package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
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

func localUIGetGame(ui *GraphicalUI, imGen ImageGenerator) (AdventureGame, bool, error) {
	var game AdventureGame
	loadingOldGame := false
	//Lets show main menu
	menuSel, errMenuSel := ui.RunMainMenu()
	if errMenuSel != nil {
		return AdventureGame{}, false, fmt.Errorf("menu selection fail %s", errMenuSel)
	}
	var cat GameCatalogue
	var catErr error
	switch menuSel {
	case MAINMENUSELECT_NEWGAME:
		cat, catErr = ListNewGames(GAMESDIR)
	case MAINMENUSELECT_CONTINUE, MAINMENUSELECT_BRANCHOLD:
		loadingOldGame = true
		cat, catErr = ListSavedGames(SAVEGAMEDIR)
	}
	if catErr != nil {
		return AdventureGame{}, false, fmt.Errorf("error catalogueing %s", catErr)
	}

	//TODO GENERATE MISSING GRAPHICS WHEN BUILDING CATALOGUE
	index := 0
	for gamename, catItem := range cat {
		index++
		ui.generatingText = fmt.Sprintf("Generating start image %v/%v : %s", index, len(cat), catItem.Game.GameName)
		ui.Render()
		errPrepare := catItem.PrepareImage(imGen)
		if errPrepare != nil {
			fmt.Printf("failed preparing %s\n", errPrepare)
		}
		cat[gamename] = catItem
	}

	pickResult, errPick := ui.PickFromCatalogue(cat)
	if errPick != nil {
		return AdventureGame{}, false, fmt.Errorf("localUIGetGame: error picking %s\n", errPick)
	}
	game = pickResult.Game

	if menuSel == MAINMENUSELECT_BRANCHOLD {
		game.StartTime = time.UnixMilli(0)
	}

	return game, loadingOldGame, nil
}

const INITIALPROMPT = "introduce game"

/*
Local interactive game flow  (vs stateless server)
*/
func localUIFlow(ui *GraphicalUI, loadingOldGame bool, game AdventureGame, temperature float64, imageTemperature float64, imGen ImageGenerator) error { //Run game on localhost screen with flow. Not web ui with stateless
	var errUser error
	userInput := INITIALPROMPT //Starting query, not printing on ui, part of system prompt getting started

	if loadingOldGame {
		last := game.PromptEntries[len(game.PromptEntries)-1]
		if last.Type == PROMPTTYPE_DM {
			userInput = "" //Nope... loaded game wait input
		}
		//get pic and latest text
		txt := game.promptFormatter.ExtractText(PLAYER, DUNGEONMASTER, last.Text)
		fmt.Printf("\n\nEXTRACTED TEXT IS\n%s\n\n", txt)

		ui.SetDungeonMasterText(txt)
		ui.SetImagePromptText(last.PictureDescription) //Just set for debug?
		ui.SetPicture(last.PictureFileName)
	}

	for {
		if 0 < len(userInput) {
			fmt.Printf("---process user input:%s ---\n", userInput)
			ui.SetGenerating("Running LLM....")
			ui.Render()

			dungeonMastersays, errRun := game.UserInteraction(temperature, userInput)
			if errRun != nil {
				return fmt.Errorf("ERROR: Failed running game %s\n", errRun)
			}
			game.StoreLatestTextOutput()
			//ui.PrintDungeonMaster(dungeonMastersays)
			ui.SetDungeonMasterText(dungeonMastersays)
			renderErr := ui.Render()
			if renderErr != nil {
				return fmt.Errorf("error rendering UI %s\n", renderErr)
			}

			ui.SetGenerating("Generating picture...")
			ui.Render()
			picDescription, picFileName, errPic := game.GeneratePicture(imageTemperature, dungeonMastersays, imGen)
			if errPic != nil {
				return fmt.Errorf("\n\nERROR generating picture:%s\n", errPic)
			}
			fmt.Printf("picture filename:%s\n", picFileName)
			color.Yellow(picDescription)
			fmt.Printf("\n\n")
			ui.SetPicture(picFileName)

			ui.SetImagePromptText(picDescription)
			renderErr = ui.Render()
			if renderErr != nil {
				return fmt.Errorf("error rendering UI %s", renderErr)
			}

			errSave := game.SaveGame()
			if errSave != nil {
				return fmt.Errorf("error saving game %s", errSave)
			}
		}
		fmt.Printf("--get user input--\n")
		userInput, errUser = ui.GetPrompt()
		if errUser != nil {
			fmt.Printf("\n\nExit by %s\n\n", errUser)
			break
		}
		userInput = strings.TrimSpace(userInput)

		updated, errUpdate := game.CheckOutputChanges()
		if errUpdate != nil {
			return fmt.Errorf("\n\nERR %s\n", errUpdate)
		}
		if updated {
			fmt.Printf("\n\nUSER UPDATED PROMPT!!\n")
		}
	}
	return nil
}
