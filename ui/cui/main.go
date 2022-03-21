package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

const urlPrefix = "http://localhost:5000/"

type (
	Cmd struct {
		tunnel, disconnect, local, remote, list *bool
		port, srcPort                           *int
	}
	respPayload struct {
		Data  interface{} `json:"data"`
		Error error       `json:"err"`
	}
	tunnel struct {
		IsRemote bool `json:"isRemote"`
		Src      int  `json:"src"`
		Tar      int  `json:"tar"`
	}
	listTunnelP struct {
		Tunnels []tunnel `json:"tunnels"`
	}
	listTunnelsResp struct {
		Error error       `json:"err"`
		Data  listTunnelP `json:"data"`
	}
	createRemoteTunnelResp struct {
		Error error               `json:"err"`
		Data  createRemoteTunnelP `json:"data"`
	}
	createRemoteTunnelP struct {
		Src int `json:"src"`
	}
)

func main() {
	var cmd Cmd

	cmd.tunnel = flag.Bool("tunnel", false, "Want to create a tunnel")
	cmd.disconnect = flag.Bool("disconnect", false, "Want to disconnect tunnel")
	cmd.local = flag.Bool("local", false, "Subject is local tunnel")
	cmd.remote = flag.Bool("remote", false, "Subject is remote tunnel")
	cmd.list = flag.Bool("list", false, "Want to create a local tunnel")
	cmd.port = flag.Int("port", 0, "Port to redirect tunnel traffic to")
	cmd.srcPort = flag.Int("srcPort", 0, "Local port to be open in local tunneling")

	flag.Parse()
	for !flag.Parsed() {
	}

	executeCmd(cmd)
}

func executeCmd(cmd Cmd) {
	var (
		err  error
		resp *http.Response
	)
	switch {
	case *cmd.tunnel && *cmd.local && *cmd.srcPort > 0 && *cmd.port > 0 && *cmd.port != *cmd.srcPort:
		resp, err = http.Post(urlPrefix+"tunnel/local", "application/json", getReqPayloadReader(

			map[string]interface{}{
				"src": *cmd.srcPort,
				"tar": *cmd.port,
			},
		))
		if err == nil {
			createLocalTunnel(resp)
		}
	case *cmd.tunnel && *cmd.remote && *cmd.port > 0:
		resp, err = http.Post(urlPrefix+"tunnel/remote", "application/json", getReqPayloadReader(
			map[string]interface{}{
				"tar": *cmd.port,
			},
		))
		if err == nil {
			createRemoteTunnel(resp)
		}
	case *cmd.disconnect && (*cmd.remote || *cmd.local) && *cmd.port > 0:
		req, _ := http.NewRequest(http.MethodDelete, urlPrefix+"tunnel", getReqPayloadReader(
			map[string]interface{}{
				"tar":      *cmd.port,
				"isRemote": *cmd.remote,
			},
		))
		req.Header.Set("content-type", "application/json")
		resp, err = http.DefaultClient.Do(req)

		if err == nil {
			disconnectTunnel(resp, *cmd.remote)
		}
	case *cmd.list:
		resp, err = http.Get(urlPrefix + "tunnels")
		if err == nil {
			listTunnels(resp)
		}
	}
	if err != nil {
		fmt.Print("Couldn't perform action: ", err.Error())
	}

}

func getReqPayloadReader(i interface{}) io.Reader {
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(i)
	return buf
}

func listTunnels(resp *http.Response) {
	var rep listTunnelsResp
	json.NewDecoder(resp.Body).Decode(&rep)

	if resp.StatusCode == http.StatusOK {
		fmt.Print("\nTunnels!!")
		tunnels := rep.Data.Tunnels
		if len(tunnels) > 0 {
			fmt.Print("\n\tType\tPort\tSecondary Port")
			for _, v := range tunnels {
				var t string = "Local"
				var src, tar string
				if v.IsRemote {
					t = "Remote"
					src = strconv.Itoa(v.Src)
				}
				tar = strconv.Itoa(v.Tar)
				fmt.Printf("\n\t%s\t%s\t%s", t, tar, src)
			}
		} else {
			fmt.Print("\n 0 Tunnel(s)")
		}
	} else {
		fmt.Print("\nError occured while listing tunnel(s): ", rep.Error.Error())
	}
}
func createLocalTunnel(resp *http.Response) {
	if resp.StatusCode == http.StatusOK {
		fmt.Print("\nLocal Tunnel Created!!")
	} else {
		var rep respPayload
		json.NewDecoder(resp.Body).Decode(&rep)
		fmt.Print("\nError occured while creating local tunnel: ", rep.Error.Error())
	}
}

func createRemoteTunnel(resp *http.Response) {
	var rep createRemoteTunnelResp
	json.NewDecoder(resp.Body).Decode(&rep)

	if resp.StatusCode == http.StatusOK {
		fmt.Print("\nRemote Tunnel Created!!")
		fmt.Print("\n Assigned Port: ", rep.Data.Src)
	} else {
		fmt.Print("\nError occured while creating remote tunnel: ", rep.Error.Error())
	}

}
func disconnectTunnel(resp *http.Response, isRemote bool) {
	var t = "Local"
	if isRemote {
		t = "Remote"
	}
	if resp.StatusCode == http.StatusOK {

		fmt.Printf("\n%s Tunnel Disconnected!!", t)
	} else {
		var rep respPayload
		json.NewDecoder(resp.Body).Decode(&rep)
		fmt.Printf("\nError occured while disconnected %s tunnel: %s", t, rep.Error.Error())
	}
}
