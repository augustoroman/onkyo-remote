package eiscp

import (
	"log"
	"net"
	"runtime"
	"strings"
)

type Device interface {
	Info() DeviceInfo
	Messages() <-chan Message
	Send(cmd string, params ...string) error
}

type DeviceInfo struct {
	Model      string
	Category   DeviceCategory
	DestArea   string
	Identifier string
	Port       int
}

type DeviceCategory byte

func (d DeviceCategory) String() string {
	return string(d)
}

const CategoryDevice = DeviceCategory('1')
const CategoryAny = DeviceCategory('x')

type device struct {
	conn   net.Conn
	info   DeviceInfo
	recv   chan Message
	remote net.Addr
}

func newDevice(addr net.Addr, info DeviceInfo) (*device, error) {
	log.Println(addr)
	conn, err := net.Dial("tcp", addr.String())
	if err != nil {
		return nil, err
	}
	d := &device{
		conn:   conn,
		info:   info,
		recv:   make(chan Message),
		remote: addr,
	}

	go d.listen()

	return d, nil
}

func (r *device) Info() DeviceInfo         { return r.info }
func (r *device) Messages() <-chan Message { return r.recv }
func (r *device) Send(cmd string, params ...string) error {
	cmdstring := cmd + strings.Join(params, "")
	packet := encodePacket(cmdstring, r.info.Category)
	_, err := r.conn.Write(packet.bytes())
	return err
}
func (r *device) listen() {
	runtime.SetFinalizer(r, func(r *device) { r.conn.Close() })
	data := make([]byte, maxPacketSize)
	for {
		datalen, err := r.conn.Read(data)
		if err != nil {
			log.Println(err)
			break
		}
		packets, err := decodePackets(data[:datalen])
		if err != nil {
			log.Println(err)
		}
		for _, pkt := range packets {
			r.recv <- pkt.Message()
		}
	}
	runtime.SetFinalizer(r, nil)
	r.conn.Close()
	close(r.recv)
}
