package ebpf

import (
	"bytes"
	"encoding/binary"
	"log"
	"net"
	"strconv"
	"time"
)

type decoder struct {
	v16 uint16
	v32 uint32
	v64 uint64

	c uint16

	v4 bool

	ip net.IP
}

func newDecoder(v4 bool) *decoder {
	size := 4
	if !v4 {
		size = 32
	}

	return &decoder{
		ip: make(net.IP, size),
		v4: v4,
	}
}

func (d *decoder) decode(data []byte, fields []string, buf *bytes.Buffer) {
	var prop FieldAttrs

	d.c = 0

	buf.WriteRune('{')

	for _, field := range fields {
		if d.v4 {
			prop = fieldsModel4[field]
		} else {
			prop = fieldsModel6[field]
		}

		buf.WriteRune('"')
		buf.Write([]byte(field))
		buf.WriteRune('"')
		buf.WriteRune(':')

		switch prop.CType {
		case u8:

			buf.Write([]byte(strconv.FormatUint(uint64(data[d.c]), 10)))
			buf.WriteRune(',')

			d.c++

		case u16:
			if d.c%2 > 0 {
				d.c += (2 - (d.c % 2))
			}

			d.v16 = bytesToUint16(prop.BigEndian, data, d.c)

			buf.Write([]byte(strconv.FormatUint(uint64(d.v16), 10)))
			buf.WriteRune(',')

			d.c += 2

		case u32:
			if d.c%4 > 0 {
				d.c += (4 - (d.c % 4))
			}

			if prop.DType == IP {
				d.ip = data[d.c : d.c+4]
				buf.WriteRune('"')
				buf.Write([]byte(d.ip.String()))
				buf.WriteRune('"')
			} else {
				d.v32 = bytesToUint32(prop.BigEndian, data, d.c)
				buf.Write([]byte(strconv.FormatUint(uint64(d.v32), 10)))
			}

			buf.WriteRune(',')

			d.c += 4

		case u64:
			if d.c%8 > 0 {
				d.c += (8 - (d.c % 8))
			}

			d.v64 = bytesToUint64(prop.BigEndian, data, d.c)

			buf.Write([]byte(strconv.FormatUint(d.v64, 10)))
			buf.WriteRune(',')

			d.c += 8

		case u128:
			if d.c%16 > 0 {
				d.c += (16 - (d.c % 16))
			}

			d.ip = data[d.c : d.c+16]
			buf.WriteRune('"')
			buf.Write([]byte(d.ip.String()))
			buf.WriteRune('"')
			buf.WriteRune(',')

			d.c += 16

		case char:
			// TODO padding

			buf.WriteRune('"')
			buf.Write(trim(data[d.c : d.c+16]))
			buf.WriteRune('"')
			buf.WriteRune(',')

			d.c += 16

		default:
			log.Fatal("unknown data type")
		}
	}

	buf.WriteRune('"')
	buf.Write([]byte("Timestamp"))
	buf.WriteRune('"')
	buf.WriteRune(':')
	buf.Write([]byte(strconv.FormatInt(time.Now().Unix(), 10)))
	buf.WriteRune('}')
}

func bytesToUint16(isBigEndian bool, data []byte, index uint16) uint16 {
	if !isBigEndian {
		return binary.LittleEndian.Uint16(data[index:])
	}
	return binary.BigEndian.Uint16(data[index:])
}

func bytesToUint32(isBigEndian bool, data []byte, index uint16) uint32 {
	if !isBigEndian {
		return binary.LittleEndian.Uint32(data[index:])
	}
	return binary.BigEndian.Uint32(data[index:])
}

func bytesToUint64(isBigEndian bool, data []byte, index uint16) uint64 {
	if !isBigEndian {
		return binary.LittleEndian.Uint64(data[index:])
	}
	return binary.BigEndian.Uint64(data[index:])
}

func trim(b []byte) []byte {
	for i := 0; i < len(b); i++ {
		if b[i] == '\x00' {
			return b[:i]
		}
	}
	return b
}
