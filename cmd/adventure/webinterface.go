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

func RunAsWebServer(portNumber int, imGen *ImageGenerator, llama *llamainterface.LLamaServer, textTemperature float64, imageTemperature float64) error {
	webUi, errWebUi := InitWebInterface(imGen, llama)
	webUi.textTemperature = textTemperature
	webUi.imageTemperature = imageTemperature
	if errWebUi != nil {
		return fmt.Errorf("web ui init fail %s", errWebUi)
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

func loadGameById(gameid string, llama *llamainterface.LLamaServer) (AdventureGame, error) {
	dirname := path.Join(SAVEGAMEDIR, gameid)
	jsonList, errJsonList := getFirstJsonFilesFromDir(dirname)
	if errJsonList != nil {
		return AdventureGame{}, errJsonList
	}
	if len(jsonList) != 1 {
		return AdventureGame{}, fmt.Errorf("invalid number of json files on game dir %s on %v", dirname, len(jsonList))
	}
	return loadAdventure(jsonList[0], llama, &textPromptFormatter)
}

//go:embed webui/gamepageStart.html
var emb_gamepageStart []byte

func PromptEntryToHtml(gameTitle string, savedir string, page AdventurePage, thisPageNumber int, pages int) string {
	var sb strings.Builder

	sb.Write(emb_gamepageStart)

	sb.WriteString("<div id=\"game-container\">\n")

	sb.WriteString("<div class=\"nav-buttons\">")
	if 0 < thisPageNumber {
		sb.WriteString(fmt.Sprintf("<a href=\"%v\" class=\"nav-button\" id=\"back-button\">Back</a>\n", thisPageNumber-1))
	} else {
		sb.WriteString("<button disabled class=\"nav-button\" id=\"back-button\">Back</button>\n")

	}
	if thisPageNumber < pages {
		//sb.WriteString("<button class=\"nav-button\" id=\"forward-button\">Forward</button>\n")
		sb.WriteString(fmt.Sprintf("<a href=\"%v\" class=\"nav-button\" id=\"forward-button\">Forward</a>\n", thisPageNumber+1))
	}
	sb.WriteString("</div>\n\n")

	picfilename := path.Join(savedir, page.PictureFileName())

	byt, errReadPic := os.ReadFile(picfilename)
	if errReadPic != nil {
		fmt.Printf("ERROOR READING PIC |%s|  err:%s\n", picfilename, errReadPic)
	} else {
		sb.WriteString("<img src=\"data:image/png;charset=utf-8;base64,")
		codedString := b64.StdEncoding.EncodeToString(byt)
		sb.WriteString(codedString)
		sb.WriteString("\"></img>\n")
	}

	sb.WriteString("<div id=\"game-text\">\n")
	sb.WriteString(page.Text)
	sb.WriteString("</div>\n\n")

	if len(page.UserResponse) == 0 {
		sb.WriteString("<form method=\"post\">\n")
		sb.WriteString("<input type=\"text\" id=\"command-input\" name=\"userPrompt\" placeholder=\"Enter your command here\" />\n")
		sb.WriteString("<input id=\"submit-button\" type=\"submit\" value=\"Submit\"/>\n")
		sb.WriteString("</form>\n")
		//sb.WriteString(<button id=\"submit-button\">Submit</button></div>")
	} else {
		sb.WriteString("<div id=\"game-text\">\n")
		sb.WriteString(page.UserResponse)
		sb.WriteString("</div>\n\n")
	}

	sb.WriteString(fmt.Sprintf("<details><summary>Summary</summary><pre>%s</pre></details>\n", page.Summary))
	sb.WriteString("<br>\n")
	sb.WriteString(fmt.Sprintf("<details><summary>PicturePrompt</summary><pre>%s</pre></details>\n", page.PictureDescription))

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

func InitWebInterface(imGen *ImageGenerator, llama *llamainterface.LLamaServer) (WebInterface, error) {

	result := WebInterface{
		router: gin.Default(),
		imGen:  imGen,
		llama:  llama,
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

		game, errLoad := loadGameById(c.Param("gameid"), result.llama)
		if errLoad != nil {
			c.Error(errLoad)
			return
		}
		redirectURL := fmt.Sprintf("/savegames/%s/%v", c.Param("gameid"), len(game.Pages))
		fmt.Printf("Asking latest, redirecting to %s\n", redirectURL)
		c.Redirect(http.StatusFound, redirectURL)
	})

	result.router.GET("/savegames/:gameid/:pageNumber", func(c *gin.Context) {
		fmt.Printf("SAVEGAME HAE %s entry %s\n", c.Param("gameid"), c.Param("index"))

		game, errLoad := loadGameById(c.Param("gameid"), result.llama)
		if errLoad != nil {
			c.Error(errLoad)
			return
		}
		pageNumberPar, errPageNumber := strconv.ParseInt(c.Param("pageNumber"), 10, 64)
		if errPageNumber != nil {
			c.Error(fmt.Errorf("invalid pageNumber %s err:%s", c.Param("pageNumber"), errPageNumber))
			return
		}

		if len(game.Pages) == 0 {
			c.Error(fmt.Errorf("no pages on game %s", c.Param("gameid")))
			return
		}

		pageNumber := min(int(pageNumberPar), len(game.Pages)-1)
		page := game.Pages[pageNumber]

		if !game.Textmode {
			tImageCreateStart := time.Now() //aftertought TODO REFACTOR
			errCreatePic := CreatePngIfNotFound(*imGen, path.Join(game.GetSaveDir(), page.PictureFileName()), page.PictureDescription)
			if errCreatePic != nil {
				c.Error(fmt.Errorf("error creating png %s", errCreatePic))
				return
			}
			page.GenerationTimes.Picture = int(time.Since(tImageCreateStart).Milliseconds())
		}

		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(PromptEntryToHtml(game.GameName, game.GetSaveDir(), page, pageNumber, len(game.Pages))))

	})

	result.router.GET("/newgame/", func(c *gin.Context) {
		jsonFileList, errJsonList := ListNewGamesJson(GAMESDIR, llama)
		if errJsonList != nil {
			c.Error(fmt.Errorf("error listing new games %s", errJsonList))
			return
		}

		newcat, newcatErr := CreateToCatalogue(jsonFileList, llama)
		if newcatErr != nil {
			c.Error(fmt.Errorf("error listing new games %s", newcatErr))
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n\t<meta charset=\"UTF-8\">\n\t<meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n\t<meta http-equiv=\"X-UA-Compatible\" content=\"IE=edge\">\n\t<title>Start new game</title>\n\t<link rel=\"stylesheet\" href=\"/gallery.css\">\n</head>\n<body>\n"+newcat.ToHtml()+"</body>\n</html>\n"))
	})

	result.router.GET("/newgame/:gamename", func(c *gin.Context) { //Latest entry on page

		gamename := c.Param("gamename")
		newgamefilename := path.Join(GAMESDIR, gamename+".json")
		fmt.Printf("Going to start %s\n", newgamefilename)
		game, errLoad := loadAdventure(newgamefilename, result.llama, &textPromptFormatter)
		if errLoad != nil {
			c.Error(fmt.Errorf("internal error:%s", errLoad))
			return
		}

		errRun := game.UserInteraction(result.textTemperature, result.imageTemperature, INITIALPROMPT)
		if errRun != nil {
			c.Error(fmt.Errorf("error running:%s", errRun))
			return
		}

		errSave := game.SaveGame()
		if errSave != nil {
			c.Error(fmt.Errorf("error saving game %s", errSave))
			return
		}

		redirectURL := fmt.Sprintf("/savegames/%s", game.GameId())
		fmt.Printf("Asking latest, redirecting to %s\n", redirectURL)
		c.Redirect(http.StatusFound, redirectURL)
	})

	result.router.GET("/continuegame", func(c *gin.Context) { //Get catalog of all savegames
		jsonList, errJsonLlist := ListSavedGamesJson(SAVEGAMEDIR, llama)

		if errJsonLlist != nil {
			c.Error(errJsonLlist)
			return
		}

		cat, catErr := CreateToCatalogue(jsonList, llama)
		if catErr != nil {
			fmt.Printf("TODO ERR %s\n", catErr)
		}

		fmt.Printf("\n\nCONTINUE GAME:%#v\n", cat)

		s := "<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n\t<meta charset=\"UTF-8\">\n\t<meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n\t<meta http-equiv=\"X-UA-Compatible\" content=\"IE=edge\">\n\t<title>Load existing game</title>\n\t<link rel=\"stylesheet\" href=\"gallery.css\">\n</head>\n<body>\n" + catalogLatestToHtml(cat) + "</body>\n</html>\n"
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(s))

	})

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

		game, errLoad := loadGameById(c.Param("gameid"), result.llama)
		if errLoad != nil {
			c.Error(fmt.Errorf("game load err:%s", errLoad))
			return
		}
		errRun := game.UserInteraction(result.textTemperature, result.imageTemperature, s)
		if errRun != nil {
			c.Error(fmt.Errorf("error running:%s", errRun))
			return
		}

		errSave := game.SaveGame()
		if errSave != nil {
			c.Error(fmt.Errorf("error saving game %s", errSave))
			return
		}

		last := game.Pages[len(game.Pages)-1]

		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(PromptEntryToHtml(game.GameName, SAVEGAMEDIR, last, len(game.Pages)-1, len(game.Pages)-1)))
		redirectURL := fmt.Sprintf("/savegames/%s/%v", c.Param("gameid"), len(game.Pages))
		fmt.Printf("Asking latest, redirecting to %s\n", redirectURL)
		c.Redirect(http.StatusFound, redirectURL)
	})

	return result, nil

}

