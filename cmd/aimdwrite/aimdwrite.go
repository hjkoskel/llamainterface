package main

import (
	"flag"
	"fmt"
	"llamainterface"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
)

// Get work, and also check that work is still same after run
func GetWorkContextFromFile(fname string) (string, error) {
	inputDoc, errLoad := LoadMdFileFromFile(fname)
	if errLoad != nil {
		return "", fmt.Errorf("error loading %s\n", errLoad.Error())
	}
	cmdDoc := inputDoc.GetFileToCmd()
	if cmdDoc == nil {
		//No work to do
		return "", nil
	}
	return cmdDoc.ToCleanText(), nil
}

func GetPromptFromDoc(doc MdFile, maxTokens int, srv llamainterface.LLamaServer) (string, error) {
	takeSummary := 0
	prompt := ""
	for {
		cmdDoc := doc.GetFileToCmd()
		if cmdDoc == nil {
			//No work to do
			return "", nil
		}
		cmdDoc = cmdDoc.GetSummaryFromN(takeSummary)
		if len(cmdDoc) == 0 {
			return "", fmt.Errorf("ERR, please add summaries, summaries asked=%v\n", takeSummary)
		}

		prompt = cmdDoc.ToCleanText()
		prompt = strings.Replace(prompt, "\n", "\\n", -1) //NO NYT ON VAIHDETTU

		tokenVec, errTokenize := srv.PostTokenize(prompt, time.Second*60)
		if errTokenize != nil {
			return "", fmt.Errorf("err tokenize=%s\n", errTokenize.Error())
		}

		color.Yellow("Tokens used=%v/%v = %.2f percent \n", len(tokenVec), maxTokens, 100*float64(len(tokenVec))/float64(maxTokens))

		if len(tokenVec) <= maxTokens {
			return prompt, nil
		}
		takeSummary++
	}

}

const (
	ERRLLAMASERVERCONNECTION = 1
)

const (
	DIR_TEMP             = "TEMP"
	DIR_N                = "N"
	DIR_TOPK             = "TOPK"
	DIR_TOPP             = "TOPP"
	DIR_NKEEP            = "NKEEP"
	DIR_TFS_Z            = "TFS_Z"
	DIR_TYPICALP         = "TYPICALP"
	DIR_REPEATPENALTY    = "REPEATPENALTY"
	DIR_REPEATLASTN      = "REPEATLASTN"
	DIR_PENALIZENL       = "PENALIZENL"
	DIR_PRESENCEPENALTY  = "PRESENCEPENALTY"
	DIR_FREQUENCYPEBALTY = "FREQUENCYPEBALTY"
	DIR_MIROSTAT         = "MIROSTAT"
	DIR_MIROSTATTAU      = "MIROSTATTAU"
	DIR_MIROSTATETA      = "MIROSTATETA"
	DIR_SEED             = "SEED"
)

// Returns string without directive... directive name and value part. If not found, return same string and no name
func pickDirectiveValue(s string) (string, string, string) {
	dirArr := []string{
		DIR_TEMP,
		DIR_N,
		DIR_TOPK,
		DIR_TOPP,
		DIR_NKEEP,
		DIR_TFS_Z,
		DIR_TYPICALP,
		DIR_REPEATPENALTY,
		DIR_REPEATLASTN,
		DIR_PENALIZENL,
		DIR_PRESENCEPENALTY,
		DIR_FREQUENCYPEBALTY,
		DIR_MIROSTAT,
		DIR_MIROSTATTAU,
		DIR_MIROSTATETA,
		DIR_SEED}
	for _, name := range dirArr {
		iStart := strings.Index(s, "$"+name+"=")
		//fmt.Printf("iStart=%v\n", iStart)
		if 0 <= iStart {
			//fmt.Printf("ALKAVA %s\n", s[iStart:])
			iSpace := strings.Index(s[iStart:], " ")
			iNewline := strings.Index(s[iStart:], "\n")
			iNewlineEscaped := strings.Index(s[iStart:], "\\n")

			if iSpace < 0 {
				iSpace = len(s) - 1 - iStart
			}
			if iNewline < 0 {
				iNewline = len(s) - 1 - iStart
			}
			if iNewlineEscaped < 0 {
				iNewlineEscaped = len(s) - 1 - iStart
			}
			iEnd := min(iSpace, iNewline, iNewlineEscaped) + iStart
			assignment := s[iStart+1 : iEnd] //+1 removes $
			arr := strings.Split(assignment, "=")
			if len(arr) == 2 {
				return strings.Replace(s, "$"+assignment, "", 1), arr[0], arr[1]
			}
		}
	}
	return s, "", ""
}

