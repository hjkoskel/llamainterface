/*
Web interface

Allows browsing games and adding entrys


*/

package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/png"
	"io"
	"llamainterface"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	b64 "encoding/base64"

	"github.com/gin-gonic/gin"
	"golang.org/x/image/draw"
)

const WANTEDTHUMBNAILWIDTH = 128

type WebInterface struct {
	router *gin.Engine
	//mainPageCode    string
	//newGamePageCode string //Pre-generated on init
	imGen            *ImageGenerator
	llama            *llamainterface.LLamaServer
	textTemperature  float64
	imageTemperature float64
}

func RunAsWebServer(portNumber int, imGen ImageGenerator, llama llamainterface.LLamaServer, textTemperature float64, imageTemperature float64) error {
	webUi, errWebUi := InitWebInterface(imGen, llama)
	webUi.textTemperature = textTemperature
	webUi.imageTemperature = imageTemperature
	if errWebUi != nil {
		fmt.Errorf("web ui init fail %s\n", errWebUi)
		os.Exit(-1)
	}
	webUiRuntimeErr := webUi.Listen(portNumber)
	fmt.Printf("web router fail %s\n", webUiRuntimeErr)
	return webUiRuntimeErr
}

func (p *WebInterface) Listen(portNumber int) error {
	connString := fmt.Sprintf(":%v", portNumber)
	fmt.Printf("\nLISTENING UI ON %s\n", connString)
	return p.router.Run(connString)
}

//go:embed webui/gallery.css
var emb_galleryCss []byte

func loadGameById(gameid string) (AdventureGame, error) {
	dirname := path.Join(SAVEGAMEDIR, gameid)
	jsonList, errJsonList := getFirstJsonFilesFromDir(dirname)
	if errJsonList != nil {
		return AdventureGame{}, errJsonList
	}
	if len(jsonList) != 1 {
		return AdventureGame{}, fmt.Errorf("invalid number of json files on game dir %s on %s", dirname, len(jsonList))
	}
	return loadAdventure(jsonList[0])
}

//go:embed webui/gamepageStart.html
var emb_gamepageStart []byte

func PromptEntryToHtml(gameTitle string, entry PromptEntry, userPrompt string, thisPage int, pages int) string {
	var sb strings.Builder

	sb.Write(emb_gamepageStart)

	sb.WriteString("<div id=\"game-container\">\n")

	sb.WriteString("<div class=\"nav-buttons\">")
	if 0 < thisPage {
		//sb.WriteString("<button class=\"nav-button\" id=\"back-button\">Back</button>\n")
		sb.WriteString(fmt.Sprintf("<a href=\"%v\" class=\"nav-button\" id=\"back-button\">Back</a>\n", thisPage-1))
	} else {
		sb.WriteString("<button disabled class=\"nav-button\" id=\"back-button\">Back</button>\n")

	}
	if thisPage < pages {
		//sb.WriteString("<button class=\"nav-button\" id=\"forward-button\">Forward</button>\n")
		sb.WriteString(fmt.Sprintf("<a href=\"%v\" class=\"nav-button\" id=\"forward-button\">Forward</a>\n", thisPage+1))
	}
	sb.WriteString("</div>\n\n")

	if len(entry.PictureFileName) != 0 {
		sb.WriteString("<img src=\"data:image/png;charset=utf-8;base64,")
		byt, errReadPic := os.ReadFile(entry.PictureFileName)
		if errReadPic != nil {
			fmt.Printf("ERROOR READING PIC |%s|  err:%s\n", entry.PictureFileName, errReadPic)
			return "ERROR!"
		}
		codedString := b64.StdEncoding.EncodeToString(byt)
		sb.WriteString(codedString)
		sb.WriteString("\"></img>\n")
	}

	sb.WriteString("<div id=\"game-text\">\n")
	sb.WriteString(entry.Text)
	sb.WriteString("</div>\n\n")

	if len(userPrompt) == 0 {
		sb.WriteString("<form method=\"post\">\n")
		sb.WriteString("<input type=\"text\" id=\"command-input\" name=\"userPrompt\" placeholder=\"Enter your command here\" />\n")
		sb.WriteString("<input id=\"submit-button\" type=\"submit\" value=\"Submit\"/>\n")
		sb.WriteString("</form>\n")
		//sb.WriteString(<button id=\"submit-button\">Submit</button></div>")
	} else {
		sb.WriteString("<div id=\"game-text\">\n")
		sb.WriteString(userPrompt)
		sb.WriteString("</div>\n\n")
	}

	sb.WriteString("</body>\n</html>\n")
	return sb.String()
}

