/*
Two AI agents having debate on common topic
*/

package main

import (
	"flag"
	"fmt"
	"llamainterface"
	"os"
	"path"
	"strings"
	"time"
)

func ReadFileToOneLine(fname string) (string, error) {
	raw, errRead := os.ReadFile(fname)
	if errRead != nil {
		return "", errRead
	}
	//TODO replace all strange chars with newline
	rows := strings.Split(string(raw), "\n")
	for i := range rows {
		rows[i] = strings.TrimSpace(strings.TrimSpace(rows[i]))
	}
	return strings.TrimSpace(strings.Join(rows, " ")), nil
}

func ReadRowsSkipEmpty(fname string) ([]string, error) {
	raw, errRead := os.ReadFile(fname)
	if errRead != nil {
		return nil, errRead
	}
	//TODO replace all strange chars with newline
	rows := strings.Split(string(raw), "\n")

	result := []string{}
	for i := range rows {
		result = append(result, strings.TrimSpace(rows[i]))
	}
	return result, nil
}

func NameFromFilename(fname string) string {
	return strings.Split(path.Base(fname), ".")[0]
}

func main() {
	pServerHost := flag.String("h", "127.0.0.1", "llama.cpp server host")
	pServerPort := flag.Int("p", 8080, "llama.cpp server port")

	pInitialPromptFileA := flag.String("fia", "./persons/winloser.txt", "text file for initial prompt for first person")
	pInitialPromptFileB := flag.String("fib", "./persons/archuser.txt", "text file for initial prompt for second person")

	pNameOfA := flag.String("na", "", "name of first person, leave empty and name comes from prompt filename")
	pNameOfB := flag.String("nb", "", "name of second person, leave empty and name comes from prompt filename")

	pStartChatFile := flag.String("f", "./chats/linuxuse.txt", "initial chat. One turn per row. first row is first person, line by line taking turns")

	pContextLimit := flag.Int("c", 400, "context token count limit when posting. Must be less than context size")

	//Initial part prefix and suffix
	pInitialPartPrefix := flag.String("ipp", "<|im_start|>system ", "initial part prefix (system prompt)")
	pInitialPartSuffix := flag.String("ips", "<|im_end|>", "initial part suffix")

	pAgentPartPrefix := flag.String("app", "<|im_start|>assistant ", "agent part prefix")
	pAgentPartSuffix := flag.String("aps", "<|im_end|>", "agent part suffix")

	pUserPartPrefix := flag.String("upp", "<|im_start|>user ", " user part prefix")
	pUserPartSuffix := flag.String("ups", "<|im_end|>", "user part suffix")

	pLastPromptFileName := flag.String("lp", "lastprompt.txt", "last prompt file")

	//Run settings
	pTempA := flag.Float64("tempA", 0.7, "temperature of first person")
	pTempB := flag.Float64("tempB", 0.7, "temperature of first person")

	//Translation server. Convert back and forth. For example text is finnish and model in english. Then translation is needed
	//todo github
	/*
		pTransServerHost := flag.String("th", "127.0.0.1", "translator server host used if needed")
		pTransServerPort := flag.Int("tp", 8000, "translator server port used if needed")
		pTransSourceLang := flag.String("tsl", "", "source parameter on translator server. Empty no translation")
		pTransTargetLang := flag.String("ttl", "", "target parameter on translator server. Emptu no translation")
	*/

	pOutputLogFile := flag.String("of", "text.log", "output filename")

	flag.Parse()

	srv, errSrv := llamainterface.InitLLamaServer(*pServerHost, *pServerPort)
	if errSrv != nil {
		fmt.Printf("err %s\n", errSrv)
		return
	}

	prt := Arena{

		InitialPart: PrefixSuffix{Prefix: *pInitialPartPrefix, Suffix: *pInitialPartSuffix},
		AgentPart:   PrefixSuffix{Prefix: *pAgentPartPrefix, Suffix: *pAgentPartSuffix},
		UserPart:    PrefixSuffix{Prefix: *pUserPartPrefix, Suffix: *pUserPartSuffix},
	}

	var errRead error
	prt.MyConf.InitialPrompt, errRead = ReadFileToOneLine(*pInitialPromptFileA)
	if errRead != nil {
		fmt.Printf("InitialPrompt A READ ERR %s\n", errRead)
		return
	}
	if len(*pNameOfA) == 0 {
		prt.MyConf.Name = NameFromFilename(*pInitialPromptFileA)
	} else {
		prt.MyConf.Name = *pNameOfA
	}

	prt.AnotherConf.InitialPrompt, errRead = ReadFileToOneLine(*pInitialPromptFileB)
	if errRead != nil {
		fmt.Printf("InitialPrompt A READ ERR %s\n", errRead)
		return
	}
	if len(*pNameOfB) == 0 {
		prt.AnotherConf.Name = NameFromFilename(*pInitialPromptFileB)
	} else {
		prt.AnotherConf.Name = *pNameOfB
	}

	prt.Turns, errRead = ReadRowsSkipEmpty(*pStartChatFile)

	cmdComplA := llamainterface.DefaultQueryCompletion()
	cmdComplB := llamainterface.DefaultQueryCompletion()

	cmdComplA.Temperature = *pTempA
	cmdComplB.Temperature = *pTempB

	cmdComplA.Stop = []string{"<|im_end|>", "<|im_end|><|im_start|>"}
	cmdComplB.Stop = []string{"<|im_end|>", "<|im_end|><|im_start|>"}

	promptCounter := 0
	for !prt.Repeating() {

		prompt := prt.Prompt()
		tokenVec, errTokenize := srv.PostTokenize(prompt, time.Second*60)
		if errTokenize != nil {
			fmt.Printf("err tokenize=%s\n", errTokenize.Error())
			return
		}

		for *pContextLimit <= len(tokenVec) {
			fmt.Printf("Too much tokens %v, going to reduce\n", len(tokenVec))
			prt.StartContextTurnIndex++
			prompt = prt.Prompt()
			tokenVec, errTokenize = srv.PostTokenize(prompt, time.Second*60)
			if errTokenize != nil {
				fmt.Printf("err tokenize=%s\n", errTokenize.Error())
				return
			}

		}

		prompt = strings.Replace(prompt, "\n", "\\n", -1) //escaping

		if 0 < len(*pLastPromptFileName) {
			errLastPrompt := os.WriteFile(*pLastPromptFileName, []byte(prompt), 0666)
			if errLastPrompt != nil {
				fmt.Printf("errLastPrompt=%s\n", errLastPrompt)
				return
			}
		}

		if prt.MyTurnNext() { //Win loser wants to write.. so linux user asks
			cmdComplA.Prompt = prompt
			fmt.Printf("---PROMPT %d winLoser answers this ---\n%s\n\n", promptCounter, prompt)

			resultA, errA := srv.PostCompletion(cmdComplA, nil, time.Minute*20)
			if errA != nil {
				fmt.Printf("A fail %s\n", errA)
				return
			}
			fmt.Printf("\n\n******** Result is ******: %#v\n", resultA.Content)
			s := strings.TrimSpace(resultA.Content)
			if len(s) == 0 {
				break
			}
			appendErr := prt.Append(s, true)
			if appendErr != nil {
				fmt.Printf("appending fail %s\n", appendErr)
				return
			}
		} else {
			cmdComplB.Prompt = prompt
			fmt.Printf("---PROMPT %d linux user answers this ---\n%s\n\n", promptCounter, cmdComplB.Prompt)

			resultB, errB := srv.PostCompletion(cmdComplB, nil, time.Minute*20)
			if errB != nil {
				fmt.Printf("B fail %s\n", errB)
				return
			}
			fmt.Printf("\n\n******** Result is ******: %#v\n", resultB.Content)
			s := strings.TrimSpace(resultB.Content)
			if len(s) == 0 {
				break
			}
			appendErr := prt.Append(s, false)
			if appendErr != nil {
				fmt.Printf("appending fail %s\n", appendErr)
				return
			}
		}

		errLogWrite := os.WriteFile(*pOutputLogFile, []byte(prt.FullUserPrintout()), 0666)
		if errLogWrite != nil {
			fmt.Printf("log writing error %s\n", errLogWrite)
			return
		}
		promptCounter++
	}

	fmt.Printf("\n\nAnd whole thing repeats")
}
