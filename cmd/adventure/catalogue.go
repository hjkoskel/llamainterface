/*
Game catalogue

List new games and existing statuses
Used for rendering menu
*/
package main

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type GameCatalogueEntry struct {
	Game               AdventureGame
	TitleImageFileName string //Whole path
	MenuImage          rl.Texture2D
	Description        string
}

// Load existing or generate
func (p *GameCatalogueEntry) PrepareImage(gen ImageGenerator) error {
	//Get latest available pic
	p.Description = "New game:" + p.Game.GameName
	for i := len(p.Game.PromptEntries) - 1; 0 <= i; i++ {
		texture := rl.LoadTexture(p.Game.PromptEntries[i].PictureFileName)
		if texture.Width != 0 && texture.Height != 0 {
			p.MenuImage = texture //Shrink?
			when := p.Game.LastTime()
			p.Description = fmt.Sprintf("%s entry%d [%s]", p.Game.GameName, i, when.Local().Format("2006-01-02 15:04:05"))
			return nil
		}
	}
	p.MenuImage = rl.LoadTexture(p.TitleImageFileName)
	if p.MenuImage.Width != 0 && p.MenuImage.Height != 0 {
		return nil
	}
	if p.TitleImageFileName == "" {
		return fmt.Errorf("internal error:image file name not defined")
	}

	if p.Game.TitleGraphicPrompt == "" {
		return fmt.Errorf("internal error:No game TitleGraphicPrompt")
	}

	img, imgErr := gen.CreatePic(p.Game.TitleGraphicPrompt)
	if imgErr != nil {
		return fmt.Errorf("error creating pic %s", imgErr)
	}
	errSave := SavePng(p.TitleImageFileName, img)
	if errSave != nil {
		return fmt.Errorf("error saving title picture to %s  err:%s", p.TitleImageFileName, errSave)
	}
	p.MenuImage = rl.LoadTexture(p.TitleImageFileName)
	if p.MenuImage.Width != 0 && p.MenuImage.Height != 0 {
		return nil
	}
	return fmt.Errorf("internal error, still no title picture %s", p.TitleImageFileName)
}

type GameCatalogue []GameCatalogueEntry

// Sort last played
func (m GameCatalogue) Len() int { return len(m) }
func (m GameCatalogue) Less(i, j int) bool {
	a := []GameCatalogueEntry(m)[i]
	b := []GameCatalogueEntry(m)[j]

	return a.Game.LastTime().After(b.Game.LastTime())
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

func ListSavedGames(savegamedir string) (GameCatalogue, error) {
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
		jsonFileList, errListJson := getFirstJsonFilesFromDir(path.Join(savegamedir, dirEntry.Name()))
		if errListJson != nil || len(jsonFileList) == 0 {
			continue
		}
		if 0 < len(jsonFileList) {
			fmt.Printf("WARNING MULTIPLE JSON FILES %#v\n", jsonFileList) //TODO HOW TO HANDLE?
		}
		jsonfilename := jsonFileList[0]

		g, errG := loadAdventure(jsonfilename)
		if errG != nil {
			fmt.Printf("was not able to open game %s error:%s", jsonfilename, errG)
			continue
		}

		//Is there title picture. If there is, load
		result = append(result, GameCatalogueEntry{
			Game:               g,
			TitleImageFileName: strings.Replace(jsonfilename, ".json", ".png", 1),
			//TitleImage     rl.Texture2D
		})
	}
	sort.Sort(GameCatalogue(result)) //If there is timestamp?
	return result, nil

}

// ListNewGames.. if there are no pictures of games then those are rendered after...
func ListNewGames(newgameDirName string) (GameCatalogue, error) {
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
		g, errG := loadAdventure(totalFilename)
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
