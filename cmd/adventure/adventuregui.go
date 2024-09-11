/*
Adventure GUI. Text based (colors and texcels?) or
*/
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	imco "image/color"

	"github.com/fatih/color"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type TextUi struct {
	rd *bufio.Reader
}

func InitTextUI() TextUi {
	return TextUi{rd: bufio.NewReader(os.Stdin)}
}

func (p *TextUi) PrintDungeonMaster(txt string) {
	//fmt.Printf("%s", txt)
	color.Green(fmt.Sprintf("%s", txt))

}
func (p *TextUi) PrintUser(txt string) {
	color.Cyan(fmt.Sprintf("%s", txt))
}

func (p *TextUi) GetUserInput() (string, error) {
	p.PrintUser("user> ")
	return p.rd.ReadString('\n')
}

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

	MAINFONTSIZE = 20
)

type GraphicalUI struct {
	fnt               rl.Font
	dungeonMasterText string
	userPromptText    string //This is what ui shows
	imagePromptText   string //If no picture, this is shown
	pic               rl.Texture2D

	scrollPosition float32
	scrollSpeed    float32

	showPrompt     bool
	generatingText string //Instead of prompt print text generating....

	TextColumnWidth int
	TextSpacing     int
	FontSize        int
}

// Set text and do wrapping
func (p *GraphicalUI) SetDungeonMasterText(txt string) {
	p.dungeonMasterText = WordWrap(txt, p.TextColumnWidth)
	p.scrollPosition = -SCREEN_H / 3
}
func (p *GraphicalUI) SetImagePromptText(txt string) {
	p.imagePromptText = WordWrap(txt, p.TextColumnWidth)
}

func (p *GraphicalUI) SetGenerating(txt string) {
	p.generatingText = txt
}

func (p *GraphicalUI) Render() error {
	if rl.WindowShouldClose() {
		return fmt.Errorf("window should close, not possible to draw")
	}

	rl.BeginDrawing()
	rl.ClearBackground(rl.Black)
	rl.DrawTextEx(p.fnt, p.dungeonMasterText, rl.Vector2{X: 0, Y: -p.scrollPosition}, MAINFONTSIZE, 6, imco.RGBA{R: 255, G: 255, B: 255, A: 255})
	if 0 < len(p.imagePromptText) {
		rl.DrawTextEx(p.fnt, p.imagePromptText, rl.Vector2{X: float32(SCREEN_W - p.pic.Width), Y: 0}, MAINFONTSIZE, 6, imco.RGBA{R: 255, G: 255, B: 0, A: 255})
	}
	if 0 < p.pic.Width && 0 < p.pic.Height {
		rl.DrawTexture(p.pic, SCREEN_W-p.pic.Width, 0, rl.White)
	}

	shadedAreaHeight := int32(MAINFONTSIZE * 6)
	rl.DrawRectangle(0, SCREEN_H-shadedAreaHeight, SCREEN_W, shadedAreaHeight, imco.RGBA{R: 0, G: 0, B: 0, A: 200})

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

func (p *GraphicalUI) GetPrompt() (string, error) {
	p.userPromptText = "" //clear old
	p.showPrompt = true

	maxScroll := float32(numberOfLines(p.dungeonMasterText)*MAINFONTSIZE - (SCREEN_H*2)/3)
	minScroll := -SCREEN_H / 3
	autoscroll := true
	p.scrollSpeed = MINSCROLLSPEED
	wasScrolling := true

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
		if rl.IsKeyPressed(rl.KeyBackspace) {
			if 0 < len(p.userPromptText) {
				p.userPromptText = p.userPromptText[0 : len(p.userPromptText)-1]
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

func InitGraphicalUI(title string) (*GraphicalUI, error) {
	rl.InitWindow(SCREEN_W, SCREEN_H, title)
	rl.SetTargetFPS(60)

	result := GraphicalUI{
		fnt:             rl.LoadFont("pixantiqua.ttf"),
		TextColumnWidth: 45,
		TextSpacing:     6,
		FontSize:        MAINFONTSIZE,
	}
	go func() {
		for {
			if !result.showPrompt {
				result.Render() //HACK!
				time.Sleep(time.Second)
			}
		}
	}()

	return &result, nil
}
