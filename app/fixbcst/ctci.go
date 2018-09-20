package fixbcst

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"net"
	"time"

	"github.com/lunixbochs/struc"
)

type CtciHDR struct {
	MsgLen int `struc:"int16,little"`
	//Seq    int `struc:"uint32,little"`
}
type CtciLR struct {
	Hdr          string `struc:"[2]byte,little"`
	LastRecvSeq  int    `struc:"int32,little"`
	LastRecvSeq2 int    `struc:"int32,little"`
	StartFlag    int    `struc:"int32,little"`
}

type CtciER struct {
	Hdr  string `struc:"[2]byte,little"`
	Code int    `struc:"int32,little"`
}

type CtciAK struct {
	Hdr         string `struc:"[2]byte,little"`
	LastRecvSeq uint32 `struc:"uint32,little"`
}

var (
	ErrInvalidMsgLenByte = errors.New("ctci: invalid Message HDR Byte")
)

// Ctci tcp protocol for ETS/DTS
type Ctci struct {
	Conn       net.Conn
	connString string
	Outgoing   chan []byte
	Incoming   chan []byte
	stopReader chan bool
	stopWriter chan bool
	stopAcker  chan bool
	errCh      chan error
	stop       chan bool
	logging    Logging
}

// newConn
func CtciNew(connString string, logging Logging) *Ctci {
	return &Ctci{
		connString: connString,
		Outgoing:   make(chan []byte, 128),
		Incoming:   make(chan []byte, 128),
		stopReader: make(chan bool, 1),
		stopWriter: make(chan bool, 1),
		stopAcker:  make(chan bool, 1),
		errCh:      make(chan error, 1),
		stop:       make(chan bool, 1),
		logging:    logging,
	}
}

// Start initializes goroutines to read responses and process messages
func (ctci *Ctci) connect() {
	Established := false
	for {
		c, err := net.Dial("tcp", ctci.connString)
		if err != nil {
			ctci.logging.logger.Errorf("Create Connection Fail!")
			time.Sleep(1 * time.Second)
		} else {
			ctci.Conn = c
			Established = true
			ctci.logging.logger.Noticef("Create Connection Successful!")
		}
		if Established {

			go ctci.reader()
			go ctci.writer()
			go ctci.acker()

			select {
			case <-ctci.errCh:
				ctci.stopWriter <- true
				ctci.stopAcker <- true
			case <-ctci.stop:
				ctci.stopReader <- true
				ctci.stopWriter <- true
				ctci.stopAcker <- true
				return
			}
			Established = false
		}

	}
}

// Dial
func (ctci *Ctci) Dial() error {
	go ctci.connect()
	return nil
}

// Close
func (ctci *Ctci) Close() {
	ctci.logging.logger.Noticef("Normal EXIT")
	ctci.stop <- true
	if ctci.Conn != nil {
		ctci.Conn.Close()
	}

}

func (ctci *Ctci) writer() {
	defer ctci.logging.logger.Noticef("FLUSHER has been stopped.")
	ctci.logging.logger.Noticef("FLUSHER has been created.")

	for {
		select {
		case out := <-ctci.Outgoing:
			var hdr CtciHDR
			hdr.MsgLen = len(out)
			var bb bytes.Buffer
			struc.Pack(&bb, &hdr)
			_, err := ctci.Conn.Write(bb.Bytes())
			if err != nil {
				ctci.logging.logger.Errorf("Outgoing error1 : ", err)
				break
			}
			_, err = ctci.Conn.Write(out)
			if err != nil {
				ctci.logging.logger.Errorf("Outgoing error2 : ", err)
				break
			}
		case <-ctci.stopWriter:
			return
		}
	}
}

func (ctci *Ctci) reader() {
	defer ctci.logging.logger.Noticef("READER has been stopped.")
	ctci.logging.logger.Noticef("READER has been created.")

	connReader := bufio.NewReader(ctci.Conn)
	for {
		var nMsgLen uint16
		err := binary.Read(connReader, binary.LittleEndian, &nMsgLen)
		if err != nil {
			ctci.logging.logger.Errorf("Incoming error1 : ", err)
			ctci.errCh <- err
			return
		}
		if nMsgLen > 1024 {
			ctci.logging.logger.Errorf("Incoming error with Length : ", nMsgLen)
			return
		}
		data := make([]byte, nMsgLen)
		err = binary.Read(connReader, binary.LittleEndian, data)
		if err != nil {
			ctci.logging.logger.Errorf("Incoming error2 : ", err)
			return
		}
		ctci.Incoming <- data
	}
}

func (ctci *Ctci) acker() {
	defer ctci.logging.logger.Noticef("ACKER has been stopped.")
	ctci.logging.logger.Noticef("ACKER has been created.")

	ticker := time.NewTicker(5 * time.Second).C
	for {
		select {
		case <-ticker:
			var ak CtciAK
			var bb bytes.Buffer
			ak.Hdr = "AK"
			ak.LastRecvSeq = 0
			struc.Pack(&bb, &ak)
			ctci.Outgoing <- bb.Bytes()
			ctci.logging.logger.Tracef("SEND AK")
		case <-ctci.stopAcker:
			return
		}
	}
}

// SendMsg to ctci
func (ctci *Ctci) SendMsg(msg []byte) {
	ctci.Outgoing <- msg
}
