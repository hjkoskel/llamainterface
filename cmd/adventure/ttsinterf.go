/*
ttsinterf
*/
package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

// Game requires to save sound file anyways on disk
type TTSInterf interface {
	Run(text string, outputFilename string) error //TODO common wave format?
}

func CleanForTTS(s string) string {
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\\n", "")
	s = strings.ReplaceAll(s, "\"", "")
	return s
}

type TTSType string

const (
	TTSTYPE_NO    TTSType = ""
	TTSTYPE_Piper TTSType = "piper"
	TTSTYPE_Flite TTSType = "flite"
	TTSTYPE_Bark  TTSType = "bark"
)

func GetAllowedTTStypes() []string {
	return []string{string(TTSTYPE_Piper), string(TTSTYPE_Flite), string(TTSTYPE_Bark)}
}

func ChooseTTS(pTTStype string, executableFile string, model string, configFile string) (TTSInterf, error) {

	switch TTSType(strings.ToLower(pTTStype)) {
	case TTSTYPE_Piper:
		a := PiperDefault(executableFile, model, configFile)
		return &a, nil
	case TTSTYPE_Flite:
		a := Flite{
			Executablename: executableFile, //Can be empty
			Voice:          model}
		return &a, nil
	case TTSTYPE_Bark:
		a := Bark{
			ModelFilename:      model,
			ExecutableFileName: executableFile,
		}
		return &a, nil
	case TTSTYPE_NO:
		return &NoTTS{}, nil
	}

	return &NoTTS{}, fmt.Errorf("Unknown tts type: %s", pTTStype)
}

type NoTTS struct {
}

func CreateTTSAudioIfNotFound(gen TTSInterf, fname string, text string) error {
	fmt.Printf("\nCREATRE AUDIO IF %s is not found\n", fname)
	_, err := os.Stat(fname)
	if err == nil {
		return nil //Exists
	}
	return gen.Run(text, fname)
}

func (p *NoTTS) Run(text string, outputFilename string) error {
	fmt.Printf("NO TTS: to file %s\n", outputFilename)
	return nil
}

/***********
** PIPER
************
usage: ./piper [options]

options:
   -h        --help              show this message and exit
   -m  FILE  --model       FILE  path to onnx model file
   -c  FILE  --config      FILE  path to model config file (default: model path + .json)
   -f  FILE  --output_file FILE  path to output WAV file ('-' for stdout)
   -d  DIR   --output_dir  DIR   path to output directory (default: cwd)
   --output_raw                  output raw audio to stdout as it becomes available
   -s  NUM   --speaker     NUM   id of speaker (default: 0)
   --noise_scale           NUM   generator noise (default: 0.667)
   --length_scale          NUM   phoneme length (default: 1.0)
   --noise_w               NUM   phoneme width noise (default: 0.8)
   --sentence_silence      NUM   seconds of silence after each sentence (default: 0.2)
   --espeak_data           DIR   path to espeak-ng data directory
   --tashkeel_model        FILE  path to libtashkeel onnx model (arabic)
   --json-input                  stdin input is lines of JSON instead of plain text
   --debug                       print DEBUG messages to the console
*/

const (
	PIPERDEFAULT_NOISESCALE      = 0.667
	PIPERDEFAULT_LENGTHSCALE     = 1.0
	PIPERDEFAULT_NOISEW          = 0.8
	PIPERDEFAULT_SENTENCESILENCE = 0.2
)

type Piper struct {
	PiperExecutableFileName string
	ModelFile               string
	ConfigFile              string // (default: model path + .json)
	// -f ot
	//-d, -s
	NoiseScale      float64
	LengthScale     float64
	NoiseW          float64
	SentenceSilence float64
	EspeakDataDir   string
	TashkeelModel   string
	//Depends on function call JsonInput       bool
}

func PiperDefault(executableFile string, modelFile string, configFile string) Piper {
	return Piper{
		PiperExecutableFileName: executableFile,
		ModelFile:               modelFile,
		ConfigFile:              configFile,
		NoiseScale:              PIPERDEFAULT_NOISESCALE,
		LengthScale:             PIPERDEFAULT_LENGTHSCALE,
		NoiseW:                  PIPERDEFAULT_NOISEW,
		SentenceSilence:         PIPERDEFAULT_SENTENCESILENCE}
}

