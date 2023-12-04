package llamainterface

import (
	"io"
	"net/http"
	"net/url"
)

type Props struct {
	AssistantName string `json:"assistant_name"`
	UserName      string `json:"user_name"`
}

func (p *LLamaServer) GetProps() ([]byte, error) {

	url, errJoin := url.JoinPath(p.baseUrl, "/props")
	if errJoin != nil {
		return []byte(""), errJoin
	}

	res, err := http.Get(url)
	if err != nil {
		return []byte(""), err
	}
	return io.ReadAll(res.Body)
}
