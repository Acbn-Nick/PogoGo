package client

import "context"

type Client struct {
	config *Configuration
	ctx    context.Context
	done   chan interface{}
}

func New(ctx context.Context) (*Client, chan interface{}) {
	c := &Client{done: make(chan interface{})}
	return c, c.done
}

func (c *Client) Start() {

}