func catalogLatestToHtml(cat GameCatalogue) string {
	var sb strings.Builder

	sb.WriteString("\t<div class=\"gallery\">\n")
	for _, item := range cat {
		sb.WriteString(fmt.Sprintf("<a href=\"/savegames/%s\">", item.Id))
		sb.WriteString("\t\t<div class=\"gallery-item\">\n")

		byt, errReadFile := loadPngToSmallPng(item.MenuPicture, WANTEDTHUMBNAILWIDTH)
		if errReadFile != nil {
			continue //BAD HACK
		}

		sb.WriteString("<img src=\"data:image/png;charset=utf-8;base64,")
		codedString := b64.StdEncoding.EncodeToString(byt)
		sb.WriteString(codedString)
		sb.WriteString("\"></img>\n")

		sb.WriteString(fmt.Sprintf("\t\t\t<div class=\"title\">%s</div>\n", item.Label))
		sb.WriteString("\t\t</div>\n")
		sb.WriteString("</a>\n")
	}
	sb.WriteString("\t\t</div>\n")
	return sb.String()
}

func (p *GameCatalogue) ToHtml() string {
	var sb strings.Builder

	sb.WriteString("\t<div class=\"gallery\">\n")
	for _, item := range *p {
		sb.WriteString(fmt.Sprintf("<a href=\"%s\">", item.Name))
		sb.WriteString("\t\t<div class=\"gallery-item\">\n")
		byt, errReadFile := loadPngToSmallPng("./"+item.TitleImageFileName, WANTEDTHUMBNAILWIDTH)
		if errReadFile != nil {
			fmt.Printf("ERROR LOADING %s  err:%s\n", item.TitleImageFileName, errReadFile)
			continue //BAD HACK
		}

		sb.WriteString("<img src=\"data:image/png;charset=utf-8;base64,")
		codedString := b64.StdEncoding.EncodeToString(byt)
		sb.WriteString(codedString)
		sb.WriteString("\"></img>\n")

		sb.WriteString(fmt.Sprintf("\t\t\t<div class=\"title\">%s</div>\n", item.Name))
		sb.WriteString("\t\t</div>\n")
		sb.WriteString("</a>\n")
	}
	sb.WriteString("\t\t</div>\n")
	return sb.String()
}
