package relay

import (
	"encoding/gob"
	"log"
	"net"

	"github.com/devansh42/shree2/netio"
)

func (r *Relay) StartApplicationServer() {

	listener, err := net.Listen("tcp", applicationServerPort)
	if err != nil {
		log.Fatal("Couldn't start listener: ", err)
	}
	defer listener.Close()
	log.Print("Started Application Listener at ", applicationServerPort)

	for {
		newConn, err := listener.Accept()
		if err != nil {
			log.Print("Couldn't accept new connections: ", err)
			continue
		}
		r.handleApplicationOps(newConn)
	}
}

func (r *Relay) handleApplicationOps(conn net.Conn) {
	defer conn.Close()
	dec := gob.NewDecoder(conn)
	var req netio.ReqCmd
	err := dec.Decode(&req)
	if err != nil {
		log.Print("Couldn't decode application request payload: ", err)
		return
	}
	switch req.Code {
	case netio.CreateCertificate:
		certificate, err := r.ca.Create()
		if err != nil {
			log.Print("Couldn't create certificate: ", err)
			return
		}
		enc := gob.NewEncoder(conn)
		enc.Encode(&netio.CreateCertificateResp{
			Bytes: certificate.Certificate.Bytes(),
		})
	case netio.AttachTunnel:
		nreq := req.Payload.(netio.AttachTunnelReq)
		r.attachTunnel(nreq.Uid, nreq.PortTuple)
	case netio.DisconnectTunnel:
		nreq := req.Payload.(netio.DisconnectTunnelReq)
		r.disconnectTunnel(nreq.Uid, nreq.Port)
	case netio.ListTunnel:
		nreq := req.Payload.(netio.ListTunnelsReq)
		ports := r.getPorts(nreq.Uid)
		enc := gob.NewEncoder(conn)
		err = enc.Encode(netio.ListTunnelResp{
			Ports: ports,
		})
		if err != nil {
			log.Print("Couldn't encode list ports response payload: ", err)
		}
	}
}
