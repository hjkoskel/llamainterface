/*

This module should have structs for converting image prompt to query.
It is not yet clear what kind of model is used for flux.1 interference. Or is flux right choice at first place


Each third party service have its own set of parameters for generating pictures in flux.1
https://docs.together.ai/reference/post_images-generations



Black forest lab API does not have negative prompt
https://api.bfl.ml/scalar#tag/tasks/POST/v1/flux-dev

Also no negative prompt on diffusers (faster stable-diffusion.cpp, not possibe to run 100% on CPU because bug on some dependencies)
https://github.com/huggingface/diffusers/blob/main/src/diffusers/pipelines/flux/pipeline_flux.py

Also tinygrad implemention does not have it (not tested, got  __truncsfbf2 error, clang.  I do not have good enough GPU)
https://github.com/tinygrad/tinygrad/blob/master/examples/flux1.py


stable-diffussion.cpp have negative prompt
https://github.com/leejet/stable-diffusion.cpp/blob/master/docs/flux.md
https://github.com/leejet/stable-diffusion.cpp/blob/master/examples/cli/main.cpp

*/

package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TODO are going to use these?
type ImgGenTogetherAiParameters struct {
	Prompt         string `json:"prompt"`
	Model          string `json:"model"`
	Steps          int    `json:"steps"` //20 default
	Seed           int    `json:"seed"`
	N              int    `json:"n"`      //1, how many images to generate
	Height         int    `json:"height"` //1024
	Width          int    `json:"width"`  //1024
	NegativePrompt string `json:"negative_prompt"`
}

type BflGenParameters struct {
	Prompt           string  `json:"prompt"`
	Width            int     `json:"width"`             //1024,
	Height           int     `json:"height"`            // 768,
	Steps            int     `json:"steps"`             // 28,
	Promptupsampling bool    `json:"prompt_upsampling"` // false, TODO how to implement?
	Seed             int64   `json:"seed"`              // 42,
	Guidance         float64 `json:"guidance"`          // 3,  1.5-6 Guidance scale for image generation. High guidance scales improve prompt adherence at the cost of reduced realism.
	SafetyTolerance  int     `json:"safety_tolerance"`  //2 // Tolerance level for input and output moderation. Between 0 and 6, 0 being most strict, 6 being least strict.  (ok lets use 6 instead)
}

/*
type DiffusersImageGenParameters struct{
	string `json:"prompt: Union = None
	string `json:"prompt_2: Union = None
	int `json:"height: Optional = None
	int `json:"width: Optional = None
	int `json:"num_inference_steps: int = 28
	[]int `json:"timesteps: List = None
	float64 `json:"guidance_scale: float = 3.5
	int `json:"num_images_per_prompt: Optional = 1
	//`json:"generator: Union = None
	//`json:"latents: Optional = None
	//`json:"prompt_embeds: Optional = None
	//`json:"pooled_prompt_embeds: Optional = None
	//`json:"output_type: Optional = 'pil'
	//`json:"return_dict: bool = True
	//`json:"joint_attention_kwargs: Optional = None
	//`json:"callback_on_step_end: Optional = None
	//`json:"callback_on_step_end_tensor_inputs: List = ['latents']
	//`json:"max_sequence_length: int = 512
}
*/

// --------- image generator

type FluxImageGenerator struct {
	Host string
	Port int
}

func (p *FluxImageGenerator) CreatePic(prompt string) (image.Image, error) {
	u, uErr := url.JoinPath(fmt.Sprintf("http://%s:%v", p.Host, p.Port), "/generate")
	if uErr != nil {
		return nil, uErr
	}

	request, errRequesting := http.NewRequest("POST", u, bytes.NewBuffer([]byte(strings.ReplaceAll(prompt, "\n", " "))))
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
		return nil, fmt.Errorf("Flux.1 generate query returned with code %v  %s ", response.StatusCode, response.Status)
	}

	return png.Decode(response.Body)
}

func InitFluxImageGen(host string, port int) (ImageGenerator, error) { //TODO instead of separate server... create flux.1 library?
	g := &FluxImageGenerator{Host: host, Port: port}
	return ImageGenerator(g), nil
}
