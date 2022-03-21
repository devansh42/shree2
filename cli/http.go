package main

import (
	"encoding/json"
	"net/http"

	"github.com/devansh42/shree2/cli/app"
)

type (
	httpServer struct {
		cli *app.Client
	}
	respPayload struct {
		Data  interface{} `json:"data"`
		Error error       `json:"err"`
	}
	createLocalTunnelReq struct {
		Src int `json:"src"`
		Tar int `json:"tar"`
	}
	createRemoteTunnelReq struct {
		Tar int `json:"tar"`
	}
	createRemoteTunnelResp struct {
		Src int `json:"src"`
	}
	disconnectTunnel struct {
		IsRemote bool `json:"isRemote"`
		Port     int  `json:"port"`
	}
	tunnel struct {
		IsRemote bool `json:"isRemote"`
		Src      int  `json:"src"`
		Tar      int  `json:"tar"`
	}
	listTunnels struct {
		Tunnels []tunnel `json:"tunnels"`
	}
)

func decode(r *http.Request, i interface{}) error {
	return json.NewDecoder(r.Body).Decode(i)
}
func encode(w http.ResponseWriter, i interface{}, err error) {
	wr := json.NewEncoder(w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		wr.Encode(&respPayload{Error: err})
		return
	}
	w.WriteHeader(http.StatusOK)
	wr.Encode(&respPayload{Data: i})

}

func (h httpServer) CreateLocalTunnel(w http.ResponseWriter, r *http.Request) {
	var cr createLocalTunnelReq
	err := decode(r, &cr)
	if err != nil {
		encode(w, nil, err)
		return
	}
	err = h.cli.CreateLocalTunnel(cr.Src, cr.Tar)
	encode(w, nil, err)
}

func (h httpServer) CreateRemoteTunnel(w http.ResponseWriter, r *http.Request) {
	var cr createRemoteTunnelReq
	err := decode(r, &cr)
	if err != nil {
		encode(w, nil, err)
		return
	}
	tport, err := h.cli.CreateRemoteTunnel(cr.Tar)
	encode(w, createRemoteTunnelResp{Src: tport}, err)
}

func (h httpServer) ListTunnels(w http.ResponseWriter, r *http.Request) {
	localports := h.cli.ListLocalTunnel()
	remotePorts := h.cli.ListRemoteTunnel()
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
func (h httpServer) DisconnectTunnel(w http.ResponseWriter, r *http.Request) {
	var dt disconnectTunnel
	err := decode(r, &dt)
	if err != nil {
		encode(w, nil, err)
		return
	}
	if dt.IsRemote {
		err = h.cli.DisconnectRemoteTunnel(dt.Port)
	} else {
		h.cli.DisconnectLocalTunnel(dt.Port)
	}
	encode(w, nil, err)
}
