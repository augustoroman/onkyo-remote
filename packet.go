package eiscp

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/augustoroman/hexdump"
	"github.com/augustoroman/multierror"
	"strconv"
	"strings"
)

const maxPacketSize = 4096

const headerSize = 16

const magic = "ISCP"
const terminator = "\000\r\n"
const version = 1

func init() {
	fmt.Printf("TERM: %q  %x\n", terminator, terminator)
}

type packet []byte

func (p packet) bytes() []byte        { return []byte(p) }
func (p packet) hex() string          { return hex.EncodeToString([]byte(p)) }
func (p packet) magic() string        { return string(p[0:4]) }
func (p packet) headerSize() uint32   { return binary.BigEndian.Uint32(p[4:8]) }
func (p packet) messageLen() int      { return int(binary.BigEndian.Uint32(p[8:12])) }
func (p packet) version() uint8       { return p[12] }
func (p packet) equals(o packet) bool { return bytes.Equal([]byte(p), []byte(o)) }
func (p packet) debug() string        { return hexdump.Dump(p.bytes()) }

func (p packet) message() []byte { return p[16 : 16+p.messageLen()] }
func (p packet) Message() Message {
	return Message(bytes.TrimRight(p.message(), terminator+"\x1a\x0d"))
}

func (p packet) String() string { return fmt.Sprintf("%s", strings.TrimSpace(string(p))) }

func (p packet) validateHeader() error {
	if len(p) < headerSize {
		return fmt.Errorf("Incomplete header (%d bytes): %v", len(p), p)
	}
	if p.magic() != magic {
		return fmt.Errorf("Malformed header (bad magic string \"%s\"): %v", p.magic(), p)
	}
	if p.headerSize() != headerSize {
		return fmt.Errorf("Bad header size (%d bytes): %v", p.headerSize(), p)
	}
	if p.version() != version {
		return fmt.Errorf("Unknown version (%d): %v", p.version(), p)
	}
	return nil
}

// Requires a valid header.
func (p packet) validateMessage() error {
	if len(p) < headerSize+p.messageLen() {
		return fmt.Errorf("Packet too short (%d bytes): %v", len(p), p)
	}
	return nil
}
func (p packet) validate() error {
	if err := p.validateHeader(); err != nil {
		return err
	}
	if err := p.validateMessage(); err != nil {
		return err
	}
	return nil
}

func (p packet) parseInfo(info *DeviceInfo) error {
	err := p.validate()
	if err != nil {
		return err
	}

	// Expecting:
	//   !cECNnnnnnn/ppppp/dd/iiiiiiiiiiiitt
	// where
	//   c = device type
	//   nnnnnn = device model (up to 64 char)
	//   ppppp = port # (5 digits)
	//   dd = destination area (2 char)
	//   iiiiii = identifier (up to 12 char)
	//   tt = terminating chars (one or both of \r\n)
	parts := bytes.Split(p.message(), []byte("/"))
	if len(parts) != 4 ||
		len(parts[0]) < 5 || len(parts[1]) != 5 || len(parts[2]) != 2 ||
		parts[0][0] != '!' || !bytes.Equal(parts[0][2:5], []byte("ECN")) {
		return fmt.Errorf("Malformed device info: %v", p)
	}

	info.Category = DeviceCategory(parts[0][1])
	info.Model = string(parts[0][5:])
	info.Port, err = strconv.Atoi(string(parts[1]))
	info.DestArea = string(parts[2])
	info.Identifier = strings.TrimSpace(string(parts[3]))

	return err
}

func encodePacket(message string, cat DeviceCategory) packet {
	var b bytes.Buffer

	// start char + unit + message + endchar
	var messageLen uint32 = uint32(1 + 1 + len(message) + len(terminator))

	// Header:
	b.WriteString(magic)
	binary.Write(&b, binary.BigEndian, uint32(16))
	binary.Write(&b, binary.BigEndian, messageLen)
	b.Write([]byte{version, 0, 0, 0}) // version + reserved

	// Message:
	b.Write([]byte{'!', byte(cat)})
	b.WriteString(message)
	b.WriteString(terminator)

	return packet(b.Bytes())
}

func decodePackets(data []byte) ([]packet, error) {
	var errs multierror.MultiError
	var packets []packet

	for len(data) > 0 {
		if len(data) < headerSize {
			errs.Pushf("Incomplete packet, %d bytes remaining: %q", len(data), data)
			break
		}

		p := packet(data[:headerSize])
		if e := p.validateHeader(); e != nil {
			errs.Pushf("Bad packet: %v", e)
			break
		}
		totalSize := int(p.headerSize()) + p.messageLen()
		p, data = packet(data[:totalSize]), data[totalSize:]
		packets = append(packets, p)
	}
	return packets, errs.Error()
}
