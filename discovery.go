package eiscp

import (
	"log"
	"net"
	"time"
)

func Discover(timeout time.Duration, errs chan<- error) <-chan Device {
	d := newDiscoverer(errs)
	if d.start() {
		// Shutdown the connection after the timeout.
		time.AfterFunc(timeout, func() { d.conn.Close() })
	}
	return d.devices
}

const discoveryPort = 60128

var discoverListenAddr = &net.UDPAddr{IP: net.IPv4zero, Port: 0}
var discoverBroadcastAddr = &net.UDPAddr{IP: net.IPv4bcast, Port: discoveryPort}
var discoverPacket = encodePacket("ECNQSTN", CategoryAny)

type discoverer struct {
	conn    *net.UDPConn
	devices chan Device
	errs    chan<- error
}

func newDiscoverer(errs chan<- error) *discoverer {
	d := &discoverer{devices: make(chan Device), errs: errs}
	return d
}

func (d *discoverer) start() bool {
	var err error
	d.conn, err = net.ListenUDP("udp", discoverListenAddr)
	if d.isErr(err) {
		log.Println("x")
		close(d.devices)
		return false
	}

	log.Println("Sending discovery packet:\n" + discoverPacket.debug())

	_, err = d.conn.WriteToUDP(discoverPacket.bytes(), discoverBroadcastAddr)
	if d.isErr(err) {
		log.Println("y")
		d.conn.Close()
		close(d.devices)
		return false
	}

	go d.monitor()
	return true
}

func (d *discoverer) monitor() {
	log.Println("Listening...")
	data := make([]byte, maxPacketSize)
	for {
		msglen, from, err := d.conn.ReadFromUDP(data)
		if d.isErr(err) {
			break
		}
		p := packet(data[:msglen])
		// log.Println("\n" + pretty(p.hex()))
		if !p.equals(discoverPacket) {
			d.createDevice(p, from)
		}
	}
	log.Println("Done")
	d.conn.Close()
	close(d.devices)
}

func (d *discoverer) createDevice(p packet, from *net.UDPAddr) {
	var info DeviceInfo
	if d.isErr(p.parseInfo(&info)) {
		return
	}

	device, err := newDevice(from, info)
	if d.isErr(err) {
		return
	}

	d.devices <- device
}

func (d *discoverer) isErr(e error) bool {
	if e != nil {
		if d.errs != nil {
			d.errs <- e
		} else {
			log.Println(e)
		}
		return true
	}
	return false
}
