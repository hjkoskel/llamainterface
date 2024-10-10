/*
See examples
*/
package llamainterface

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

type LLamaServer struct {
	baseUrl        string
	llamaCrashlErr error
	stdoutBuf      bytes.Buffer
	stderrBuf      bytes.Buffer
}

func GetFreePort() (port int, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}
	return
}

// TODO cmd rakennus komennosta ja sitten execute erikseen. voi ajaa llama.cpp serveriä ja llamafileä
func InitLlamafileServerCmd(ctx context.Context, llamafilename string, parameters ServerCommandLineFlags) (*exec.Cmd, error) {
	paramList, errParams := parameters.GetArgs()
	if errParams != nil {
		return nil, errParams
	}
	paramList = append(paramList, "--nobrowser")
	//file says that .llamafile is DOS/MBR boot sector; partition 1 : ID=0x7f, active, start-CHS (0x0,0,1), end-CHS (0x3ff,255,63), startsector 0, 4294967295 sectors
	//HACK that prevents exec format error
	paramList = append([]string{llamafilename}, paramList...)

	//TODO Cgo solution?
	cmd := exec.CommandContext(ctx, "/bin/sh", paramList...)
	cmd.Env = os.Environ()
	cmd.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGTERM}
	return cmd, nil
}

// Start llamafile https://github.com/Mozilla-Ocho/llamafile
func InitLlamafileServer(ctx context.Context, llamafilename string, parameters ServerCommandLineFlags) (LLamaServer, error) {

	//This have special feature, get free TCP port if port is set to 0.
	if parameters.ListenPort == 0 {
		portNumber, errPortNumber := GetFreePort()
		if errPortNumber != nil {
			return LLamaServer{}, fmt.Errorf("error getting TCP port number for llamafile %s", errPortNumber)
		}
		parameters.ListenPort = portNumber
	}

	paramList, errParams := parameters.GetArgs()
	if errParams != nil {
		return LLamaServer{}, errParams
	}
	paramList = append(paramList, "--nobrowser")
	//file says that .llamafile is DOS/MBR boot sector; partition 1 : ID=0x7f, active, start-CHS (0x0,0,1), end-CHS (0x3ff,255,63), startsector 0, 4294967295 sectors
	//HACK that prevents exec format error
	paramList = append([]string{llamafilename}, paramList...)

	//TODO Cgo solution?
	cmd := exec.CommandContext(ctx, "/bin/sh", paramList...)
	cmd.Env = os.Environ()
	cmd.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGTERM}

	result := LLamaServer{baseUrl: fmt.Sprintf("http://127.0.0.1:%v", parameters.ListenPort)}

	//llamaServerCommand.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	//llamaServerCommand.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)
	cmd.Stdout = os.Stdout // or any other io.Writer
	cmd.Stderr = os.Stderr // or any other io.Writer

	go func() {
		runtimeErr := cmd.Run()
		if runtimeErr != nil {
			fmt.Printf("LLAMA runtime err %s\n", runtimeErr)
			os.Exit(-1)
		}

		out, errOut := cmd.CombinedOutput()
		if errOut != nil {
			fmt.Printf("err %s\n", errOut)
			return
		}
		fmt.Printf("LLAMA: %s\n", out)
	}()

	//Wait this line
	//"llama server listening at http://127.0.0.1:37915"

	return result, nil
}

func InitLLamaServer(hostname string, port int) (LLamaServer, error) {
	//TODO error check? Other options. Spin up server if local?
	return LLamaServer{baseUrl: fmt.Sprintf("http://%s:%v", hostname, port)}, nil
}

func (p *LLamaServer) PostQuery(functionality string, payload []byte, feed chan string, timeout time.Duration) (string, error) {
	if len(p.baseUrl) == 0 { //For easier debug if mis using library
		return "", fmt.Errorf("error llamaserver does not have baseUrl defined")
	}

	postUrl, errJoin := url.JoinPath(p.baseUrl, functionality)
	if errJoin != nil {
		return "", fmt.Errorf("error formatting url base=%s functionality=%v", p.baseUrl, functionality)
	}
	request, errRequesting := http.NewRequest("POST", postUrl, bytes.NewBuffer(payload))
	if errRequesting != nil {
		return "", errRequesting
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	client.Timeout = timeout
	response, errDo := client.Do(request)
	if errDo != nil {
		return "", fmt.Errorf("error while doing request %s", errDo.Error())
	}
	defer response.Body.Close()

	fmt.Println("response Status:", response.Status)
	fmt.Println("response Headers:", response.Header)

	if response.StatusCode != 200 {
		b, _ := io.ReadAll(response.Body)

		return "", fmt.Errorf("failed with HTTP code %v err=%s payload=%s", response.StatusCode, response.Status, b)
	}

	reader := bufio.NewReader(response.Body)

	var resultString strings.Builder
	for {
		line, errRead := reader.ReadBytes('\n')
		//fmt.Printf("line=%s\n", line)
		if 0 < cap(feed) {
			feed <- string(line)
		}
		resultString.Write(line)
		resultString.WriteString("\n")
		if errRead == io.EOF {
			if feed != nil {
				close(feed)
			}
			return resultString.String(), nil
		}
		if errRead != nil {
			return resultString.String(), errRead
		}

	}
}
