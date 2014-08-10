package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"

	"github.com/SjB/rs232"
)

const (
	NAK = 0x15
	ACK = 0x6
)

var serialport string

func init() {
	flag.StringVar(&serialport, "port", "/dev/ttyUSB0", "specify com port")
}

func main() {
	flag.Parse()

	cfg := rs232.Config{serialport,
		38400,
		rs232.ParityNone,
		8,
		rs232.StopBits1,
		5}

	s, err := rs232.Open(cfg)
	if err != nil {
		panic(err)
	}
	s.Configure(cfg)
	defer s.Close()

	Send(s, []byte{0x02, 0x34, 0x3, 0x43, 0x36})
	Send(s, []byte{0x02, 0x49, 0x3, 0x42, 0x31})
	Send(s, []byte{0x02, 0x40, 0x3, 0x43, 0x36})
}

func Send(rw io.ReadWriter, packet []byte) {
	n, err := rw.Write(packet)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Send %d bytes. [%X] ", n, packet)

	data, err := Read(rw)
	if err == io.EOF {
		Send(rw, packet)
	} else if err != nil {
		panic(err)
	} else if len(data) == 1 {
		fmt.Printf("Received 1 bytes [%X]\n", data)
	} else {
		fmt.Printf("Revceied %d bytes. [%s]\n", len(data), data)
	}
}

func Read(r io.Reader) ([]byte, error) {
	c := 0
	data := make([]byte, 256)
	for {
		n, err := r.Read(data[c:])
		c += n
		if err != nil {
			return nil, err
		} else if data[0] == ACK || data[0] == NAK {
			return data[:1], nil
		} else if bytes.HasPrefix(data, []byte{0x02}) && bytes.HasSuffix(data[:c-2], []byte{0x03}) {
			return data[:c], nil
		}
	}
}