// Sets correct prompt, remove directives
func PickDirectives(prompt string, initialQueryCompletion llamainterface.QueryCompletion) (llamainterface.QueryCompletion, error) {

	result := initialQueryCompletion
	result.Prompt = prompt

	varname := ""
	varvalue := ""

	for {
		result.Prompt, varname, varvalue = pickDirectiveValue(result.Prompt)
		if varname == "" {
			break
		}
		var parseErr error
		var i int64
		switch varname {
		case DIR_TEMP:
			result.Temperature, parseErr = strconv.ParseFloat(varvalue, 64)
		case DIR_N:
			i, parseErr = strconv.ParseInt(varvalue, 10, 32)
		case DIR_TOPK:
			result.Top_k, parseErr = strconv.ParseFloat(varvalue, 64)
		case DIR_TOPP:
			result.Top_p, parseErr = strconv.ParseFloat(varvalue, 64)
		case DIR_NKEEP:
			i, parseErr = strconv.ParseInt(varvalue, 10, 32)
		case DIR_TFS_Z:
			result.Tfs_z, parseErr = strconv.ParseFloat(varvalue, 64)
		case DIR_TYPICALP:
			result.Typical_p, parseErr = strconv.ParseFloat(varvalue, 64)
		case DIR_REPEATPENALTY:
			result.Repeat_penalty, parseErr = strconv.ParseFloat(varvalue, 64)
		case DIR_REPEATLASTN:
			i, parseErr = strconv.ParseInt(varvalue, 10, 32)
			result.Repeat_last_n = int(i)
		case DIR_PENALIZENL:
			result.Penalize_nl, parseErr = strconv.ParseBool(varvalue)
		case DIR_PRESENCEPENALTY:
			result.Presence_penalty, parseErr = strconv.ParseFloat(varvalue, 64)
		case DIR_FREQUENCYPEBALTY:
			result.Frequency_penalty, parseErr = strconv.ParseFloat(varvalue, 64)
		case DIR_MIROSTAT:
			i, parseErr = strconv.ParseInt(varvalue, 10, 32)
			result.Mirostat = int(i)
		case DIR_MIROSTATTAU:
			result.Mirostat_tau, parseErr = strconv.ParseFloat(varvalue, 64)
		case DIR_MIROSTATETA:
			result.Mirostat_eta, parseErr = strconv.ParseFloat(varvalue, 64)
		case DIR_SEED:
			result.Seed, parseErr = strconv.ParseInt(varvalue, 0, 64)
		default:
			return result, fmt.Errorf("unknown variable name %v", varname)
		}

		if parseErr != nil {
			return result, fmt.Errorf("error parsing %s=%s err=%s", varname, varvalue, parseErr.Error())
		}
	}

	return result, nil
}

func main() {
	pServerHost := flag.String("h", "127.0.0.1", "llama.cpp server host")
	pServerPort := flag.Int("p", 8080, "llama.cpp server port")

	pFileName := flag.String("f", "example.md", "file for reading and writing")
	pMaxInputTokenCount := flag.Int("mt", 1024, "max tokens in input prompt") //Context = inputTokens+outputTokens
	flag.Parse()

	srv, errSrv := llamainterface.InitLLamaServer(*pServerHost, *pServerPort)
	if errSrv != nil {
		color.Red("err %s\n", errSrv.Error())
		os.Exit(-1)
	}

	fwatcher, errFWatch := fsnotify.NewWatcher()
	if errFWatch != nil {
		color.Red("err %s\n", errFWatch.Error())
		os.Exit(-2)
	}
	defer fwatcher.Close()

	errFWatch = fwatcher.Add(*pFileName)
	if errFWatch != nil {
		color.Red("err %s\n", errFWatch.Error())
		os.Exit(-2)
	}

	for {
		inputDoc, errLoad := LoadMdFileFromFile(*pFileName)
		if errLoad != nil {
			fmt.Printf("error loading %s\n", errLoad.Error())
			os.Exit(-2)
		}

		prompt, promptErr := GetPromptFromDoc(inputDoc, *pMaxInputTokenCount, srv)
		if promptErr != nil {
			color.Red(fmt.Sprintf("ERR %v\n", promptErr.Error()))
			return
		}

		if len(prompt) == 0 {
			fmt.Printf("\n\nDONE! Waiting changes on file %s\n", *pFileName)
			select {
			case event, ok := <-fwatcher.Events:
				if !ok {
					fmt.Printf("Watching file ended\n")
					return
				}
				if event.Has(fsnotify.Write) {
					fmt.Printf("modified file:%s\n", event.Name)
				}
			case err, ok := <-fwatcher.Errors:
				if !ok {
					return
				}
				color.Red("file watch error %s\n", err.Error())
				return
			}
			continue
		}

		errLastPrompt := os.WriteFile("lastprompt.txt", []byte(prompt), 0666)
		if errLastPrompt != nil {
			color.Red(fmt.Sprintf("errLastPrompt=%s\n", errLastPrompt))
			return
		}

		completionSetting, errCompletionSetting := PickDirectives(prompt, llamainterface.DefaultQueryCompletion())
		if errCompletionSetting != nil {
			color.Red(fmt.Sprintf("errCompletionSetting=%s\n", errCompletionSetting))
		}
		//completionSetting.Stream = true
		//completionSetting.N_predict = 10

		color.Blue(completionSetting.DiffReport(llamainterface.DefaultQueryCompletion()))

		fmt.Printf("\n\n")
		color.Green(completionSetting.Prompt)
		fmt.Printf("\n\n")

		runResult, errRun := srv.PostCompletion(completionSetting, nil, time.Minute*60)
		if errRun != nil {
			color.Red(fmt.Sprintf("Error running %s\n", errRun))
			return
		}

		//Is this going to be same work. for what answer is generated
		refDoc, errRefLoad := LoadMdFileFromFile(*pFileName)
		if errRefLoad != nil {
			color.Red("error loading %s\n", errRefLoad.Error())
			return
		}
		fRef := refDoc.GetFileToCmd()
		fDoc := inputDoc.GetFileToCmd()
		if fRef.ToCleanText() != fDoc.ToCleanText() {
			color.Magenta("WARNING!!! Document changed on disk while calculating. Re-running....\n")
		} else {
			//WRite to ref doc... if user have writed something after running mark while LLM was running
			errSet := refDoc.SetToCmd(runResult.Content)
			if errSet != nil {
				color.Red("error setting to doc %s\n", errSet.Error())
				return
			}

			finalResult, errResult := refDoc.ToFileContent()
			if errResult != nil {
				color.Red("extracting document to content err=%s\n", errResult)
				return
			}
			writeErr := os.WriteFile(*pFileName, []byte(finalResult), 0666)
			if writeErr != nil {
				color.Red("error writing file %s ,err=%s\n", *pFileName, writeErr.Error())
				return
			}
		}
	}
}
