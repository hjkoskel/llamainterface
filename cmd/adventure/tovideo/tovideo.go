/*
tovideo
Creates video from pictures and sound. Also subtitles

This is kept on separate module. Experimental, not sure is this feature needed
*/
package tovideo

import (
	"bytes"
	"fmt"
	"image"
	"image/color"

	"image/png"
	"math"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"golang.org/x/image/draw"
)

const (
	SRTMAXROWS = 18 //3
	STRMAXLEN  = 50 // 47
)

func getMediaLen(fname string) (time.Duration, error) {
	args := []string{"-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", fname}
	cmd := exec.Command("ffprobe", args...)
	stdoutput, err := cmd.CombinedOutput()
	if err != nil {
		return time.Duration(0), fmt.Errorf("getMediaLen on %s failed %s", fname, err)
	}
	s := strings.TrimSpace(string(stdoutput))
	sec, errConv := strconv.ParseFloat(s, 64)
	if errConv != nil {
		return time.Duration(0), fmt.Errorf("error parsing %s  err:%s", s, errConv)
	}

	return time.Duration(sec*1000*1000) * time.Microsecond, nil

}

func runffmpeg(args []string) error {
	cmd := exec.Command("ffmpeg", args...)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("ERR %s but output is %s\n", err, stdoutStderr)
		return err
	}
	fmt.Printf("%s\n", stdoutStderr)
	return nil
}

func createSilentAudioArg(dur time.Duration, outputFilename string) []string {
	return []string{"-f", "lavfi", "-t", fmt.Sprintf("%v", dur.Seconds()), "-i", "anullsrc=channel_layout=stereo:sample_rate=44100", outputFilename}
}

func createSilentAudio(dur time.Duration, outputFilename string) error {
	//args := []string{"-f", "lavfi", "-t", fmt.Sprintf("%v", dur.Seconds()), "-i", "anullsrc=channel_layout=stereo:sample_rate=44100", outputFilename}
	//fmt.Printf("\n\n-----CREATE SILENT AUDIO---\n%s\n\n", strings.Join(args, " "))
	os.Remove(outputFilename)
	return runffmpeg(createSilentAudioArg(dur, outputFilename))
}

const (
	BG_X = 1920
	BG_Y = 1080
)

// Fixed resolution? blinear scaling to 1920x1080 -> scale to 1024x1024 and place on right side (AI generated pics are power of two)
func createBgImage(imgInputFileName string, outputInputFilename string) error {
	f, errOpen := os.Open(imgInputFileName)
	if errOpen != nil {
		return fmt.Errorf("error opening png file %s err:%s", imgInputFileName, errOpen)
	}
	img, errDecode := png.Decode(f)
	if errDecode != nil {
		return fmt.Errorf("error decoding png data on %s err:%s", imgInputFileName, errOpen)
	}
	//TODO TÄSSÄ 1024x1024
	result := image.NewRGBA(image.Rect(0, 0, BG_X, BG_Y))

	//draw.Draw(result, image.Rect(BG_X-img.Bounds().Dx(), 0, BG_X, BG_Y), img, image.Point{X: 0}, draw.Over)

	//Ajatus:
	kernel := draw.BiLinear
	scale := float64(BG_Y) / float64(img.Bounds().Dy()) //reserved 1080x1080 area, keep ratio same
	scale = math.Min(scale, float64(BG_Y)/float64(img.Bounds().Dx()))
	imgW := int(float64(img.Bounds().Dx()) * scale)
	imgH := int(float64(img.Bounds().Dy()) * scale)
	fmt.Printf("scale %.3f  -> image %v X %v\n", scale, imgW, imgH)

	kernel.Scale(result, image.Rect(BG_X-imgW, 0, BG_X, BG_Y), img, img.Bounds(), draw.Over, nil)

	/*
		//////////////////////
		ratio := float64(wantedWidth) / float64(img.Bounds().Dx())
		result := image.NewRGBA(image.Rect(0, 0, wantedWidth, int(float64(img.Bounds().Dx())*ratio)))

		kernel := draw.BiLinear
		kernel.Scale(result, result.Rect, img, img.Bounds(), draw.Over, nil)

		bw := bytes.NewBuffer(nil)
		errEncode := png.Encode(bw, result)
		return bw.Bytes(), errEncode*/

	bw := bytes.NewBuffer(nil)
	errEncode := png.Encode(bw, result)
	if errEncode != nil {
		return errEncode
	}

	return os.WriteFile(outputInputFilename, bw.Bytes(), 0666)
}

