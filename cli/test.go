package main

import (
	"math/rand"
	"net/http"

	"github.com/devansh42/shree2/netio"
)

type testhttpServer struct {
}

func (h testhttpServer) CreateLocalTunnel(w http.ResponseWriter, r *http.Request) {
	var cr createLocalTunnelReq
	err := decode(r, &cr)
	if err != nil {
		encode(w, nil, err)
		return
	}
	encode(w, nil, nil)
}

func (h testhttpServer) CreateRemoteTunnel(w http.ResponseWriter, r *http.Request) {
	var cr createRemoteTunnelReq
	err := decode(r, &cr)
	if err != nil {
		encode(w, nil, err)
		return
	}
	x := 30000 + rand.Intn(5000)
	encode(w, createRemoteTunnelResp{Src: x}, err)
}

func (h testhttpServer) ListTunnels(w http.ResponseWriter, r *http.Request) {
	var remotePorts, localports []netio.PortTuple

	for i := 0; i < 5; i++ {
		remotePorts = append(remotePorts, netio.PortTuple{
			ClientPort:      6000 + rand.Intn(1000),
			RelayServerPort: 30000 + rand.Intn(1000),
		})
		localports = append(localports, netio.PortTuple{
			ClientPort:      7000 + rand.Intn(1000),
			RelayServerPort: 9000 + rand.Intn(1000),
		})
	}
	var ts []tunnel
	for _, v := range localports {
		ts = append(ts, tunnel{
			Src: v.RelayServerPort,
			Tar: v.ClientPort,
		})
	}
	for _, v := range remotePorts {
		ts = append(ts, tunnel{
			IsRemote: true,
			Src:      v.RelayServerPort,
			Tar:      v.ClientPort,
		})
	}
	encode(w, listTunnels{ts}, nil)
}
func (h testhttpServer) DisconnectTunnel(w http.ResponseWriter, r *http.Request) {
	var dt disconnectTunnel
	err := decode(r, &dt)
	if err != nil {
		encode(w, nil, err)
		return
	}
	encode(w, nil, nil)
}
