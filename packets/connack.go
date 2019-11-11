package packets

import (
	"bytes"
	"fmt"
	"io"
)

// ConnackPacket 连接确认包
type ConnackPacket struct {
	FixedHeader
	SessionPresent bool
	ReturnCode     byte
}

// String ...
func (p *ConnackPacket) String() string {
	str := fmt.Sprintf("%s", p.FixedHeader)
	str += " "
	str += fmt.Sprintf("SessionPresent: %t ReturnCode: %d", p.SessionPresent, p.ReturnCode)
	return str
}

// Write 写入
func (p *ConnackPacket) Write(w io.Writer) error {
	var body bytes.Buffer
	var err error

	body.WriteByte(boolToByte(p.SessionPresent))
	body.WriteByte(p.ReturnCode)

	p.FixedHeader.RemainingLength = body.Len()
	packet := p.FixedHeader.pack()

	packet.Write(body.Bytes())
	_, err = packet.WriteTo(w)

	return err
}

// Unpack 解包
func (p *ConnackPacket) Unpack(r io.Reader) error {
	flags, err := decodeByte(r)
	if err != nil {
		return err
	}
	p.SessionPresent = 1&flags > 0
	p.ReturnCode, err = decodeByte(r)
	if err != nil {
		return err
	}
	return nil
}