func createImageSoundVideoArg(audioFilename string, imageFilename string, keyframesList string, outputFilename string) []string {
	//https://stackoverflow.com/questions/30979714/change-keyframe-interval
	return []string{
		"-loop", "1",
		"-i", imageFilename,
		"-i", audioFilename,
		"-vf", "scale=1920:1080", //,pad=1920:1080:1900:0",
		"-c:v", "libx264",
		"-force_key_frames", keyframesList,
		//"-x264-params", "keyint=60:scenecut=0",
		"-tune", "stillimage",
		"-c:a", "aac",
		"-b:a", "192k",
		"-pix_fmt", "yuv420p",
		"-shortest", outputFilename}

}

func createImageSoundVideo(audioFilename string, imageFilename string, keyframesList string, outputFilename string) error {
	os.Remove(outputFilename)
	return runffmpeg(createImageSoundVideoArg(audioFilename, imageFilename, keyframesList, outputFilename))
}

func JoinVideos(fnames []string, segmentsFilename string, outputFilename string) error {
	//keyframescount := 30  -g ?
	os.Remove(outputFilename)
	segmentsContent := "file '" + strings.Join(fnames, "'\nfile '") + "'\n"
	errWriteSegments := os.WriteFile(segmentsFilename, []byte(segmentsContent), 0666)
	if errWriteSegments != nil {
		return fmt.Errorf("error creating segments file %s  err:%s", segmentsFilename, errWriteSegments)
	}
	return runffmpeg([]string{"-f", "concat", "-safe", "0", "-i", segmentsFilename, "-c", "copy", outputFilename})
}

// https://ikyle.me/blog/2020/add-mp4-chapters-ffmpeg
type Chapter struct {
	Title string
	Start time.Duration
	End   time.Duration
}

func (p *Chapter) ToText() string {
	return fmt.Sprintf("[CHAPTER]\nTIMEBASE=1/1000\nSTART=%v\nEND=%v\ntitle=%s\n\n", p.Start.Milliseconds(), p.End.Milliseconds(), p.Title)
}

type Chapters []Chapter

func (p *Chapters) WriteChaptersToVideo(inputVideoFilename string, outputFilename string) error {
	tmpTxt, errTxt := GetTmpFileNames(1, "txt")
	if errTxt != nil {
		return fmt.Errorf("error creating tmp text file %s", errTxt)
	}

	wErr := os.WriteFile(tmpTxt[0], []byte(p.ToText()), 0666)
	if wErr != nil {
		return fmt.Errorf("error writing chapters %s", wErr)
	}

	args := []string{
		"-i", inputVideoFilename,
		"-i", tmpTxt[0],
		"-map_metadata", "1",
		"-codec", "copy", outputFilename}

	os.Remove(outputFilename)
	errRun := runffmpeg(args)
	RemoveFileList(tmpTxt)
	return errRun
	//ffmpeg
}

func (p *Chapters) ToText() string {
	var sb strings.Builder
	for _, ch := range *p {
		sb.WriteString(ch.ToText())
	}
	return sb.String()
}

func CreateChapters(titles []string, durations []time.Duration) (Chapters, error) {
	if len(titles) != len(durations) {
		return nil, fmt.Errorf("there are %v titles but %v durations", len(titles), len(durations))
	}
	result := make([]Chapter, len(titles))
	var dur time.Duration
	for i, title := range titles {
		result[i] = Chapter{
			Title: title,
			Start: dur,
			End:   dur + durations[i]}
		dur += durations[i]
	}
	return result, nil
}

func CreateChaptersFromFiles(videofiles []string, prefix string) (Chapters, error) {
	titles := make([]string, len(videofiles))
	durations := make([]time.Duration, len(videofiles))
	var errDur error
	for i, fname := range videofiles {
		titles[i] = fmt.Sprintf("%s%v", prefix, i)
		durations[i], errDur = getMediaLen(fname)
		if errDur != nil {
			return Chapters{}, fmt.Errorf("error on %s (%v) err:%s", fname, i, errDur)
		}
	}
	return CreateChapters(titles, durations)
}

type SubtitleEntry struct {
	Dur       time.Duration
	TextLines []string //Max 3 lines 47 chars
}

type SubtitleArr []SubtitleEntry

func (p *SubtitleArr) ToKeyframeString() string {
	keyframes := []string{}
	var d time.Duration
	for _, a := range *p {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		seconds := int(d.Seconds()) % 60
		keyframes = append(keyframes, fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds))
		d += a.Dur
	}
	return strings.Join(keyframes, ",")
}

func durToSrt(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	milliseconds := int(d.Milliseconds()) % 1000
	return fmt.Sprintf("%02d:%02d:%02d,%03d", hours, minutes, seconds, milliseconds)
}

