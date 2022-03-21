package netio

import (
	"io"
	"log"
	"net"
)

type ClosureCh chan struct{}

func HangeTunnelIO(newConn, dest io.ReadWriteCloser, tunnelClosed ClosureCh) {
	defer newConn.Close()
	exitCopying := make(ClosureCh)

	go copyIO(newConn, dest, exitCopying)
	go copyIO(dest, newConn, exitCopying)

	select {
	case <-tunnelClosed:
		return
	case <-exitCopying:
		return
	}
}

func PipeIO(a, b net.Conn, closePipe ClosureCh) {
	defer a.Close()
	defer b.Close()

	ra, wa := io.Pipe()
	rb, wb := io.Pipe()
	exitCopying := make(ClosureCh)
	go copyIO(wa, a, exitCopying)
	go copyIO(b, ra, exitCopying)
	go copyIO(wb, b, exitCopying)
	go copyIO(a, rb, exitCopying)

	select {
	case <-closePipe:
		return
	case <-exitCopying:
		return
	}
}

func copyIO(wr io.Writer, red io.Reader, exitCh ClosureCh) {

	_, err := io.Copy(wr, red)
	if err != nil {
		log.Print("Failed copying io for connection: ", err)
	}
	defer close(exitCh)
}