//go:embed webui/main.html
var emb_mainHtml []byte

//go:embed title.png
var emb_titlePng []byte

func loadPngToSmallPng(fname string, wantedWidth int) ([]byte, error) {
	f, errOpen := os.Open(fname)
	if errOpen != nil {
		return nil, fmt.Errorf("error opening png file %s err:%s", fname, errOpen)
	}
	img, errDecode := png.Decode(f)
	if errDecode != nil {
		return nil, fmt.Errorf("error decoding png data on %s err:%s", fname, errDecode)
	}
	ratio := float64(wantedWidth) / float64(img.Bounds().Dx())
	result := image.NewRGBA(image.Rect(0, 0, wantedWidth, int(float64(img.Bounds().Dx())*ratio)))

	kernel := draw.BiLinear
	kernel.Scale(result, result.Rect, img, img.Bounds(), draw.Over, nil)

	bw := bytes.NewBuffer(nil)
	errEncode := png.Encode(bw, result)
	return bw.Bytes(), errEncode

}

func InitWebInterface(imGen ImageGenerator, llama llamainterface.LLamaServer) (WebInterface, error) {

	result := WebInterface{
		router: gin.Default(),
		imGen:  &imGen,
		llama:  &llama,
	}
	result.router.GET("/gallery.css", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/css; charset=utf-8", emb_galleryCss)
	})

	result.router.GET("/", func(c *gin.Context) {
		//fmt.Printf("MAIN PAGE")
		c.Data(http.StatusOK, "text/html; charset=utf-8", emb_mainHtml)
	})

	result.router.GET("/title.png", func(c *gin.Context) {
		c.Data(http.StatusOK, "image/png", emb_titlePng)
	})

	result.router.StaticFS("/games", http.Dir("games"))

	result.router.GET("/savegames/:gameid", func(c *gin.Context) {
		fmt.Printf("SAVEGAME HAE VIIMEISIN %s\n", c.Param("gameid"))

		game, errLoad := loadGameById(c.Param("gameid"))
		if errLoad != nil {
			c.Error(errLoad)
			return
		}
		redirectURL := fmt.Sprintf("/savegames/%s/%v", c.Param("gameid"), len(game.PromptEntries)/2)
		fmt.Printf("Asking latest, redirecting to %s\n", redirectURL)
		c.Redirect(http.StatusFound, redirectURL)
	})

	result.router.GET("/savegames/:gameid/:pageNumber", func(c *gin.Context) {
		fmt.Printf("SAVEGAME HAE %s entry %s\n", c.Param("gameid"), c.Param("index"))

		game, errLoad := loadGameById(c.Param("gameid"))
		if errLoad != nil {
			c.Error(errLoad)
			return
		}
		pageNumber, errPageNumber := strconv.ParseInt(c.Param("pageNumber"), 10, 64)
		if errPageNumber != nil {
			c.Error(fmt.Errorf("invalid pageNumber %s err:%s", c.Param("pageNumber"), errPageNumber))
			return
		}

		i := int(pageNumber) * 2
		pages := len(game.PromptEntries) / 2

		last := game.PromptEntries[len(game.PromptEntries)-1]
		if i+1 < len(game.PromptEntries) {
			a := game.PromptEntries[i]
			b := game.PromptEntries[i+1]
			last = a
			prompted := b.Text
			if a.Type == PROMPTTYPE_PLAYER {
				prompted = a.Text
				last = b
			}
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(PromptEntryToHtml(game.GameName, last, prompted, int(pageNumber), pages)))
		} else {
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(PromptEntryToHtml(game.GameName, last, "", pages, pages)))
		}
	})

	//TODO TEE PIKKUKUVAT, clientti muuten lataa liikaa
	/*	result.router.GET("/latestpic/:gameid", func(c *gin.Context) {
		game, errLoad := loadGameById(c.Param("gameid"))
		if errLoad != nil {
			c.Error(errLoad)
			return
		}
		i := game.latestEntryIndexForDM()
		byt, errReadFile := loadPngToSmallPng(game.PromptEntries[i].PictureFileName, WANTEDTHUMBNAILWIDTH)

		//byt, errReadFile := os.ReadFile(game.PromptEntries[i].PictureFileName)
		if errReadFile != nil {
			c.Error(errReadFile)
			return
		}
		c.Data(http.StatusOK, "image/png", byt)
	})*/

	result.router.GET("/newgame/", func(c *gin.Context) {
		newcat, newcatErr := ListNewGames(GAMESDIR)
		if newcatErr != nil {
			c.Error(fmt.Errorf("error listing new games %s", newcatErr))
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n\t<meta charset=\"UTF-8\">\n\t<meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n\t<meta http-equiv=\"X-UA-Compatible\" content=\"IE=edge\">\n\t<title>Start new game</title>\n\t<link rel=\"stylesheet\" href=\"/gallery.css\">\n</head>\n<body>\n"+catalogNewToHtml(newcat)+"</body>\n</html>\n"))
	})

	result.router.GET("/newgame/:gamename", func(c *gin.Context) { //Latest entry on page
		gamename := c.Param("gamename")
		newgamefilename := path.Join(GAMESDIR, gamename+".json")
		fmt.Printf("Going to start %s\n", newgamefilename)
		game, errLoad := loadAdventure(newgamefilename)
		if errLoad != nil {
			c.Error(fmt.Errorf("internal error:%s", errLoad))
			return
		}
		game.llama = *result.llama

		if len(game.PromptEntries) == 0 { //NEED TO INITIALIZE.  TODO PUT INTO LOAD METHOD!
			query := llamainterface.DefaultQueryCompletion()
			query.Prompt = game.promptFormatter.Format(game.Introprompt, nil, PLAYER, DUNGEONMASTER, "")
			tokens, errTokenize := llama.PostTokenize(query.Prompt, time.Minute*5)
			if errTokenize != nil {
				c.Error(fmt.Errorf("system prompt tokenization error %s", errTokenize))
				return
			}
			game.PromptEntries = []PromptEntry{PromptEntry{Text: query.Prompt, Tokens: tokens}}
		}

		dungeonMastersays, errRun := game.UserInteraction(result.textTemperature, INITIALPROMPT)
		if errRun != nil {
			c.Error(fmt.Errorf("error running:%s", errRun))
			return
		}
		storeErr := game.StoreLatestTextOutput()
		if storeErr != nil {
			c.Error(storeErr)
			return
		}

		errSave := game.SaveGame()
		if errSave != nil {
			c.Error(fmt.Errorf("error saving game %s", errSave))
			return
		}

		fmt.Printf("\nDUNGEONMASTER SAID %s\n", dungeonMastersays)

		_, _, errPic := game.GeneratePicture(result.imageTemperature, dungeonMastersays, imGen)
		if errPic != nil {
			c.Error(fmt.Errorf("\n\nERROR generating picture:%s\n", errPic))
			return
		}

		errSave = game.SaveGame()
		if errSave != nil {
			c.Error(fmt.Errorf("error saving game %s", errSave))
			return
		}

		redirectURL := fmt.Sprintf("/savegames/%s", game.GameId())
		fmt.Printf("Asking latest, redirecting to %s\n", redirectURL)
		c.Redirect(http.StatusFound, redirectURL)

	})

	result.router.GET("/continuegame", func(c *gin.Context) { //Get catalog of all savegames
		cat, catErr := ListSavedGames(SAVEGAMEDIR)
		if catErr != nil {
			fmt.Printf("TODO ERR %s\n", catErr)
		}

		fmt.Printf("\n\nCONTINUE GAME:%#v\n", cat)

		s := "<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n\t<meta charset=\"UTF-8\">\n\t<meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n\t<meta http-equiv=\"X-UA-Compatible\" content=\"IE=edge\">\n\t<title>Load existing game</title>\n\t<link rel=\"stylesheet\" href=\"gallery.css\">\n</head>\n<body>\n" + catalogLatestToHtml(cat) + "</body>\n</html>\n"
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(s))

	})

	/*result.router.POST("/savegames/:gameid", func(c *gin.Context) { //Post prompt
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("JAAAHA"))
	})*/

	result.router.POST("/savegames/:gameid/:pageNumber", func(c *gin.Context) { //Post prompt
		fmt.Printf("running post gameid=%s\n", c.Param("gameid"))
		requestBody, errRequestBody := io.ReadAll(c.Request.Body)
		if errRequestBody != nil {
			c.Error(errRequestBody)
			return
		}

		s, errS := url.QueryUnescape(string(requestBody))
		if errS != nil {
			c.Error(errS)
			return
		}

		s = strings.Replace(s, "userPrompt=", "", 1)
		fmt.Printf("RUUMIS=%s\n", s)

		game, errLoad := loadGameById(c.Param("gameid"))
		if errLoad != nil {
			c.Error(fmt.Errorf("game load err:%s", errLoad))
			return
		}
		game.llama = *result.llama
		dungeonMastersays, errRun := game.UserInteraction(result.textTemperature, s)
		if errRun != nil {
			c.Error(fmt.Errorf("error running:%s", errRun))
			return
		}
		storeErr := game.StoreLatestTextOutput()
		if storeErr != nil {
			c.Error(storeErr)
			return
		}

		fmt.Printf("\nDUNGEONMASTER SAID %s\n", dungeonMastersays)

		_, _, errPic := game.GeneratePicture(result.imageTemperature, dungeonMastersays, imGen)
		if errPic != nil {
			c.Error(fmt.Errorf("\n\nERROR generating picture:%s\n", errPic))
			return
		}

		errSave := game.SaveGame()
		if errSave != nil {
			c.Error(fmt.Errorf("error saving game %s", errSave))
			return
		}

		i := len(game.PromptEntries) - 1
		pages := len(game.PromptEntries) / 2
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(PromptEntryToHtml(game.GameName, game.PromptEntries[i], "", pages, pages)))
		redirectURL := fmt.Sprintf("/savegames/%s/%v", c.Param("gameid"), len(game.PromptEntries)/2)
		fmt.Printf("Asking latest, redirecting to %s\n", redirectURL)
		c.Redirect(http.StatusFound, redirectURL)

	})

	/*
		result.router.GET("/data/:savegameid", func(c *gin.Context) {
			savegameid := c.Param("savegameid")
			fmt.Printf("Get latest :%s\n", savegameid)
			c.String(http.StatusOK, "TODO")

		})

		result.router.GET("/data/:savegameid/:entryid", func(c *gin.Context) {
			savegameid := c.Param("savegameid")
			entryid := c.Param("entryid")
			fmt.Printf("TODO Report from disk savegame:%s  entry:%s!\n", savegameid, entryid)
			c.String(http.StatusOK, "TODO")
		})
	*/

	return result, nil

}

