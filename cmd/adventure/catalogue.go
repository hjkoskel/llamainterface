/*
Game catalogue

List new games and existing statuses
Used for rendering menu
*/
package main

import (
	"fmt"
	"llamainterface"
	"math"
	"os"
	"path"
	"strings"
	"time"
)

type GameCatalogueEntry struct {
	//Game               AdventureGame  REALLY DO JUST CATALOGUEING
	Id                 string
	Name               string
	TitleDescription   string
	TitleImageFileName string //Whole path to title image or to the latest picture in game
	MenuPicture        string

	Label        string
	GameFileName string
	LastPlayed   time.Time
}

type GameCatalogue []GameCatalogueEntry

// Sort last played
func (m GameCatalogue) Len() int { return len(m) }
func (m GameCatalogue) Less(i, j int) bool {
	a := []GameCatalogueEntry(m)[i].LastPlayed
	b := []GameCatalogueEntry(m)[j].LastPlayed
	return a.After(b)
}

func (m GameCatalogue) Swap(i, j int) { m[i], m[j] = m[j], m[i] }

func getFirstJsonFilesFromDir(dirname string) ([]string, error) {
	content, errcontent := os.ReadDir(dirname)
	if errcontent != nil {
		return nil, errcontent
	}

	names := []string{}
	for _, entry := range content {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".json") {
			names = append(names, path.Join(dirname, entry.Name()))
		}
	}
	return names, nil
}

func ListSavedGamesJson(savegamedir string, llama *llamainterface.LLamaServer) ([]string, error) {
	result := []string{}
	saveDirContent, errDirContent := os.ReadDir(savegamedir)
	if errDirContent != nil {
		return nil, errDirContent
	}
	for _, dirEntry := range saveDirContent {
		if !dirEntry.IsDir() {
			continue
		}
		jsonFiles, errJsonFiles := getFirstJsonFilesFromDir(path.Join(savegamedir, dirEntry.Name()))
		if errJsonFiles != nil {
			fmt.Printf("error listing savegame %s\n", dirEntry.Name())
			continue
		}
		if len(jsonFiles) != 1 {
			fmt.Printf("invalid number of json files %d\n", len(jsonFiles))
			continue
		}
		//Get json from dir
		result = append(result, jsonFiles[0])
	}
	return result, nil
}

func ListNewGamesJson(newgameDirName string, llama *llamainterface.LLamaServer) ([]string, error) {
	result := []string{}

	dirContent, errDirContent := os.ReadDir(newgameDirName)
	if errDirContent != nil {
		return nil, errDirContent
	}
	for _, fileEntry := range dirContent {
		if fileEntry.IsDir() || !strings.HasSuffix(fileEntry.Name(), ".json") {
			continue
		}
		result = append(result, path.Join(newgameDirName, fileEntry.Name()))
	}
	return result, nil
}

func GameJsonToCatalogueEntry(fname string, llama *llamainterface.LLamaServer) (GameCatalogueEntry, error) {
	g, errG := loadAdventure(fname, llama)
	if errG != nil {
		return GameCatalogueEntry{}, fmt.Errorf("was not able to open game %s error:%s", fname, errG)
	}
	gametime := g.StartTime
	labeltext := "New:" + g.GameName
	if 0 < len(g.Pages) {
		gametime = g.Pages[len(g.Pages)-1].Timestamp
		sinceLastPlay := time.Duration(math.Floor(time.Since(gametime).Seconds())) * time.Second
		labeltext = fmt.Sprintf("%s n:%v %s", g.GameName, len(g.Pages), sinceLastPlay)
	}

	return GameCatalogueEntry{
		Id:                 g.GameId(),
		Name:               g.GameName,
		TitleDescription:   g.TitleGraphicPrompt,
		TitleImageFileName: g.gameTitlepictureFilename,
		MenuPicture:        g.GetMenuPictureFile(),
		Label:              labeltext,
		GameFileName:       fname,
		LastPlayed:         gametime,
	}, nil
}

func CreateToCatalogue(filelist []string, llama *llamainterface.LLamaServer) (GameCatalogue, error) {
	result := make([]GameCatalogueEntry, len(filelist))
	var errLoad error
	for i, fname := range filelist {
		result[i], errLoad = GameJsonToCatalogueEntry(fname, llama)
		if errLoad != nil {
			return GameCatalogue{}, errLoad
		}
	}
	return result, nil
}

/*
func ListSavedGames(savegamedir string, llama *llamainterface.LLamaServer) (GameCatalogue, error) {
	result := []GameCatalogueEntry{}

	saveDirContent, errDirContent := os.ReadDir(savegamedir)
	if errDirContent != nil {
		return GameCatalogue{}, errDirContent
	}
	for _, dirEntry := range saveDirContent {
		if !dirEntry.IsDir() {
			continue
		}
		//Get json from dir
		jsonFilename := path.Join(savegamedir, dirEntry.Name())
		jsonFileList, errListJson := getFirstJsonFilesFromDir(jsonFilename)
		if errListJson != nil || len(jsonFileList) == 0 {
			continue
		}
		if 1 < len(jsonFileList) {
			fmt.Printf("WARNING MULTIPLE JSON FILES %#v\n", jsonFileList) //TODO HOW TO HANDLE?
		}
		jsonfilename := jsonFileList[0]

		g, errG := loadAdventure(jsonfilename, llama)
		if errG != nil {
			fmt.Printf("was not able to open game %s error:%s", jsonfilename, errG)
			continue
		}

		gametime := g.StartTime
		if 0 < len(g.Pages) {
			gametime = g.Pages[len(g.Pages)-1].Timestamp
		}

		//Is there title picture. If there is, load
		result = append(result, GameCatalogueEntry{
			TitleImageFileName: g.gameTitlepictureFilename,
			Label:              "New:" + g.GameName,
			GameFileName:       jsonFilename,
			LastPlayed:         gametime,
		})
	}
	sort.Sort(GameCatalogue(result)) //If there is timestamp?
	return result, nil

}

// ListNewGames.. if there are no pictures of games then those are rendered after...
func ListNewGames(newgameDirName string, llama *llamainterface.LLamaServer) (GameCatalogue, error) {
	result := []GameCatalogueEntry{}

	dirContent, errDirContent := os.ReadDir(newgameDirName)
	if errDirContent != nil {
		return GameCatalogue{}, errDirContent
	}
	for _, fileEntry := range dirContent {
		if fileEntry.IsDir() || !strings.HasSuffix(fileEntry.Name(), ".json") {
			continue
		}
		totalFilename := path.Join(newgameDirName, fileEntry.Name())
		g, errG := loadAdventure(totalFilename, llama)
		if errG != nil {
			fmt.Printf("was not able to open game %s error:%s", totalFilename, errG)
			continue
		}

		//Is there title picture. If there is, load
		result = append(result, GameCatalogueEntry{
			Game:               g,
			TitleImageFileName: strings.Replace(totalFilename, ".json", ".png", 1),
			//TitleImage     rl.Texture2D
		})
	}
	sort.Sort(GameCatalogue(result)) //If there is timestamp?
	return result, nil
}
*/
