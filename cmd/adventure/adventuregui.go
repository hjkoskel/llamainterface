/*
Adventure GUI. Text based (colors and texcels?) or
*/
package main

import (
	"fmt"
	"math"
	"strings"
	"time"

	imco "image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func numberOfLines(txt string) int {
	return strings.Count(txt, "\n")
}

/*
Graphical UI
*/
func WordWrap(text string, limit int) string {
	//words := strings.Fields(strings.TrimSpace(text)) // Split the text into words
	words := strings.Split(strings.TrimSpace(text), " ")
	if len(words) == 0 {
		return ""
	}

	var wrappedText strings.Builder
	currentLine := ""

	for _, word := range words {
		if len(currentLine)+len(word) > limit {
			wrappedText.WriteString(strings.TrimSpace(currentLine) + "\n")
			currentLine = word + " "
		} else {
			currentLine += word + " "
		}
	}

	if currentLine != "" {
		wrappedText.WriteString(strings.TrimSpace(currentLine))
	}

	return wrappedText.String()
}

const (
	SCREEN_W = 1920
	SCREEN_H = 1080

	MAINFONTSIZE = 30
)

type GraphicalUI struct {
	autoUpdate              bool
	fnt                     rl.Font
	dungeonMasterText       string
	dungeonMasterTextDouble string //A little stupid HACK

	userPromptText  string //This is what ui shows
	imagePromptText string //If no picture, this is shown
	centeredPic     bool   //used for
	pic             rl.Texture2D

	scrollPosition float32
	scrollSpeed    float32

	showPrompt     bool
	generatingText string //Instead of prompt print text generating....

	TextColumnWidth int
	TextSpacing     int
	FontSize        int
	ChosenLanguage  LanguageName
}

// Splash screen.. goes to splash screen mode.
func (p *GraphicalUI) SplashScreen(picFileName string) error {
	if 0 < p.pic.Width {
		rl.UnloadTexture(p.pic)
	}

	p.pic = rl.LoadTexture(picFileName)
	if p.pic.Width == 0 || p.pic.Height == 0 {
		return fmt.Errorf("failed loading %s as texture", picFileName)
	}
	p.centeredPic = true
	return nil
}

func (p *GraphicalUI) SetPage(pictureFileName string, page AdventurePage) {
	p.centeredPic = false
	p.dungeonMasterText = WordWrap(page.GetText(p.ChosenLanguage), p.TextColumnWidth)
	p.dungeonMasterTextDouble = WordWrap(page.GetText(p.ChosenLanguage), p.TextColumnWidth*2)
	p.imagePromptText = WordWrap(page.PictureDescription, p.TextColumnWidth)
	p.scrollPosition = -16 //-SCREEN_H / 3
	p.userPromptText = page.GetUserResponse(p.ChosenLanguage)
	p.SetPicture(pictureFileName)

}

// Set text and do wrapping
/*
func (p *GraphicalUI) SetDungeonMasterText(txt string) {
	p.centeredPic = false
	p.dungeonMasterText = WordWrap(txt, p.TextColumnWidth)
	p.scrollPosition = -SCREEN_H / 3
}
func (p *GraphicalUI) SetImagePromptText(txt string) {
	p.centeredPic = false
	p.imagePromptText = WordWrap(txt, p.TextColumnWidth)
}*/

func (p *GraphicalUI) SetGenerating(txt string) {
	p.generatingText = txt
}

