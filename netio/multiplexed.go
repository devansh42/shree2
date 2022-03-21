package netio

import (
	"bytes"
	"encoding/binary"
	"io"
	"io/ioutil"
	"log"
	"sync"
)

type mwr struct {
	id         int
	wr         io.WriteCloser
	red        io.Reader
	wbuf, rbuf *bytes.Buffer
	newData    ClosureCh
}

type ConnMultiplexer struct {
	mul               io.ReadWriteCloser
	readerMap, bufMap *sync.Map
	newConn           chan int
}
type mreader struct {
	id  int
	red io.ReadCloser
	buf *bytes.Buffer
}

const (
	closeConnByte    byte = 0
	initConnByte     byte = 1
	continueConnByte byte = 2
)

func (w *mwr) Write(b []byte) (int, error) {
	defer w.wbuf.Reset()

	binary.Write(w.wbuf, binary.BigEndian, w.id)
	binary.Write(w.wbuf, binary.BigEndian, continueConnByte)
	binary.Write(w.wbuf, binary.BigEndian, len(b))
	binary.Write(w.wbuf, binary.BigEndian, b)
	n, err := io.Copy(w.wr, w.wbuf)
	return int(n), err
}
func (w *mwr) init() {
	defer w.wbuf.Reset()
	binary.Write(w.wbuf, binary.BigEndian, w.id)
	binary.Write(w.wbuf, binary.BigEndian, initConnByte)
	_, err := io.Copy(w.wr, w.wbuf)
	if err != nil {
		log.Print("Couldn't send initial message: ", err)
	}
}
func (w *mwr) Close() error {
	defer w.wbuf.Reset()
	defer w.wr.Close()
	defer close(w.newData)

	binary.Write(w.wbuf, binary.BigEndian, w.id)
	binary.Write(w.wbuf, binary.BigEndian, closeConnByte)
	_, err := io.Copy(w.wr, w.wbuf)
	if err != nil {
		return err
	}
	return nil
}

func (w *mwr) Read(b []byte) (int, error) {
	<-w.newData
	return w.rbuf.Read(b)
}

func (w *ConnMultiplexer) startReading() {
	var err error
	for {
		var id, length int
		err = binary.Read(w.mul, binary.BigEndian, &id)
		if err != nil {
			log.Print("Couldn't read id from connection: ", err)
			return
		}
		var flagByte byte
		err = binary.Read(w.mul, binary.BigEndian, &flagByte)
		if err != nil {
			log.Print("Couldn't read flag from connection: ", err)
			return
		}
		switch flagByte {

		case initConnByte:
			w.newConn <- id
		case closeConnByte:
			w.bufMap.Delete(id)
			val, ok := w.readerMap.LoadAndDelete(id)
			if ok {
				close(val.(ClosureCh))
			}
		case continueConnByte:
			err = binary.Read(w.mul, binary.BigEndian, &length)
			if err != nil {
				log.Print("Couldn't read content length: ", err)
				return
			}
			val, ok := w.bufMap.Load(id)
			limitedReader := io.LimitReader(w.mul, int64(length))
			if ok {
				_, err = io.Copy(val.(*bytes.Buffer), limitedReader)
				rd, okk := w.readerMap.Load(id)
				if okk {
					rd.(ClosureCh) <- struct{}{}
				}
			} else {
				_, err = io.Copy(ioutil.Discard, limitedReader)
			}
			if err != nil {
				log.Print("Couldn't read from tunnel connection: ", err)
			}
		}

	}
}

func NewInitialzedMultiplexedConn(rw io.ReadWriteCloser) *ConnMultiplexer {
	w := &ConnMultiplexer{
		mul:       rw,
		bufMap:    new(sync.Map),
		readerMap: new(sync.Map),
		newConn:   make(chan int),
	}
	return w
}

func (m *ConnMultiplexer) GetReaderWriterCloser(id int) io.ReadWriteCloser {
	w := &mwr{
		id:      id,
		wr:      m.mul,
		red:     m.mul,
		wbuf:    new(bytes.Buffer),
		rbuf:    new(bytes.Buffer),
		newData: make(ClosureCh),
	}
	m.bufMap.Store(id, w.rbuf)
	m.readerMap.Store(id, w.newData)

	return w
}

func (m *ConnMultiplexer) NewConn() <-chan int {
	return m.newConn
}
