package packets

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// CONNECT 客户端请求连接服务端
// CONNACK 连接报文确认
// PUBLISH 发布消息
// PUBACK QoS 1消息发布收到确认
// PUBREC 发布收到（保证交付第一步）
// PUBREL 发布释放（保证交付第二步
// PUBCOMP QoS 2消息发布完成（保证交互第三步）
// SUBSCRIBE 客户端订阅请求
// SUBACK 订阅请求报文确认
// UNSUBSCRIBE 客户端取消订阅请求
// UNSUBACK 取消订阅报文确认
// PINGREQ 心跳请求
// PINGRESP 心跳响应
// DISCONNECT 客户端断开连接
const (
	CONNECT     = 1
	CONNACK     = 2
	PUBLISH     = 3
	PUBACK      = 4
	PUBREC      = 5
	PUBREL      = 6
	PUBCOMP     = 7
	SUBSCRIBE   = 8
	SUBACK      = 9
	UNSUBSCRIBE = 10
	UNSUBACK    = 11
	PINGREQ     = 12
	PINGRESP    = 13
	DISCONNECT  = 14
)

// PacketNames 包类型名称
var PacketNames = map[uint8]string{
	CONNECT:     "CONNECT",
	CONNACK:     "CONNACK",
	PUBLISH:     "PUBLISH",
	PUBACK:      "PUBACK",
	PUBREC:      "PUBREC",
	PUBREL:      "PUBREL",
	PUBCOMP:     "PUBCOMP",
	SUBSCRIBE:   "SUBSCRIBE",
	SUBACK:      "SUBACK",
	UNSUBSCRIBE: "UNSUBSCRIBE",
	UNSUBACK:    "UNSUBACK",
	PINGREQ:     "PINGREQ",
	PINGRESP:    "PINGRESP",
	DISCONNECT:  "DISCONNECT",
}

// FixedHeader 固定头部
type FixedHeader struct {
	PacketType      byte // MQTT控制报文的类型
	Dup             bool // 控制报文的重复分发标志
	Qos             byte // PUBLISH报文的服务质量等级
	Retain          bool // PUBLISH报文的保留标志
	RemainingLength int  // 剩余长度
}

func (fh FixedHeader) String() string {
	return fmt.Sprintf("%s: dup: %t qos: %d retain: %t rLength: %d", PacketNames[fh.PacketType], fh.Dup, fh.Qos, fh.Retain, fh.RemainingLength)
}

// pack 打包固定头部
func (fh *FixedHeader) pack() bytes.Buffer {
	var header bytes.Buffer
	header.WriteByte(fh.PacketType<<4 | boolToByte(fh.Dup)<<3 | fh.Qos<<1 | boolToByte(fh.Retain))
	header.Write(encodeRemainingLength(fh.RemainingLength))
	return header
}

// unpack 打开固定头部
func (fh *FixedHeader) unpack(r io.Reader) error {
	b := make([]byte, 1)
	_, err := r.Read(b)
	if err != nil {
		return err
	}

	fh.PacketType = b[0] >> 4
	fh.Dup = (b[0]>>3)&0x01 > 0
	fh.Qos = (b[0] >> 1) & 0x03
	fh.Retain = b[0]&0x01 > 0

	fh.RemainingLength, err = decodeRemainingLength(r)

	return err
}

// ControlPacket 控制报文接口
type ControlPacket interface {
	Write(io.Writer) error
	Unpack(io.Reader) error
	String() string
}

// ReadPacket 读包
func ReadPacket(r io.Reader) (ControlPacket, error) {
	var fh FixedHeader

	err := fh.unpack(r)
	if err != nil {
		return nil, err
	}

	packet, err := NewControlPacketWithHeader(fh)
	if err != nil {
		return nil, err
	}

	// 验证数据是否正常
	if fh.RemainingLength > 0 {
		packetBytes := make([]byte, fh.RemainingLength)
		n, err := io.ReadFull(r, packetBytes)
		if err != nil {
			return nil, err
		}
		if n != fh.RemainingLength {
			return nil, fmt.Errorf("Failed to read expected data")
		}

		err = packet.Unpack(bytes.NewBuffer(packetBytes))
		if err != nil {
			return nil, err
		}
	}

	return packet, nil
}

// NewControlPacket 新建控制报文
func NewControlPacket(packetType byte) ControlPacket {
	if packetType < 1 || packetType > 14 {
		return nil
	}
	fh := FixedHeader{PacketType: packetType}

	switch packetType {
	case PUBREL:
		fh.Qos = 1
	case SUBSCRIBE:
		fh.Qos = 1
	case UNSUBSCRIBE:
		fh.Qos = 1
	}
	packet, _ := NewControlPacketWithHeader(fh)

	return packet
}

