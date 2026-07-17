package base

import "net/http"

type Common struct {
	HTTP *http.Client
}

func (c *Common) SetDeps(client *http.Client) {
	c.HTTP = client
}

func (c *Common) Init() {

}
