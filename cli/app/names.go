package app

import (
	"sync"
)

type Client struct {
	certBytes     []byte
	portMap       *sync.Map
	localPortMap  *sync.Map
	remotePortMap *sync.Map
	uid           int
}

func New(certBytes []byte, uid int) *Client {
	return &Client{
		certBytes:     certBytes,
		portMap:       new(sync.Map),
		localPortMap:  new(sync.Map),
		remotePortMap: new(sync.Map),
		uid:           uid,
	}
}
