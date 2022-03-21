package app

import (
	"log"
	"net"
	"strconv"

	"github.com/devansh42/shree2/netio"
)

func (c *Client) CreateLocalTunnel(tunnelPort, actualPort int) error {

	listener, err := net.Listen("tcp", net.JoinHostPort("", strconv.Itoa(tunnelPort)))
	if err != nil {
		log.Print("Couldn't create local tunnel: ", err)
		return err
	}
	go c.manageLocalTunnel(listener, tunnelPort, actualPort)

	return nil
}

func (c *Client) DisconnectLocalTunnel(tunnelPort int) {
	val, ok := c.portMap.LoadAndDelete(tunnelPort)
	if ok {
		closureCh := val.(netio.ClosureCh)
		close(closureCh)
		c.localPortMap.Delete(tunnelPort)
	}
}

func (c *Client) ListLocalTunnel() []netio.PortTuple {
	var ports []netio.PortTuple
	c.localPortMap.Range(func(key, value interface{}) bool {
		tPort := key.(int)
		dPort := value.(int)
		ports = append(ports, netio.PortTuple{
			ClientPort:      dPort,
			RelayServerPort: tPort,
		})
		return false
	})
	return ports
}

func (c *Client) manageLocalTunnel(listener net.Listener, tunnelPort, actualPort int) {
	defer listener.Close()
	tunnelClosed := make(netio.ClosureCh)

	for {
		select {
		case <-tunnelClosed:
			return
		default:
			newConn, err := listener.Accept()
			if err != nil {
				log.Print("Couldn't accept new connection: ", err)
				continue
			}
			newTargetConn, err := net.Dial("tcp", net.JoinHostPort("", strconv.Itoa(actualPort)))
			if err != nil {
				log.Print("Couldn't dial to actual port ", actualPort, " due to: ", err)
				continue
			}
			c.portMap.Store(tunnelPort, tunnelClosed)
			go netio.PipeIO(newConn, newTargetConn, tunnelClosed)

		}
	}
}