func (p *Piper) commandArgs() ([]string, error) {
	result := []string{}
	if len(p.ModelFile) == 0 && len(p.TashkeelModel) == 0 {
		return nil, fmt.Errorf("model not set")
	}
	if 0 < len(p.ModelFile) {
		result = append(result, []string{"-m", p.ModelFile}...)
	}
	if 0 < len(p.ConfigFile) {
		result = append(result, []string{"-c", p.ConfigFile}...)
	}
	if p.NoiseScale != PIPERDEFAULT_NOISESCALE {
		result = append(result, []string{"--noise_scale", fmt.Sprintf("%.4", p.NoiseScale)}...)
	}
	if p.LengthScale != PIPERDEFAULT_LENGTHSCALE {
		result = append(result, []string{"--length_scale", fmt.Sprintf("%.4", p.LengthScale)}...)
	}

	if p.NoiseW != PIPERDEFAULT_NOISEW {
		result = append(result, []string{"--noise_w", fmt.Sprintf("%.4", p.NoiseScale)}...)
	}

	if p.SentenceSilence != PIPERDEFAULT_SENTENCESILENCE {
		result = append(result, []string{"--sentence_silence", fmt.Sprintf("%.4", p.SentenceSilence)}...)
	}

	if 0 < len(p.EspeakDataDir) {
		result = append(result, []string{"--espeak_data ", fmt.Sprintf("%.4", p.EspeakDataDir)}...)
	}
	if 0 < len(p.TashkeelModel) {
		result = append(result, []string{"--tashkeel_model ", fmt.Sprintf("%.4", p.TashkeelModel)}...)
	}
	return result, nil
}

func ExecWithStdinInput(exefile string, args []string, input string) error {
	cmd := exec.Command(exefile, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, input+"\n")
	}()

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", out)

	cmd.Wait()
	return nil
}

func (p *Piper) Run(text string, outputFilename string) error {
	fmt.Printf("running piper\n")
	args, argsErr := p.commandArgs()
	if argsErr != nil {
		return argsErr
	}

	args = append(args, []string{"-f", outputFilename}...)
	if len(p.PiperExecutableFileName) == 0 {
		p.PiperExecutableFileName = "piper"
	}
	return ExecWithStdinInput(p.PiperExecutableFileName, args, text)
}

/*********
*  FLITE
**********/
type Flite struct {
	Executablename string
	Voice          string //NAME Use voice NAME (NAME can be pathname/url to flitevox file)
	Voicedir       string //NAME Directory containing (clunit) voice data

	//TODO what are features?
	IntFeatures    map[string]int     //
	FloatFeatures  map[string]float64 // --setf	F=V
	StringFeatures map[string]string
}

func (p *Flite) commandArgs() ([]string, error) {
	result := []string{}

	if 0 < len(p.Voice) {
		result = append(result, []string{"-voice", p.Voice}...)
	}
	if 0 < len(p.Voicedir) {
		result = append(result, []string{"-voicedir", p.Voicedir}...) //TODO check missing values
	}
	for f, v := range p.IntFeatures {
		result = append(result, []string{f, fmt.Sprintf("%v", v)}...)
	}
	for f, v := range p.FloatFeatures {
		result = append(result, []string{f, fmt.Sprintf("%.4f", v)}...)
	}
	for f, v := range p.StringFeatures {
		result = append(result, []string{f, fmt.Sprintf("%s", v)}...)
	}
	return result, nil
}

func (p *Flite) Run(text string, outputFilename string) error {
	if p.Executablename == "" {
		p.Executablename = "flite" //is in path
	}

	args, argsErr := p.commandArgs()
	if argsErr != nil {
		return argsErr
	}

	args = append(args, []string{"-o", outputFilename}...)
	return ExecWithStdinInput(p.Executablename, args, text)
}

/*********
* Bark main program
**********/
/*
-h, --help            show this help message and exit

	-t N, --threads N     number of threads to use during computation (default: 4)
	-s N, --seed N        seed for random number generator (default: 0)
	-p PROMPT, --prompt PROMPT
	                      prompt to start generation with (default: random)
	-m FNAME, --model FNAME
	                      model path (default: ./ggml_weights)
	-o FNAME, --outwav FNAME
	                      output generated wav (default: output.wav
*/
type Bark struct {
	Threads            int
	Seed               int
	ModelFilename      string
	ExecutableFileName string
}

func (p *Bark) commandArgs() ([]string, error) {
	result := []string{}
	if 0 < p.Threads && p.Threads != 4 {
		result = append(result, []string{"-t", fmt.Sprintf("%v", p.Threads)}...)
	}
	if p.Seed != 0 {
		result = append(result, []string{"-s", fmt.Sprintf("%v", p.Seed)}...)
	}
	if p.ModelFilename != "" {
		result = append(result, []string{"-m", p.ModelFilename}...)
	}
	return result, nil
}

func (p *Bark) Run(text string, outputFilename string) error {
	fmt.Printf("running bark\n")
	args, argsErr := p.commandArgs()
	if argsErr != nil {
		return argsErr
	}
	args = append(args, []string{"-o", outputFilename}...)
	args = append(args, []string{"-p", fmt.Sprintf("\"%s\"", CleanForTTS(text))}...)

	fmt.Printf("bark args \n%s\n%s\n", p.ExecutableFileName, strings.Join(args, " "))

	if len(p.ExecutableFileName) == 0 {
		p.ExecutableFileName = "bark"
	}

	cmd := exec.Command(p.ExecutableFileName, args...)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("ERR but output is %s\n", stdoutStderr)
		return err
	}
	fmt.Printf("%s\n", stdoutStderr)
	return nil
	//return ExecWithStdinInput(p.ExecutableFileName, args, text)
}
