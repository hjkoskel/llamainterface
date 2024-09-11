/*
Get llama server health status and some utilities for interpreting status
*/
package llamainterface

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

/*
HTTP status code 503
Body: {"error": {"code": 503, "message": "Loading model", "type": "unavailable_error"}}
Explanation: the model is still being loaded.
HTTP status code 200
Body: {"status": "ok" }
Explanation: the model is successfully loaded and the server is ready.
*/

type Health200 struct {
	Status string `json:"status"`
}

type Health503Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Msgtype string `json:"type"`
}

type Health503 struct {
	Error Health503Error `json:"error"`
}

func (a Health503) String() string {
	return fmt.Sprintf("%s [%s]", a.Error.Message, a.Error.Msgtype)
}

func (p *LLamaServer) GetHealth() (string, error) {
	getUrl, errJoin := url.JoinPath(p.baseUrl, "health")
	if errJoin != nil {
		return "", fmt.Errorf("error formatting url base=%s functionality=%v", p.baseUrl, "health")
	}
	request, errRequesting := http.NewRequest("GET", getUrl, nil)
	if errRequesting != nil {
		return "", errRequesting
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	client.Timeout = time.Second * 1 //HARD CODED?
	response, errDo := client.Do(request)
	if errDo != nil {
		return "", fmt.Errorf("error while doing request %s", errDo.Error())
	}
	defer response.Body.Close()

	fmt.Println("response Status:", response.Status)
	fmt.Println("response Headers:", response.Header)
	raw, errBody := io.ReadAll(response.Body)
	if errBody != nil {
		return "", fmt.Errorf("error reading response body %s", errBody)
	}
	switch response.StatusCode {
	case 200:
		var msg Health200
		errParse := json.Unmarshal(raw, &msg)
		if errParse != nil {
			return "", fmt.Errorf("error parsing %s err:%s", raw, errParse)
		}
		return msg.Status, nil
	case 503:
		var msg Health503
		errParse := json.Unmarshal(raw, &msg)
		if errParse != nil {
			return "", fmt.Errorf("error parsing %s err:%s", raw, errParse)
		}
		return msg.String(), nil
	}
	return "", fmt.Errorf("unknown return code %s  msg:%s", response.Status, raw)

}
