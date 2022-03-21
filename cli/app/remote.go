package app

import (
	"crypto/tls"
	"encoding/gob"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/devansh42/shree2/netio"
)

const (
	remoteAppServerAddr = "localhost:9090"
	remoteRelayServer   = "localhost:9091"
)

func (c *Client) CreateRemoteTunnel(actualPort int) (int, error) {
	err := c.manageCertificate()
	if err != nil {
		return 0, err
	}
	tlsConn, err := tls.Dial("tcp", remoteRelayServer, &tls.Config{
		Certificates: []tls.Certificate{{
			Certificate: [][]byte{c.certBytes},
		}},
	})
	if err != nil {
		log.Print("Couldn't create remote tunnel: ", err)
		return 0, err
	}

	tunnelPort, err := c.getTunnelPort(tlsConn)
	if err != nil {
		return 0, err
	}
	err = c.attachTunnel(tunnelPort, actualPort)
	if err != nil {
		return 0, err
	}
	ch := make(netio.ClosureCh)
	c.remotePortMap.Store(tunnelPort, ch)
	go c.manageRemoteTunnel(tlsConn, actualPort, ch)
	return tunnelPort, nil
}

func (c *Client) manageRemoteTunnel(tlsConn net.Conn, actualPort int, closeTunnel netio.ClosureCh) {
	connMultiplexed := netio.NewInitialzedMultiplexedConn(tlsConn)
	ch := connMultiplexed.NewConn()
	defer tlsConn.Close()
	for id := range ch {
		select {
		case <-ch:
			newTunneledconn := connMultiplexed.GetReaderWriterCloser(id)
			newconn, err := net.Dial("tcp", net.JoinHostPort("", strconv.Itoa(actualPort)))
			if err != nil {
				log.Print("Couldn't create a connection to target port: ", err)
			}
			go netio.HangeTunnelIO(newconn, newTunneledconn, closeTunnel)
		case <-closeTunnel:
			return
		}

	}
}

func (c *Client) getTunnelPort(conn net.Conn) (int, error) {
	var err error
	dec := gob.NewDecoder(conn)
	var createTunnel netio.CreateTunnelResp
	err = dec.Decode(&createTunnel)
	if err != nil {
		log.Print("Couldn't decode create tunnel message: ", err)
		return 0, err
	}
	return createTunnel.Port, nil
}

func (c *Client) attachTunnel(tunnelPort, actualPort int) error {
	appConn, err := net.Dial("tcp", remoteAppServerAddr)
	if err != nil {
		log.Print("Couldn't make connection to remote tunnel: ", err)
		return nil
	}
	defer appConn.Close()
	enc := gob.NewEncoder(appConn)
	err = enc.Encode(&netio.ReqCmd{
		Code: netio.AttachTunnel,
		Payload: netio.AttachTunnelReq{
			Uid: c.uid,
			PortTuple: netio.PortTuple{
				RelayServerPort: tunnelPort,
				ClientPort:      actualPort,
			},
		},
	})
	if err != nil {
		log.Print("Couldn't encode payload for attach remote tunnel: ", err)
	}
	return err
}

func (c *Client) manageCertificate() error {
	if len(c.certBytes) > 0 {
		return nil
	}
	conn, err := net.Dial("tcp", remoteAppServerAddr)
	if err != nil {
		log.Print("Couldn't make connection to remote tunnel: ", err)
		return nil
	}
	defer conn.Close()
	enc := gob.NewEncoder(conn)
	err = enc.Encode(&netio.ReqCmd{
		Code:    netio.CreateCertificate,
		Payload: netio.CreateCertificateReq{},
	})
	if err != nil {
		log.Print("Couldn't encode payload for creating certificates: ", err)
	}
	dec := gob.NewDecoder(conn)
	var certResp netio.CreateCertificateResp
	err = dec.Decode(&certResp)
	if err != nil {
		log.Print("Couldn't decode response payload for creating certificates: ", err)
		return nil
	}
	if len(certResp.Bytes) == 0 {
		log.Print("Couldn't get the certificate for some reason")
	}
	c.certBytes = certResp.Bytes
	c.saveCertificate()
	return nil
}

func (c *Client) saveCertificate() {
	h, err := os.UserHomeDir()
	if err != nil {
		log.Print("Couldn't find the user home directory: ", err)
		return
	}
	filePath := strings.Join([]string{h, ".shree.cert.pem"}, "/")
	err = ioutil.WriteFile(filePath, c.certBytes, os.FileMode(os.O_CREATE|os.O_WRONLY))
	if err != nil {
		log.Print("Couldn't write certificate to local storage: ", err)
	}
	return
}

func (c *Client) DisconnectRemoteTunnel(tunnelPort int) error {
	conn, err := net.Dial("tcp", remoteAppServerAddr)
	if err != nil {
		log.Print("Couldn't make connection to remote tunnel: ", err)
		return nil
	}
	defer conn.Close()
	enc := gob.NewEncoder(conn)
	err = enc.Encode(&netio.ReqCmd{
		Code: netio.DisconnectTunnel,
		Payload: netio.DisconnectTunnelReq{
			Port: tunnelPort,
			Uid:  c.uid,
		},
	})
	if err != nil {
		log.Print("Couldn't encode payload for disconnect tunnels: ", err)
		return err
	}
	val, ok := c.remotePortMap.LoadAndDelete(tunnelPort)
	if ok {
		close(val.(netio.ClosureCh))
	}
	return err
}

func (c *Client) ListRemoteTunnel() []netio.PortTuple {
	conn, err := net.Dial("tcp", remoteAppServerAddr)
	if err != nil {
		log.Print("Couldn't make connection to remote tunnel: ", err)
		return nil
	}
	defer conn.Close()
	enc := gob.NewEncoder(conn)
	err = enc.Encode(&netio.ReqCmd{
		Code: netio.ListTunnel,
		Payload: netio.ListTunnelsReq{
			Uid: c.uid,
		},
	})
	if err != nil {
		log.Print("Couldn't encode payload for listing remote tunnels: ", err)
		return nil
	}
	dec := gob.NewDecoder(conn)
	var list netio.ListTunnelResp
	err = dec.Decode(&list)
	if err != nil {
		log.Print("Couldn't decode response payload for listing remote tunnels: ", err)
		return nil
	}
	return list.Ports
}
