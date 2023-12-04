/*
See examples
*/
package llamainterface

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type LLamaServer struct {
	baseUrl string
}

func InitLLamaServer(hostname string, port int) (LLamaServer, error) {
	//TODO error check? Other options. Spin up server if local?
	return LLamaServer{baseUrl: fmt.Sprintf("http://%s:%v", hostname, port)}, nil
}

func (p *LLamaServer) PostQuery(functionality string, payload []byte, feed chan string, timeout time.Duration) (string, error) {
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
