package clients

import "net/http"

type Default struct {
	client http.Client
}

func NewDefaulut() *Default {
	return &Default{client: http.Client{
		Transport: http.DefaultTransport,
	}}
}

func (http *Default) Request(r *http.Request) (*http.Response, error) {
	return http.client.Do(r)
}