func catalogLatestToHtml(cat GameCatalogue) string {
	var sb strings.Builder

	sb.WriteString("\t<div class=\"gallery\">\n")
	for _, item := range cat {
		sb.WriteString(fmt.Sprintf("<a href=\"/savegames/%s\">", item.Game.GameId()))
		sb.WriteString("\t\t<div class=\"gallery-item\">\n")
		//lastEntry := item.Game.PromptEntries[len(item.Game.PromptEntries)-1]

		//TODO META MUODOSSA KUVA!
		//sb.WriteString(fmt.Sprintf("\t\t\t<img src=\"/latestpic/%s\" alt=\"%s\">\n", item.Game.GameId(), item.Game.GameName))

		i := item.Game.latestEntryIndexForDM()
		byt, errReadFile := loadPngToSmallPng(item.Game.PromptEntries[i].PictureFileName, WANTEDTHUMBNAILWIDTH)
		if errReadFile != nil {
			continue //BAD HACK
		}

		sb.WriteString("<img src=\"data:image/png;charset=utf-8;base64,")
		codedString := b64.StdEncoding.EncodeToString(byt)
		sb.WriteString(codedString)
		sb.WriteString("\"></img>\n")

		sb.WriteString(fmt.Sprintf("\t\t\t<div class=\"title\">%s</div>\n", item.Game.GameName))
		sb.WriteString("\t\t</div>\n")
		sb.WriteString("</a>\n")
	}
	sb.WriteString("\t\t</div>\n")
	return sb.String()
}