func (p *SubtitleArr) ToSrt() string {
	var sb strings.Builder
	var durNow time.Duration

	for index, entry := range *p {
		sb.WriteString(fmt.Sprintf("%d\n%s --> %s\n",
			index+1,
			durToSrt(durNow), durToSrt(durNow+entry.Dur)))
		sb.WriteString(strings.Join(entry.TextLines, "\n") + "\n\n")
		durNow += entry.Dur

	}
	return sb.String()
}

func (p *SubtitleEntry) CharCount() int {
	result := 0
	for _, line := range p.TextLines {
		result += len(line)
	}
	return result
}

// Split. Time by
func ToSubtitleEntries(totalDur time.Duration, text string) SubtitleArr {
	rows := SplitWithLimit(text, STRMAXLEN)
	result := []SubtitleEntry{SubtitleEntry{TextLines: []string{}}}
	for _, row := range rows {
		row = strings.TrimSpace(row)
		if len(row) == 0 {
			continue
		}

		if len(result[len(result)-1].TextLines) < SRTMAXROWS {
			result[len(result)-1].TextLines = append(result[len(result)-1].TextLines, row)
		} else {
			result = append(result, SubtitleEntry{TextLines: []string{row}})
		}
	}
	totalCount := 0
	for _, e := range result {
		totalCount += e.CharCount()
	}
	for i, e := range result {
		result[i].Dur = (totalDur * time.Duration(e.CharCount())) / time.Duration(totalCount)
	}
	return result
}

// For bug hunting, TODO refactor
func burnFontsArgs(inputVideoFilename string, srtFileName string, outputFilename string, primaryColor color.RGBA) []string {
	red, green, blue, _ := primaryColor.RGBA()

	//TODO burnt font have still too much margin
	return []string{
		"-i", inputVideoFilename,
		"-vf", "subtitles=" + srtFileName + ":force_style='Alignment=9,FontName=Arial,FontSize=10,MarginL=0,MarginR=0,MarginV=0,PrimaryColour=" + fmt.Sprintf("&H%02x%02x%02x&", red>>8, green>>8, blue>>8) + "'",
		"-c:a", "copy", outputFilename}
}

/*
Burning fonts into
ffmpeg -i input.mp4 -vf "subtitles=subtitles.srt:force_style='FontName=Arial,FontSize=24,PrimaryColour=&HFFFFFF&'" -c:a copy output.mp4
*/
func BurnFonts(inputVideoFilename string, srtFileName string, outputFilename string, primaryColor color.RGBA) error {
	return runffmpeg(burnFontsArgs(inputVideoFilename, srtFileName, outputFilename, primaryColor))
}

func SplitWithLimit(text string, limit int) []string {
	words := strings.Split(strings.TrimSpace(text), " ")
	if len(words) == 0 {
		return nil
	}
	result := []string{}
	currentLine := ""

	for _, word := range words {
		if len(currentLine)+len(word) > limit {
			result = append(result, strings.TrimSpace(currentLine))
			currentLine = word + " "
		} else {
			currentLine += word + " "
		}
	}

	if currentLine != "" {
		result = append(result, strings.TrimSpace(currentLine))
	}
	return result
}

func GetTmpFileNames(n int, extension string) ([]string, error) {
	result := make([]string, n)

	for i := range result {
		f, err := os.CreateTemp("/tmp", "*."+extension)
		if err != nil {
			return nil, err
		}
		f.Close()
		result[i] = f.Name()
	}
	return result, nil
}

func RemoveFileList(lst []string) {
	for _, name := range lst {
		os.Remove(name)
	}
}

/*
Whole game video contains parts from each room

1. Description audio
2. Quet part (2s?)
3. prompt part

tmp files
0: quiet part video
1:
*/