// NewControlPacketWithHeader 根据固定头部信息新建控制报文
func NewControlPacketWithHeader(fh FixedHeader) (ControlPacket, error) {
	switch fh.PacketType {
	case CONNECT:
		return &ConnectPacket{FixedHeader: fh}, nil
	case CONNACK:
		return &ConnackPacket{FixedHeader: fh}, nil
	case PUBLISH:
		return &PublishPacket{FixedHeader: fh}, nil
	case PUBACK:
		return &PubackPacket{FixedHeader: fh}, nil
	case PUBREC:
		return &PubrecPacket{FixedHeader: fh}, nil
	case PUBREL:
		return &PubrelPacket{FixedHeader: fh}, nil
	case PUBCOMP:
		return &PubcompPacket{FixedHeader: fh}, nil
	case SUBSCRIBE:
		return &SubscribePacket{FixedHeader: fh}, nil
	case SUBACK:
		return &SubackPacket{FixedHeader: fh}, nil
	case UNSUBSCRIBE:
		return &UnsubscribePacket{FixedHeader: fh}, nil
	case UNSUBACK:
		return &UnsubackPacket{FixedHeader: fh}, nil
	case PINGREQ:
		return &PingreqPacket{FixedHeader: fh}, nil
	case PINGRESP:
		return &PingrespPacket{FixedHeader: fh}, nil
	case DISCONNECT:
		return &DisconnectPacket{FixedHeader: fh}, nil
	}
	return nil, fmt.Errorf("unsupported packet type 0x%x", fh.PacketType)
}

// boolToByte 把布尔值转成字节
func boolToByte(b bool) byte {
	switch b {
	case true:
		return 1
	default:
		return 0
	}
}

// byteToBool 把字节转成布尔值
func byteToBool(b byte) bool {
	switch b {
	case 0:
		return false
	default:
		return true
	}
}

// decodeByte 从Reader读取一个字节
func decodeByte(r io.Reader) (byte, error) {
	num := make([]byte, 1)
	_, err := r.Read(num)
	if err != nil {
		return 0, err
	}
	return num[0], nil
}

// encodeBytes 把内容编码成二进制
func encodeBytes(value []byte) []byte {
	bytes := make([]byte, 2)
	// 先写入两字节的内容长度
	binary.BigEndian.PutUint16(bytes, uint16(len(value)))
	// 再写入内容
	return append(bytes, value...)
}

// decodeBytes 从Reader读取内容
func decodeBytes(r io.Reader) ([]byte, error) {
	// 先读取两字节的内容长度
	length, err := decodeUint16(r)
	if err != nil {
		return nil, err
	}
	// 再根据长度读取内容
	value := make([]byte, length)
	_, err = r.Read(value)
	if err != nil {
		return nil, err
	}
	return value, nil
}

// encodeUint16 把uint16编码成二进制
func encodeUint16(value uint16) []byte {
	bytes := make([]byte, 2)
	binary.BigEndian.PutUint16(bytes, value)
	return bytes
}

// decodeUint16 从Reader读出uint16的数字
func decodeUint16(r io.Reader) (uint16, error) {
	bytes := make([]byte, 2)
	_, err := r.Read(bytes)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(bytes), nil
}

// encodeString 把字符串编码成二进制
func encodeString(value string) []byte {
	return encodeBytes([]byte(value))
}

// decodeString 从Reader读出字符串内容
// 字符串的编码规则是: 前两字节为字符串长度
func decodeString(r io.Reader) (string, error) {
	value, err := decodeBytes(r)
	return string(value), err
}

// 1个字节时，从0(0x00)到127(0x7f)
// 2个字节时，从128(0x80,0x01)到16383(0Xff,0x7f)
// 3个字节时，从16384(0x80,0x80,0x01)到2097151(0xFF,0xFF,0x7F)
// 4个字节时，从2097152(0x80,0x80,0x80,0x01)到268435455(0xFF,0xFF,0xFF,0x7F)
// encodeRemainingLength 编码剩余长度
func encodeRemainingLength(length int) []byte {
	var encLength []byte
	for {
		digit := byte(length % 128)
		length /= 128
		if length > 0 {
			digit |= 0x80
		}
		encLength = append(encLength, digit)
		if length == 0 {
			break
		}
	}
	return encLength
}

// decodeRemainingLength 解码剩余长度
func decodeRemainingLength(r io.Reader) (int, error) {
	b := make([]byte, 1)
	multiplier := 1
	value := 0
	for {
		_, err := io.ReadFull(r, b)
		if err != nil {
			return 0, err
		}

		digit := b[0]
		value += int(digit&127) * multiplier
		multiplier *= 128
		if (digit & 128) == 0 {
			break
		}
		if multiplier > 128*128*128 {
			return 0, errors.New("Malformed Remaining Length")
		}
	}

	return value, nil
}
