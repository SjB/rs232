package main

import (
    "fmt"
    "github.com/SjB/rs232"
)

func main() {
    cfg := rs232.Config{"/dev/ttyUSB0",
        38400,
        rs232.ParityNone,
        8,
        rs232.StopBits1,
        10}

    s, err := rs232.Open("/dev/ttyUSB0")
    if err != nil {
        panic(err)
    }
    s.Configure(cfg)
    defer s.Close()

    packet := []byte{0x02, 0x49, 0x3, 0x42, 0x31}
    n, err := s.Write(packet)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Send %d bytes. [%X]\n", n, packet)
    data := make([]byte, 128)
    n, err = s.Read(data)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Read %d bytes. [%X]\n", n, data[:n])
}