func CreateRoomVideo(imageFilenameOriginal string, audioFilename string, audioPromptFilename string, text string, promptText string, outputFilename string) error {
	//tmpFnames,errTmp:=GetTmpFileNames(3)
	tmpPngFiles, errTmpPng := GetTmpFileNames(1, "png")
	if errTmpPng != nil {
		return errTmpPng
	}
	imageFilename := tmpPngFiles[0]

	errCreateBg := createBgImage(imageFilenameOriginal, imageFilename)
	if errCreateBg != nil {
		return errCreateBg
	}

	tmpWavFiles, errTmpWav := GetTmpFileNames(1, "wav")
	if errTmpWav != nil {
		return errTmpWav
	}
	tmpTxtFiles, errTmpTxt := GetTmpFileNames(1, "txt")
	if errTmpTxt != nil {
		return errTmpTxt
	}
	tmpMp4Files, errTmpMp4 := GetTmpFileNames(5, "mp4")
	if errTmpMp4 != nil {
		return errTmpMp4
	}
	tmpSrtFiles, errTmpSrt := GetTmpFileNames(2, "srt")
	if errTmpSrt != nil {
		return errTmpSrt
	}

	RemoveFileList(tmpWavFiles)
	RemoveFileList(tmpTxtFiles)
	RemoveFileList(tmpMp4Files)
	RemoveFileList(tmpSrtFiles)

	tmpSilentWav := tmpWavFiles[0]   //.wav
	tmpSilentVideo := tmpMp4Files[0] //.mp4

	tmpDescription := tmpMp4Files[1]         //.mp4
	tmpDescriptionSrt := tmpSrtFiles[0]      //srt
	tmpDescriptionSubVideo := tmpMp4Files[2] //mp4

	tmpPrompt := tmpMp4Files[3]         //mp4
	tmpPromptSrt := tmpSrtFiles[1]      //srt
	tmpPromptSubVideo := tmpMp4Files[4] //mp4

	tmpJoinList := tmpTxtFiles[0] //txt

	os.Remove(outputFilename)

	//tmpSilentVideo part
	errSilent := createSilentAudio(time.Millisecond*2500, tmpSilentWav)
	if errSilent != nil {
		return fmt.Errorf("errSilent=%s\n", errSilent)
	}
	errSilentVideo := createImageSoundVideo(tmpSilentWav, imageFilename, "00:00:00,00:00:01,00:00:02", tmpSilentVideo)
	if errSilentVideo != nil {
		return fmt.Errorf("errSilentVideo=%s\n", errSilentVideo)
	}

	//Description part
	soundDur, errDur := getMediaLen(audioFilename) //getWavDuration(audioFilename)
	if errDur != nil {
		return fmt.Errorf("error getting %s duration err:%s", audioFilename, errDur)
	}

	subEntries := ToSubtitleEntries(soundDur, text)
	errWrite := os.WriteFile(tmpDescriptionSrt, []byte(subEntries.ToSrt()), 0666)
	if errWrite != nil {
		return fmt.Errorf("error write %s", errWrite)
	}

	errDescVideo := createImageSoundVideo(audioFilename, imageFilename, subEntries.ToKeyframeString(), tmpDescription)
	if errDescVideo != nil {
		return fmt.Errorf("error creating description video:%s", errDescVideo)
	}
	errBurnDescFonts := BurnFonts(tmpDescription, tmpDescriptionSrt, tmpDescriptionSubVideo, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	if errBurnDescFonts != nil {
		return fmt.Errorf("error burining description fonts %s", errBurnDescFonts)
	}

	if 0 < len(promptText) { //Description part
		soundPromptDur, errPromptDur := getMediaLen(audioPromptFilename) //getWavDuration(audioPromptFilename)
		if errPromptDur != nil {
			return fmt.Errorf("error getting %s duration err:%s", audioPromptFilename, errPromptDur)
		}
		subPromptEntries := ToSubtitleEntries(soundPromptDur, promptText)
		errWrite = os.WriteFile(tmpPromptSrt, []byte(subPromptEntries.ToSrt()), 0666)
		if errWrite != nil {
			return fmt.Errorf("error write %s", errWrite)
		}

		//Prompt part
		errPromptVideo := createImageSoundVideo(audioPromptFilename, imageFilename, subPromptEntries.ToKeyframeString(), tmpPrompt)
		if errPromptVideo != nil {
			return fmt.Errorf("error prompt video %s", errPromptVideo)
		}
		errBurnPrompt := BurnFonts(tmpPrompt, tmpPromptSrt, tmpPromptSubVideo, color.RGBA{R: 0, G: 255, B: 0, A: 255})
		if errBurnPrompt != nil {
			return fmt.Errorf("error burning prompt %s", errBurnPrompt)
		}
		errJoin := JoinVideos([]string{
			path.Base(tmpDescriptionSubVideo),
			//path.Base(tmpSilentVideo),
			path.Base(tmpPromptSubVideo)}, //list in same dir so just base name
			tmpJoinList, outputFilename)
		if errJoin != nil {
			return errJoin //Leave tmp files for debug :)
		}
	} else {
		//no user prompt avail
		errJoin := JoinVideos([]string{
			path.Base(tmpDescriptionSubVideo),
			//path.Base(tmpSilentVideo),
		}, //list in same dir so just base name
			tmpJoinList, outputFilename)
		if errJoin != nil {
			return errJoin //Leave tmp files for debug :)
		}

	}

	RemoveFileList(tmpWavFiles)
	RemoveFileList(tmpTxtFiles)
	RemoveFileList(tmpMp4Files)
	RemoveFileList(tmpSrtFiles)
	RemoveFileList(tmpPngFiles)
	return nil
}
