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

func (d *decoder) decode(data []byte, metrics []string, buf *bytes.Buffer) {
	var prop FieldAttrs

	d.c = 0

	buf.WriteRune('{')

	for _, metric := range metrics {
		if d.v4 {
			prop = fieldsModel4[metric]
		} else {
			prop = fieldsModel6[metric]
		}

		buf.WriteRune('"')
		buf.Write([]byte(metric))
		buf.WriteRune('"')
		buf.WriteRune(':')

		switch prop.CType {
		case u16:

			if !prop.DSNP {
				d.v16 = binary.LittleEndian.Uint16(data[d.c:])
			} else {
				d.v16 = binary.BigEndian.Uint16(data[d.c:])
			}

			buf.Write([]byte(strconv.FormatUint(uint64(d.v16), 10)))
			buf.WriteRune(',')

			d.c += 2

		case u32:
			if d.c%4 > 0 {
				d.c += (d.c % 4)
			}

			if prop.DType == IP {
				d.ip = data[d.c : d.c+4]
				buf.WriteRune('"')
				buf.Write([]byte(d.ip.String()))
				buf.WriteRune('"')
			} else {
				if !prop.DSNP {
					d.v32 = binary.LittleEndian.Uint32(data[d.c:])
				} else {
					d.v32 = binary.BigEndian.Uint32(data[d.c:])
				}
				buf.Write([]byte(strconv.FormatUint(uint64(d.v32), 10)))
			}

			buf.WriteRune(',')

			d.c += 4

		case u64:
			if d.c%8 > 0 {
				d.c += (8 - (d.c % 8))
			}

			if !prop.DSNP {
				d.v64 = binary.LittleEndian.Uint64(data[d.c:])
			} else {
				d.v64 = binary.BigEndian.Uint64(data[d.c:])
			}

			buf.Write([]byte(strconv.FormatUint(d.v64, 10)))
			buf.WriteRune(',')

			d.c += 8

		case u128:
			// TODO padding

			d.ip = data[d.c : d.c+16]
			buf.WriteRune('"')
			buf.Write([]byte(d.ip.String()))
			buf.WriteRune('"')
			buf.WriteRune(',')

			d.c += 16

		case char:
			// TODO padding

			buf.WriteRune('"')
			buf.Write(data[d.c : d.c+16])
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
