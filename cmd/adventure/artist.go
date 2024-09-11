/*
Prompt generator for generating art for game
use player and
*/
package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/png"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/hjkoskel/bindstablediff"
)

func LoadPng(fname string) (image.Image, error) {
	f, errOpen := os.Open(fname)
	if errOpen != nil {
		return nil, fmt.Errorf("error opening png file %s err:%s", fname, errOpen)
	}
	result, errDecode := png.Decode(f)
	if errDecode != nil {
		return result, fmt.Errorf("error decoding png data on %s err:%s", fname, errDecode)
	}
	return result, f.Close()
}

func SavePng(fname string, img image.Image) error {
	f, err := os.Create(fname)
	if err != nil {
		fmt.Printf("create image fail %v", err)
	}
	if err := png.Encode(f, img); err != nil {
		return fmt.Errorf("error writing %s failed err=%v", fname, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("error writing %s failed err=%v", fname, err)
	}
	return nil
}

type ImageGenerator interface {
	CreatePic(prompt string) (image.Image, error)
}

type DiffusonImageGenerator struct {
	model bindstablediff.StableDiffusionModel
	Pars  bindstablediff.TextGenPars
}

func InitDiffusionImageGen(diffusionModelFile string, steps int, nThreads int, schedule bindstablediff.EnumSchedule) (ImageGenerator, error) {
	g := DiffusonImageGenerator{
		Pars: bindstablediff.TextGenPars{
			Prompt:         "", //Set
			NegativePrompt: "",
			CfgScale:       7,
			Width:          512,
			Height:         512,
			SampleMethod:   bindstablediff.EULER,
			SampleSteps:    steps,
			Seed:           -1},
	}

	var errInit error
	g.model, errInit = bindstablediff.InitStableDiffusion(diffusionModelFile, nThreads, schedule)
	if errInit != nil {
		return nil, fmt.Errorf("error initialized %v\n", errInit)
	}
	return &g, nil
}

func (p *DiffusonImageGenerator) CreatePic(prompt string) (image.Image, error) {
	p.Pars.Prompt = prompt
	return p.model.Txt2Img(p.Pars)
}

type FluxImageGenerator struct {
	Host string
	Port int
}

func (p *FluxImageGenerator) CreatePic(prompt string) (image.Image, error) {
	u, uErr := url.JoinPath(fmt.Sprintf("http://%s:%v", p.Host, p.Port), "/generate")
	if uErr != nil {
		return nil, uErr
	}

	request, errRequesting := http.NewRequest("POST", u, bytes.NewBuffer([]byte(prompt)))
	if errRequesting != nil {
		return nil, errRequesting
	}
	request.Header.Set("Content-Type", "text/plain; charset=UTF-8")

	client := &http.Client{}
	client.Timeout = time.Minute * 10 //TODO PARAMETRIZE OR CONSTANT
	response, errDo := client.Do(request)
	if errDo != nil {
		return nil, fmt.Errorf("error while flux.1 request %s", errDo.Error())
	}
	defer response.Body.Close()

	fmt.Println("response Status:", response.Status)
	fmt.Println("response Headers:", response.Header)

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Flux.1 generate query returned with code %s  %s ", response.StatusCode, response.Status)
	}

	return png.Decode(response.Body)
}

func InitFluxImageGen(host string, port int) (ImageGenerator, error) { //TODO instead of separate server... create flux.1 library?
	g := &FluxImageGenerator{Host: host, Port: port}
	return ImageGenerator(g), nil
}

/*
go:embed promptArtist.txt
var baseTextPromptArtist string
*/

type Artist struct {
	BaseText string
	//model    bindstablediff.StableDiffusionModel
	Prefix string //Possible to modify image generation prompt by some fixed part of prompt. Like style stays same thru whole game
	Suffix string

	Gen ImageGenerator
	//Pars bindstablediff.TextGenPars
}

func InitArtist(diffusionModelFile string, baseText string, prefix string, suffix string, gen ImageGenerator) (Artist, error) {
	return Artist{
		BaseText: baseText,
		Prefix:   prefix,
		Suffix:   suffix,
		Gen:      gen,
	}, nil
}

/*
func InitArtist(diffusionModelFile string, prefix string, suffix string, steps int, nThreads int, schedule bindstablediff.EnumSchedule) (Artist, error) {
	result := Artist{
		BaseText: baseTextPromptArtist,
		Prefix:   prefix,
		Suffix:   suffix,
		Pars: bindstablediff.TextGenPars{
			Prompt:         "", //Set
			NegativePrompt: "",
			CfgScale:       7,
			Width:          512,
			Height:         512,
			SampleMethod:   bindstablediff.EULER,
			SampleSteps:    steps,
			Seed:           -1},
	}
	var errInit error

	result.model, errInit = bindstablediff.InitStableDiffusion(diffusionModelFile, nThreads, schedule)
	if errInit != nil {
		return result, fmt.Errorf("error initialized %v\n", errInit)
	}
	return result, nil
}
*/

func (p *Artist) ArtPromptText(longtext string) string {
	return strings.TrimSpace(p.BaseText) + "\n\n" + longtext + "<|eot_id|><|start_header_id|>dungeonmaster<|end_header_id|>"
}

func (p *Artist) CreatePic(prompt string) (image.Image, error) {
	return p.Gen.CreatePic(prompt)
	/*
		p.Pars.Prompt = prompt
		return p.model.Txt2Img(p.Pars)
	*/
}
