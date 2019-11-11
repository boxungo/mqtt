package packets

import (
	"bytes"
	"fmt"
	"io"
)

//连接错误码
const (
	Accepted                        = 0x00
	ErrRefusedBadProtocolLevel      = 0x01
	ErrRefusedIDRejected            = 0x02
	ErrRefusedServerUnavailable     = 0x03
	ErrRefusedBadUsernameOrPassword = 0x04
	ErrRefusedNotAuthorised         = 0x05
	ErrNetworkError                 = 0xFE
	ErrProtocolViolation            = 0xFF
)

// ConnectPacket 连接包
type ConnectPacket struct {
	FixedHeader
	ProtocolName  string // 协议名
	ProtocolLevel byte   // 协议级别
	CleanSession  bool   // 清除会话
	WillFlag      bool   // 遗嘱标志
	WillQos       byte   // 遗嘱QoS
	WillRetain    bool   // 遗嘱保留
	UsernameFlag  bool   // 用户名标志
	PasswordFlag  bool   // 密码标志
	Reserved      byte   // 保留位
	KeepAlive     uint16 // 保持连接

	// 有效负载 Payload
	ClientIdentifier string // 客户端标识符
	WillTopic        string // 遗嘱主题
	WillMessage      []byte // 遗嘱消息
	Username         string // 用户名
	Password         []byte // 密码
}

// String ...
func (p *ConnectPacket) String() string {
	str := fmt.Sprintf("%s", p.FixedHeader)
	str += " "
	str += fmt.Sprintf("protocolversion: %d protocolname: %s cleansession: %t willflag: %t WillQos: %d WillRetain: %t Usernameflag: %t Passwordflag: %t keepalive: %d clientId: %s willtopic: %s willmessage: %s Username: %s Password: %s", p.ProtocolLevel, p.ProtocolName, p.CleanSession, p.WillFlag, p.WillQos, p.WillRetain, p.UsernameFlag, p.PasswordFlag, p.KeepAlive, p.ClientIdentifier, p.WillTopic, p.WillMessage, p.Username, p.Password)
	return str
}

// Write 写入
// Bit     7          6          5          4          3          2          1          0
//	   User Name   Password   Will Retain     Will QoS         Will Flag   Clean     Reserved
//	   Flag        Flag                                                    Session
func (p *ConnectPacket) Write(w io.Writer) error {
	var body bytes.Buffer
	var err error

	body.Write(encodeString(p.ProtocolName))
	body.WriteByte(p.ProtocolLevel)
	body.WriteByte(boolToByte(p.CleanSession)<<1 | boolToByte(p.WillFlag)<<2 | p.WillQos<<3 | boolToByte(p.WillRetain)<<5 | boolToByte(p.PasswordFlag)<<6 | boolToByte(p.UsernameFlag)<<7)
	body.Write(encodeUint16(p.KeepAlive))
	body.Write(encodeString(p.ClientIdentifier))
	if p.WillFlag {
		body.Write(encodeString(p.WillTopic))
		body.Write(encodeBytes(p.WillMessage))
	}
	if p.UsernameFlag {
		body.Write(encodeString(p.Username))
	}
	if p.PasswordFlag {
		body.Write(encodeBytes(p.Password))
	}
	p.FixedHeader.RemainingLength = body.Len()
	packet := p.FixedHeader.pack()
	packet.Write(body.Bytes())
	_, err = packet.WriteTo(w)

	return err
}

// Unpack 解包
func (p *ConnectPacket) Unpack(r io.Reader) error {
	var err error

	p.ProtocolName, err = decodeString(r)
	if err != nil {
		return err
	}
	p.ProtocolLevel, err = decodeByte(r)
	if err != nil {
		return err
	}
	connectFlags, err := decodeByte(r)
	if err != nil {
		return err
	}
	p.Reserved = 1 & connectFlags
	p.CleanSession = 1&(connectFlags>>1) > 0
	p.WillFlag = 1&(connectFlags>>2) > 0
	p.WillQos = 3 & (connectFlags >> 3)
	p.WillRetain = 1&(connectFlags>>5) > 0
	p.PasswordFlag = 1&(connectFlags>>6) > 0
	p.UsernameFlag = 1&(connectFlags>>7) > 0
	p.KeepAlive, err = decodeUint16(r)
	if err != nil {
		return err
	}
	p.ClientIdentifier, err = decodeString(r)
	if err != nil {
		return err
	}
	if p.WillFlag {
		p.WillTopic, err = decodeString(r)
		if err != nil {
			return err
		}
		p.WillMessage, err = decodeBytes(r)
		if err != nil {
			return err
		}
	}
	if p.UsernameFlag {
		p.Username, err = decodeString(r)
		if err != nil {
			return err
		}
	}
	if p.PasswordFlag {
		p.Password, err = decodeBytes(r)
		if err != nil {
			return err
		}
	}
	return nil
}

// Validate 验证
func (p *ConnectPacket) Validate() byte {
	if p.PasswordFlag && !p.UsernameFlag {
		return ErrRefusedBadUsernameOrPassword
	}
	if p.Reserved != 0 {
		return ErrProtocolViolation
	}
	if p.ProtocolName != "MQTT" {
		return ErrProtocolViolation
	}
	// >= 3.1.1
	if p.ProtocolLevel != 0x04 {
		return ErrRefusedBadProtocolLevel
	}

	if len(p.ClientIdentifier) > 65535 || len(p.Username) > 65535 || len(p.Password) > 65535 {
		return ErrProtocolViolation
	}
	if len(p.ClientIdentifier) == 0 && !p.CleanSession {
		return ErrRefusedIDRejected
	}

	return Accepted
}
