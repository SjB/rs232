// Package rs232 create a rs232 connection.
package rs232

import (
	"io"
	"os"
	"syscall"

	"github.com/pkg/term/termios"
)

type Parity byte

const (
	ParityNone = Parity(iota)
	ParityEven
	ParityOdd
)

type StopBits byte

const (
	StopBits1 = StopBits(iota)
	StopBits2
)

// Rs232 represents the rs232 communication device.
type Rs232 struct {
	config Config
	fd     *os.File
	orig   syscall.Termios // current state of the terminal.
}

// Config represents the rs232 configuration for the connection.
type Config struct {
	Name        string
	BaudRate    int
	Parity      Parity
	DataBits    int
	StopBits    StopBits
	ReadTimeout int
}

func setParity(attr *syscall.Termios, parity Parity) {

	switch parity {
	case ParityNone:
		attr.Cflag &^= syscall.PARENB
	case ParityEven:
		attr.Cflag |= syscall.PARENB
		attr.Cflag &^= syscall.PARODD
	case ParityOdd:
		attr.Cflag |= syscall.PARENB
		attr.Cflag |= syscall.PARODD
	}
}

func setStopBits(attr *syscall.Termios, stopbits StopBits) {
	switch stopbits {
	case StopBits1:
		attr.Cflag &^= syscall.CSTOPB
	case StopBits2:
		attr.Cflag |= syscall.CSTOPB
	}
}

func setDataBits(attr *syscall.Termios, databits int) {
	var bits uint32
	switch databits {
	case 5:
		bits = syscall.CS5
	case 6:
		bits = syscall.CS6
	case 7:
		bits = syscall.CS7
	case 8:
		bits = syscall.CS8
	}
	attr.Cflag &^= syscall.CSIZE
	attr.Cflag |= bits
}

func setBaudRate(attr *syscall.Termios, baud int) error {
	var rate uint32

	switch baud {
	case 50:
		rate = syscall.B50
	case 75:
		rate = syscall.B75
	case 110:
		rate = syscall.B110
	case 134:
		rate = syscall.B134
	case 150:
		rate = syscall.B150
	case 200:
		rate = syscall.B200
	case 300:
		rate = syscall.B300
	case 600:
		rate = syscall.B600
	case 1200:
		rate = syscall.B1200
	case 1800:
		rate = syscall.B1800
	case 2400:
		rate = syscall.B2400
	case 4800:
		rate = syscall.B4800
	case 9600:
		rate = syscall.B9600
	case 19200:
		rate = syscall.B19200
	case 38400:
		rate = syscall.B38400
	case 57600:
		rate = syscall.B57600
	case 115200:
		rate = syscall.B115200
	case 230400:
		rate = syscall.B230400
	case 460800:
		rate = syscall.B460800
	case 500000:
		rate = syscall.B500000
	case 576000:
		rate = syscall.B576000
	case 921600:
		rate = syscall.B921600
	case 1000000:
		rate = syscall.B1000000
	case 1152000:
		rate = syscall.B1152000
	case 1500000:
		rate = syscall.B1500000
	case 2000000:
		rate = syscall.B2000000
	case 2500000:
		rate = syscall.B2500000
	case 3000000:
		rate = syscall.B3000000
	case 3500000:
		rate = syscall.B3500000
	case 4000000:
		rate = syscall.B4000000
	default:
		return syscall.EINVAL
	}
	attr.Cflag |= rate
	attr.Ispeed = rate
	attr.Ospeed = rate
	return nil
}

// Open opens an rs232 communication.
func Open(cfg Config) (*Rs232, error) {
	fd, err := os.OpenFile(cfg.Name, syscall.O_NOCTTY|syscall.O_RDWR|syscall.O_NONBLOCK, 0666)
	if err != nil {
		return nil, &os.PathError{"open", cfg.Name, err}
	}
	s := Rs232{config: cfg, fd: fd}
	if err = termios.Tcgetattr(uintptr(fd.Fd()), &s.orig); err != nil {
		return nil, err
	}
	if err = s.Configure(cfg); err != nil {
		return nil, err
	}

	if err = syscall.SetNonblock(int(fd.Fd()), false); err != nil {
		return nil, err
	}
	return &s, nil
}

// Configure allows the communication configuration to be changed.
func (s *Rs232) Configure(cfg Config) error {

	var attr syscall.Termios

	if err := termios.Tcgetattr(uintptr(s.fd.Fd()), &attr); err != nil {
		return err
	}

	// Select raw mode
	attr.Iflag &^= (syscall.IGNBRK | syscall.BRKINT | syscall.PARMRK | syscall.INLCR | syscall.IGNCR | syscall.ICRNL | syscall.ISTRIP | syscall.IXON)
	attr.Lflag &^= (syscall.ECHO | syscall.ECHONL | syscall.ICANON | syscall.IEXTEN | syscall.ISIG)
	attr.Oflag &^= syscall.OPOST
	attr.Cflag &^= syscall.CSIZE

	setBaudRate(&attr, cfg.BaudRate)
	setParity(&attr, cfg.Parity)
	setStopBits(&attr, cfg.StopBits)
	setDataBits(&attr, cfg.DataBits)

	if s.config.ReadTimeout > 0 {
		attr.Cc[syscall.VMIN] = 0
		attr.Cc[syscall.VTIME] = uint8(s.config.ReadTimeout)
	} else {
		attr.Cc[syscall.VMIN] = 1
		attr.Cc[syscall.VTIME] = 0
	}
	s.config = cfg
	return termios.Tcsetattr(uintptr(s.fd.Fd()), termios.TCSANOW, (*syscall.Termios)(&attr))
}

// Close closes the communication.
func (s *Rs232) Close() error {
	s.Restore() // try to restore the device
	return s.fd.Close()
}

// Read reads up to len(b) bytes from the rs232 device. Returning
// the number of bytes read and an error, if any. The Read can be timeout
// by setting the ReadTimeout value.
func (s *Rs232) Read(b []byte) (int, error) {
	n, err := s.fd.Read(b)
	if n < 0 {
		n = 0
	}
	if n == 0 && len(b) > 0 && err == nil {
		return 0, io.EOF
	} else if err == io.EOF {
		return n, err
	} else if err != nil {
		return n, &os.PathError{"read", s.config.Name, err}
	}
	return n, nil
}

// Write data through serial communication device
func (s *Rs232) Write(b []byte) (int, error) {
	n, err := s.fd.Write(b)
	if n < 0 {
		n = 0
	}
	if n != len(b) {
		return n, io.ErrShortWrite
	}
	if err != nil {
		return n, &os.PathError{"write", s.config.Name, err}
	}
	return n, nil
}

// Restore restores the state of the terminal captured at the point that
// the terminal was originally opened.
func (s *Rs232) Restore() error {
	return termios.Tcsetattr(uintptr(s.fd.Fd()), termios.TCIOFLUSH, &s.orig)
}

// Flush flushes both data received but not read, and data written but not transmitted.
func (s *Rs232) Flush() error {
	return termios.Tcflush(uintptr(s.fd.Fd()), termios.TCIOFLUSH)
}

// SendBreak sends a break signal.
func (s *Rs232) SendBreak() error {
	return termios.Tcsendbreak(uintptr(s.fd.Fd()), 0)
}
