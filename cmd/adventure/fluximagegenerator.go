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

Acutal server:
https://github.com/stduhpf/stable-diffusion.cpp/tree/server


*/

package main

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"math/rand"
	"net/http"
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

type SampleMethod string

const (
	EULER_A   SampleMethod = "euler_a"
	EULER     SampleMethod = "euler"
	HEUN      SampleMethod = "heun"
	DPM2      SampleMethod = "dpm2"
	DPMPP2S_A SampleMethod = "dpm++2s_a"
	DPMPP2M   SampleMethod = "dpm++2m"
	DPMPP2Mv2 SampleMethod = "dpm++2mv2"
	LCM       SampleMethod = "lcm"
)

func ListAllSamplingMethods() []string {
	return []string{"euler_a", "euler", "heun", "dpm2", "dpm++2s_a", "dpm++2m", "dpm++2mv2", "lcm"}
}

type StableDiffusionCppParameters struct {
	Prompt         string       `json:"prompt"`
	NegativePrompt string       `json:"negative_prompt"`
	ClipSkip       int          `json:"clip_skip"` //-1
	CfgScale       float64      `json:"cfg_scale"` //7.0
	Guidance       float64      `json:"guidance"`
	Width          int          `json:"width"`
	Height         int          `json:"height"`
	SampleMethod   SampleMethod `json:"sample_method"`
	SampleSteps    int          `json:"sample_steps"`
	Seed           int          `json:"seed"`
	BatchCount     int          `json:"batch_count"` //1 is default
	NormalizeInput bool         `json:"normalize_input"`
}

func (p *StableDiffusionCppParameters) SetSampleMethod(m string) error {
	s := strings.ToLower(m)
	lst := ListAllSamplingMethods()
	for _, v := range lst {
		if v == s {
			p.SampleMethod = SampleMethod(s)
			return nil
		}
	}
	return fmt.Errorf("unknown sample method %s", m)
}

func DefaultStableDiffusionCppParameters() StableDiffusionCppParameters {
	return StableDiffusionCppParameters{
		Prompt:         "",
		NegativePrompt: "",
		ClipSkip:       -1,
		CfgScale:       7.0,
		Guidance:       3.5,
		Width:          512,
		Height:         512,
		SampleMethod:   EULER_A,
		SampleSteps:    20,
		BatchCount:     1,
		NormalizeInput: false}
}

/*
IDEA: query is universal
*/
func (p *StableDiffusionCppParameters) ToJSON() string {
	d := DefaultStableDiffusionCppParameters()
	pieces := []string{fmt.Sprintf("\"prompt\":%#v", p.Prompt)}

	if 0 < len(p.NegativePrompt) {
		pieces = append(pieces, fmt.Sprintf("\"negative_prompt\":%#v", p.NegativePrompt))
	}
	if d.ClipSkip != p.ClipSkip {
		pieces = append(pieces, fmt.Sprintf("\"clip_skip\":%#v", p.ClipSkip))
	}
	if d.CfgScale != p.CfgScale {
		pieces = append(pieces, fmt.Sprintf("\"cfg_scale\":%#v", p.CfgScale))
	}
	if d.Guidance != p.Guidance {
		pieces = append(pieces, fmt.Sprintf("\"guidance\":%#v", p.Guidance))
	}
	if d.Width != p.Width {
		pieces = append(pieces, fmt.Sprintf("\"width\":%#v", p.Width))
	}
	if d.Height != p.Height {
		pieces = append(pieces, fmt.Sprintf("\"height\":%#v", p.Height))
	}
	if d.SampleMethod != p.SampleMethod {
		pieces = append(pieces, fmt.Sprintf("\"sample_method\":%#v", p.SampleMethod))
	}
	if p.Seed != 0 {
		pieces = append(pieces, fmt.Sprintf("\"seed\":%#v", p.Seed))
	}
	if d.SampleSteps != p.SampleSteps {
		pieces = append(pieces, fmt.Sprintf("\"sample_steps\":%#v", p.SampleSteps))
		pieces = append(pieces, fmt.Sprintf("\"steps\":%#v", p.SampleSteps))
		pieces = append(pieces, fmt.Sprintf("\"sampling_steps\":%#v", p.SampleSteps))
	}
	if d.BatchCount != p.BatchCount {
		pieces = append(pieces, fmt.Sprintf("\"batch_count\":%#v", p.BatchCount))
	}
	if d.NormalizeInput != p.NormalizeInput {
		pieces = append(pieces, fmt.Sprintf("\"normalize_input\":%#v", p.NormalizeInput))
	}
	return "{" + strings.Join(pieces, ",") + "}"
}

type FluxImageGenerator struct {
	Url        string
	Parameters StableDiffusionCppParameters //Set on program start
}

type Txt2ImgResponse struct {
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Channel  int    `json:"channel"`
	Data     string `json:"data"`
	Encoding string `json:"encoding"`
}

func (p *FluxImageGenerator) CreatePic(prompt string, negativePrompt string) (image.Image, error) {

	p.Parameters.Prompt = prompt
	p.Parameters.NegativePrompt = negativePrompt
	p.Parameters.Seed = rand.Int() //parameter?
	//p.Parameters.Width = 1024
	//p.Parameters.Height = 1024
	v := p.Parameters.ToJSON()
	fmt.Printf("Inputti %s\n", v)
	request, errRequesting := http.NewRequest("POST", p.Url, bytes.NewBuffer([]byte(v)))
	if errRequesting != nil {
		return nil, errRequesting
	}
	request.Header.Set("Content-Type", "text/json; charset=UTF-8")

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

	body, errBody := io.ReadAll(response.Body)
	if errBody != nil {
		return nil, errBody
	}

	var resp []Txt2ImgResponse
	errUnmarshal := json.Unmarshal(body, &resp)
	if errUnmarshal != nil {
		return nil, errUnmarshal
	}

	baseBin, baseErr := b64.StdEncoding.DecodeString(string(resp[0].Data))
	if baseErr != nil {
		return nil, baseErr
	}

	return png.Decode(bytes.NewBuffer(baseBin))
}