func (p *GraphicalUI) Render() error {
	if rl.WindowShouldClose() {
		return fmt.Errorf("window should close, not possible to draw")
	}

	rl.BeginDrawing()
	rl.ClearBackground(rl.Black)

	if 0 < p.pic.Width && 0 < p.pic.Height {
		rl.DrawTextEx(p.fnt, p.dungeonMasterText, rl.Vector2{X: 0, Y: -p.scrollPosition}, MAINFONTSIZE, 6, imco.RGBA{R: 255, G: 255, B: 255, A: 255})
		if 0 < len(p.imagePromptText) {
			rl.DrawTextEx(p.fnt, p.imagePromptText, rl.Vector2{X: float32(SCREEN_W - p.pic.Width), Y: 0}, MAINFONTSIZE, 6, imco.RGBA{R: 255, G: 255, B: 0, A: 255})
		}
		if p.centeredPic {
			rl.DrawTexture(p.pic, SCREEN_W/2-p.pic.Width/2, 0, rl.White)
		} else {
			rl.DrawTexture(p.pic, SCREEN_W-p.pic.Width, 0, rl.White)
		}
	} else {
		rl.DrawTextEx(p.fnt, p.dungeonMasterTextDouble, rl.Vector2{X: 0, Y: -p.scrollPosition}, MAINFONTSIZE, 6, imco.RGBA{R: 255, G: 255, B: 255, A: 255})
		if 0 < len(p.imagePromptText) {
			rl.DrawTextEx(p.fnt, p.imagePromptText, rl.Vector2{X: float32(SCREEN_W - p.pic.Width), Y: 0}, MAINFONTSIZE, 6, imco.RGBA{R: 255, G: 255, B: 0, A: 255})
		}
	}

	shadedAreaHeight := int32(MAINFONTSIZE * 6)

	if p.showPrompt || 0 < len(p.generatingText) {
		rl.DrawRectangle(0, SCREEN_H-shadedAreaHeight, SCREEN_W, shadedAreaHeight, imco.RGBA{R: 0, G: 0, B: 0, A: 200})
	}

	if p.showPrompt {
		rl.DrawTextEx(p.fnt, "> "+p.userPromptText, rl.Vector2{X: 0, Y: SCREEN_H - MAINFONTSIZE*4}, 25, 2, imco.RGBA{R: 0, G: 255, B: 0, A: 255})
	} else {
		rl.DrawTextEx(p.fnt, p.generatingText, rl.Vector2{X: 0, Y: SCREEN_H - MAINFONTSIZE*4}, 25, 2, imco.RGBA{R: 0, G: 0, B: 255, A: 255})
	}

	rl.EndDrawing()
	return nil
}

const (
	MINSCROLLSPEED = 0.1
	MAXSCROLLSPEED = 6
)

func (p *GraphicalUI) WaitPressAnykey() {
	for !rl.WindowShouldClose() {
		if 0 < rl.GetKeyPressed() {
			return
		}
		p.Render()
	}
}

func (p *GraphicalUI) GetPrompt() (string, error) {
	p.userPromptText = "" //clear old
	p.showPrompt = true

	maxScroll := float32(numberOfLines(p.dungeonMasterText)*MAINFONTSIZE - (SCREEN_H*2)/3)
	minScroll := -SCREEN_H / 3
	autoscroll := true
	p.scrollSpeed = MINSCROLLSPEED
	wasScrolling := true

	prevDelete := time.Now()

	for !rl.WindowShouldClose() {
		renderErr := p.Render()
		if renderErr != nil {
			return "", renderErr
		}

		chr := rl.GetCharPressed()
		for 0 < chr {
			//fmt.Printf("chr:%d as string %s\n", chr, string(chr))
			p.userPromptText += string(chr)
			chr = rl.GetCharPressed()
		}

		if rl.IsKeyDown(rl.KeyBackspace) {
			if 0 < len(p.userPromptText) && time.Millisecond*100 < time.Since(prevDelete) {
				p.userPromptText = p.userPromptText[0 : len(p.userPromptText)-1]
				prevDelete = time.Now()
			}
		}

		if rl.IsKeyDown(rl.KeyUp) {
			p.scrollPosition -= p.scrollSpeed
			if p.scrollPosition < float32(minScroll) {
				p.scrollPosition = float32(minScroll)
			}
			autoscroll = false
		}
		if rl.IsKeyDown(rl.KeyDown) || autoscroll {
			p.scrollPosition += p.scrollSpeed
			//fmt.Printf("SCROLL %v  toRow:%f\n", p.scrollPosition, float64(p.scrollPosition)/float64(MAINFONTSIZE))
			if maxScroll < p.scrollPosition {
				p.scrollPosition = maxScroll
				autoscroll = false
			}
		}

		if rl.IsKeyDown(rl.KeyDown) || rl.IsKeyDown(rl.KeyUp) {
			if wasScrolling {
				p.scrollSpeed += 0.03
				if MAXSCROLLSPEED < p.scrollSpeed {
					p.scrollSpeed = MAXSCROLLSPEED
				}
			}

			wasScrolling = true
		} else {
			p.scrollSpeed = MINSCROLLSPEED
			wasScrolling = false
		}

		if rl.IsKeyPressed(rl.KeyEnter) {
			//Clean up duplicate spaces
			s := ""
			for s != p.userPromptText {
				s = p.userPromptText
				p.userPromptText = strings.ReplaceAll(p.userPromptText, "  ", " ") //replace two spaces with one
			}
			p.userPromptText = strings.TrimSpace(p.userPromptText)
			p.showPrompt = false
			p.Render()                   //refresh immediately...
			return p.userPromptText, nil //Ok, run prompt
		}

		if rl.IsKeyPressed(rl.KeyEscape) {
			p.showPrompt = false
			return "", fmt.Errorf("user pressed escape button. too scared?")
		}

	}
	p.showPrompt = false
	return "", fmt.Errorf("window closed")
}