func catalogNewToHtml(cat GameCatalogue) string {
	var sb strings.Builder

	sb.WriteString("\t<div class=\"gallery\">\n")
	for _, item := range cat {
		sb.WriteString(fmt.Sprintf("<a href=\"%s\">", item.Game.GameName))
		sb.WriteString("\t\t<div class=\"gallery-item\">\n")
		//sb.WriteString(fmt.Sprintf("\t\t\t<img src=\"/titlepic/%s\" alt=\"%s\">\n", item.TitleImageFileName, item.Game.GameName))
		byt, errReadFile := loadPngToSmallPng("./"+item.TitleImageFileName, WANTEDTHUMBNAILWIDTH)
		if errReadFile != nil {
			fmt.Printf("ERROR LOADING %s  err:%s\n", item.TitleImageFileName, errReadFile)
			continue //BAD HACK
		}

		sb.WriteString("<img src=\"data:image/png;charset=utf-8;base64,")
		codedString := b64.StdEncoding.EncodeToString(byt)
		sb.WriteString(codedString)
		sb.WriteString("\"></img>\n")

		sb.WriteString(fmt.Sprintf("\t\t\t<div class=\"title\">%s</div>\n", item.Game.GameName))
		sb.WriteString("\t\t</div>\n")
		sb.WriteString("</a>\n")
	}
	sb.WriteString("\t\t</div>\n")
	return sb.String()
}
