package relay

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/gob"
	"errors"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/devansh42/shree2/netio"
	"github.com/devansh42/shree2/serv/ca"
)

const (
	relatServerListenerPort = ":8080"
	applicationServerPort   = ":8081"
)

type Relay struct {
	rootCertificateBytes []byte
	rootCertificate      *x509.Certificate

	relayServerCetificateBytes []byte
	listener                   net.Listener
	portMap                    *sync.Map
	uidPortMap                 *sync.Map
	connId                     *int32
	ca                         ca.CA
}

func New(rootCertificateBytes,
	rootPrivKeyBytes,
	relayServerCertificateBytes []byte) (*Relay, error) {

	rootCertificate, err := ca.DecodeCertificate(rootCertificateBytes)
	if err != nil {
		return nil, err
	}
	rootPrivKey, err := ca.DecodePrivateKey(rootPrivKeyBytes)
	if err != nil {
		return nil, err
	}
	certAuth, err := ca.New(rootCertificate, rootPrivKey)
	if err != nil {
		return nil, err
	}

	return &Relay{
		rootCertificateBytes:       rootCertificateBytes,
		rootCertificate:            rootCertificate,
		relayServerCetificateBytes: relayServerCertificateBytes,
		portMap:                    new(sync.Map),
		uidPortMap:                 new(sync.Map),
		ca:                         certAuth,
		connId:                     new(int32),
	}, nil

}

func (r Relay) verifyConnection(cs tls.ConnectionState) error {
	if len(cs.PeerCertificates) == 0 {
		return errors.New("Invalid number for tls certificates")
	}
	certs := cs.PeerCertificates
	cert := certs[0]

	return cert.CheckSignatureFrom(r.rootCertificate)
}

func (r *Relay) StartRelayServer() {
	listener, err := tls.Listen("tcp", relatServerListenerPort, &tls.Config{
		VerifyConnection: r.verifyConnection,
		Certificates:     []tls.Certificate{{Certificate: [][]byte{r.rootCertificateBytes}}},
	})
	if err != nil {
		log.Fatal("Couldn't start listener: ", err)
	}
	defer listener.Close()
	log.Print("Started Relay Listener at ", relatServerListenerPort)
	for {
		newConn, err := listener.Accept()
		if err != nil {
			log.Print("Couldn't accept new connections: ", err)
			continue
		}
		ch := make(netio.ClosureCh)
		port, err := r.initNewTunnel(newConn)
		if err != nil {
			newConn.Close()
			continue
		}
		go r.manageNewTunneledConnection(newConn, ch, port)
	}
}

func (r *Relay) getPorts(uid int) []netio.PortTuple {
	var respPorts []netio.PortTuple
	val, ok := r.uidPortMap.Load(uid)
	if ok {
		respPorts = val.([]netio.PortTuple)
	}
	return respPorts
}

func (r *Relay) attachTunnel(uid int, port netio.PortTuple) {

	val, ok := r.uidPortMap.LoadAndDelete(uid)
	if ok {
		ports := val.([]netio.PortTuple)
		ports = append(ports, port)
		r.uidPortMap.Store(uid, ports)
	}
}

func (r *Relay) disconnectTunnel(uid, port int) {
	val, ok := r.portMap.Load(port)
	if ok {
		closeCh := val.(netio.ClosureCh)
		close(closeCh)
	}
	val, ok = r.uidPortMap.LoadAndDelete(uid)
	if ok {
		ports := val.([]netio.PortTuple)
		var leftPorts []netio.PortTuple
		for _, v := range ports {
			if v.RelayServerPort != port {
				leftPorts = append(leftPorts, v)
			}
		}
		if len(leftPorts) > 0 {
			r.uidPortMap.Store(uid, leftPorts)
		}
	}
}

func (r *Relay) initNewTunnel(conn net.Conn) (int, error) {
	addr := conn.LocalAddr().String()

	ss := strings.Split(addr, ":")
	l := len(ss)
	ports := ss[l-1]
	port, err := strconv.Atoi(ports)
	if err != nil {
		log.Print("Port Decoding Error: ", err)
		return 0, err
	}
	var data = netio.CreateTunnelResp{Port: port}
	encoder := gob.NewEncoder(conn)
	err = encoder.Encode(&netio.ReqCmd{
		Payload: data,
		Code:    netio.TunnelCreated,
	})
	if err != nil {
		log.Print("GobEncoding Error: ", err)
		return 0, err
	}

	return port, err
}

func (r *Relay) manageNewTunneledConnection(conn net.Conn, tunnelClosed netio.ClosureCh, port int) {
	listener, err := tls.Listen("tcp", ":0", &tls.Config{
		Certificates: []tls.Certificate{{Certificate: [][]byte{r.relayServerCetificateBytes}}},
	})
	if err != nil {
		log.Print("Couldn't start listener for relay connection: ", err)
	}
	defer listener.Close()
	r.portMap.Store(port, tunnelClosed)
	closeTunnel := make(netio.ClosureCh)
	go r.manageNewIncommingConnectionsOnTunnel(conn, listener, closeTunnel)

	<-tunnelClosed
	close(closeTunnel)
}

func (r *Relay) manageNewIncommingConnectionsOnTunnel(tunneledConn net.Conn,
	srcListener net.Listener,
	tunnelClosed netio.ClosureCh) {
	connMultiplexer := netio.NewInitialzedMultiplexedConn(tunneledConn)

	for {
		select {
		case <-tunnelClosed:
			return
		default:
			newIncomingConn, err := srcListener.Accept()
			if err != nil {
				log.Print("Couldn't start listening for new connection: ", err)
				continue
			}
			nextId := atomic.AddInt32(r.connId, 1)
			if err != nil {
				log.Print("Couldn't initialized multiplexed writer for new connection: ", err)
				continue
			}
			rwc := connMultiplexer.GetReaderWriterCloser(int(nextId))
			go netio.HangeTunnelIO(newIncomingConn, rwc, tunnelClosed)

		}
	}
}