func (p *GraphicalUI) SetPicture(picFileName string) error { //Pass thru disk.. it is logged anyways...
	if len(picFileName) == 0 {
		p.pic = rl.Texture2D{}
		return nil
	}
	if 0 < p.pic.Width {
		rl.UnloadTexture(p.pic)
	}

	p.pic = rl.LoadTexture(picFileName)
	if p.pic.Width == 0 || p.pic.Height == 0 {
		return fmt.Errorf("failed loading %s as texture", picFileName)
	}
	return nil
}

func (p *GraphicalUI) Close() error {
	rl.CloseWindow()
	return nil
}

func InitGraphicalUI(title string, fontFilename string) (*GraphicalUI, error) {
	rl.InitWindow(SCREEN_W, SCREEN_H, title)
	rl.SetTargetFPS(60)

	if !rl.IsWindowFullscreen() {
		rl.SetWindowSize(SCREEN_W, SCREEN_H)
		rl.ToggleFullscreen()
	}
	fontChars := []rune{}
	for r := rune(32); r <= 255; r++ {
		fontChars = append(fontChars, r)
	}

	// Scandinavian characters
	//fontChars = append(fontChars, []rune{'ö', 'ä', 'å', 'Ö', 'Ä', 'Å', '<', '>', ')', '!', '?', ')', 'ü', 'Ü', '*', '-', '+', ':'}...)

	result := GraphicalUI{
		fnt:             rl.LoadFontEx(fontFilename, MAINFONTSIZE-4, fontChars, int32(len(fontChars))),
		TextColumnWidth: 40,
		TextSpacing:     6,
		FontSize:        MAINFONTSIZE,
		autoUpdate:      true,
	}
	go func() {
		for {
			if !result.showPrompt && result.autoUpdate {
				result.Render() //HACK!
			}
			if rl.IsKeyPressed(rl.KeyEnter) && (rl.IsKeyDown(rl.KeyLeftAlt) || rl.IsKeyDown(rl.KeyRightAlt)) {
				//if rl.IsWindowFullscreen() {
				rl.SetWindowSize(SCREEN_W, SCREEN_H)
				rl.ToggleFullscreen()
				//	}
			}

			time.Sleep(time.Second)
		}
	}()

	return &result, nil
}

const (
	MAINMENUSELECT_NEWGAME   MainMenuSelection = 0
	MAINMENUSELECT_BRANCHOLD MainMenuSelection = 1
	MAINMENUSELECT_CONTINUE  MainMenuSelection = 2
	MAINMENUSELECT_QUIT      MainMenuSelection = 3
)

type MainMenuSelection byte

func (p *GraphicalUI) RunMainMenu() (MainMenuSelection, error) {
	p.centeredPic = false
	p.dungeonMasterText = "** Main menu **\n A:New Game\n B:Branch\n C:Continue\n Q:Quit"
	if 0 < p.pic.Width {
		rl.UnloadTexture(p.pic)
	}

	p.pic = rl.LoadTexture(MENUPICTUREFILE)
	p.autoUpdate = false
	p.generatingText = ""
	for {
		p.Render()
		if rl.IsKeyPressed(rl.KeyA) {
			p.autoUpdate = true
			return MAINMENUSELECT_NEWGAME, nil
		}
		if rl.IsKeyPressed(rl.KeyB) {
			p.autoUpdate = true
			return MAINMENUSELECT_BRANCHOLD, nil
		}
		if rl.IsKeyPressed(rl.KeyC) {
			p.autoUpdate = true
			return MAINMENUSELECT_CONTINUE, nil
		}
		if rl.IsKeyPressed(rl.KeyQ) || rl.IsKeyPressed(rl.KeyEscape) {
			p.autoUpdate = true
			return MAINMENUSELECT_QUIT, nil
		}
	}
}

