/*
Prompt generator for generating art for game
use player and
*/
package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"

	"path"
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
	os.MkdirAll(path.Dir(fname), 0777)
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

func CreatePngIfNotFound(gen ImageGenerator, fname string, prompt string, negativePrompt string) error {
	byt, errRead := os.ReadFile(fname)
	if errRead == nil {
		_, errDecode := png.Decode(bytes.NewBuffer(byt))
		if errDecode == nil {
			return nil //OK
		}

	}
	img, imgErr := gen.CreatePic(prompt, negativePrompt)
	if imgErr != nil {
		return fmt.Errorf("img generation error %s", imgErr)
	}
	return SavePng(fname, img)
}

type ImageGenerator interface {
	CreatePic(prompt string, negativePrompt string) (image.Image, error)
}

/*
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
*/
