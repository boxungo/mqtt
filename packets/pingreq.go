package packets

import (
	"fmt"
	"io"
)

// PingreqPacket 心跳请求包
type PingreqPacket struct {
	FixedHeader
}

// String ...
func (p *PingreqPacket) String() string {
	str := fmt.Sprintf("%s", p.FixedHeader)
	return str
}

// Write 写入
func (p *PingreqPacket) Write(w io.Writer) error {
	var err error
	packet := p.FixedHeader.pack()
	_, err = packet.WriteTo(w)

	return err
}

// Unpack 解包
func (p *PingreqPacket) Unpack(r io.Reader) error {
	return nil
}