// Pick from catalogue... spinner???
func (p *GraphicalUI) PickFromCatalogue(cat GameCatalogue) (GameCatalogueEntry, error) {
	chosenIndex := 0       //Target, rolls here
	position := float64(0) //Index but decimal

	if len(cat) == 0 {
		return GameCatalogueEntry{}, fmt.Errorf("no games")
	}

	pictureArr := make([]rl.Texture2D, len(cat))
	for i, c := range cat {
		pictureArr[i] = rl.LoadTexture(c.MenuPicture)
	}

	//Draw
	p.autoUpdate = false
	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		centerPicIndex := int(math.Round(position))
		picw := p.pic.Width
		offset := int32((position-float64(centerPicIndex))*float64(p.pic.Width)) + (picw*3)/2

		//fmt.Printf("draw roller pos:%v chosen %v center:%v offset:%v\n", position, chosenIndex, centerPicIndex, offset)

		if centerPicIndex+1 < len(cat) {
			index := centerPicIndex + 1
			xpos := SCREEN_W + picw - offset
			rl.DrawTexture(pictureArr[index], xpos, 0, rl.White)
			rl.DrawRectangle(xpos, 0, picw, 100, imco.RGBA{R: 0, G: 0, B: 0, A: 230})
			rl.DrawTextEx(p.fnt, cat[index].Label, rl.Vector2{X: float32(xpos), Y: 50}, MAINFONTSIZE, 6, imco.RGBA{R: 255, G: 255, B: 0, A: 255})
		}

		//Center
		if centerPicIndex < len(cat) {
			index := centerPicIndex
			xpos := SCREEN_W - offset
			rl.DrawTexture(pictureArr[index], xpos, 0, rl.White)
			rl.DrawRectangle(xpos, 0, picw, 100, imco.RGBA{R: 0, G: 0, B: 0, A: 230})
			rl.DrawTextEx(p.fnt, cat[index].Label, rl.Vector2{X: float32(xpos), Y: 50}, MAINFONTSIZE, 6, imco.RGBA{R: 255, G: 255, B: 0, A: 255})
		}

		//First
		if 0 < centerPicIndex {
			index := centerPicIndex - 1

			xpos := SCREEN_W - p.pic.Width - offset
			rl.DrawTexture(pictureArr[index], xpos, 0, rl.White)
			rl.DrawRectangle(xpos, 0, picw, 100, imco.RGBA{R: 0, G: 0, B: 0, A: 230})
			rl.DrawTextEx(p.fnt, cat[index].Label, rl.Vector2{X: float32(xpos), Y: 50}, MAINFONTSIZE, 6, imco.RGBA{R: 255, G: 255, B: 0, A: 255})
		}
		rl.EndDrawing()

		//Get key input
		if rl.IsKeyPressed(rl.KeyEnter) {
			p.autoUpdate = true
			return cat[chosenIndex], nil
		}
		if rl.IsKeyPressed(rl.KeyRight) || rl.IsKeyPressed(rl.KeyDown) {
			chosenIndex++
			if len(cat) <= chosenIndex {
				chosenIndex = len(cat) - 1
			}
		}
		if rl.IsKeyPressed(rl.KeyLeft) || rl.IsKeyPressed(rl.KeyUp) {
			chosenIndex--
			if chosenIndex < 0 {
				chosenIndex = 0
			}
		}
		//roll to target... Scale with time?
		delta := float64(chosenIndex) - position
		step := delta * 0.1
		if 0.5 < delta {
			step = delta * 0.4
		}
		position += step
	}
	return cat[0], fmt.Errorf("window should close, not possible to draw")

}
