package dubbo

import (
	"bytes"
	"encoding/binary"
)

const (
	headerLength    = 16
	magicHigh       = byte(0xda)
	magicLow        = byte(0xbb)
	flagRequest     = byte(0x80)
	flagTwoWay      = byte(0x40)
	serializationID = byte(0x6)
)

// Dubbo ...
type Dubbo struct {
	buffer       *bytes.Buffer
	databuf      *bytes.Buffer
	header       []byte
	dubboVersion []byte
}

// NewDubbo ...
func NewDubbo(dubboVersion []byte) *Dubbo {
	return &Dubbo{bytes.NewBuffer([]byte("")), bytes.NewBuffer([]byte("")), make([]byte, headerLength), dubboVersion}
}

// Encode ...
func (dubbo *Dubbo) Encode(requestID uint64, interfaceName []byte, version []byte, method []byte, paramtypes []byte, args [][]byte) []byte {

	//data := "\"2.0.1\"\n\"com.alibaba.dubbo.performance.demo.provider.IHelloService\"\nnull\n\"hash\"\n\"Ljava/lang/String;\"\n\"raHleA61L5jdyRtMDS8qszHlbYu6ZlyaRl1JPGTkrdZx0w550DvM0DosWs8QI0UW9j02KdTRaMTeIEnJ3v7XB0Ro5WIqzwAX91XCkXUhXBSV8o2WJI8ggeNGA7eMJrKJutYVoleMR2lXVHm9NmWpF2yRUoCy8cgP7nrcZTRG9zjLGTOAtXUmS1LOIUcR4XtxQ6eWWZnbsibfwKozr1hGxatLwcnVsu3rWvFPK1ig9GkTpzChxSzaCgSa3tnGMpUaLyuoknJubmBoTns513njSl9FcZIHcsZvgTpDV6eW87eHg1xjqKLOFRTrkgDJWwa53RXeFOpORsylW5pg6mb6wtaFNqTBEYsXvYX4SlKsKoLZ2t0aD85Qb6BbZULtLWFoXXKqvrbu2mWLOoIuR9gMwTNwr1UqpsC1rdMPLioRLn9fb04ZWAG3q0ZiVuqcCDv55g8TxEzRhfrxpLfRCh1CHqgKGctxbve42jvsJVbxrxXUM36SpzfbBAPdrDjm7C1Q3QUDJAuoNT3qUWFtvYtomuDJBGDYn2DZsNHOV42IGQlBerAjfzXUx2HQN3jAQ6pdBNYRexuiLT7rGPz7UNNN96uWK6xa4SqcxbRfICxt8Aw9GF5SI9F3qRiH0Zd2F0nwYpBn5\"\n{\"path\":\"com.alibaba.dubbo.performance.demo.provider.IHelloService\"}"
	dubbo.databuf.Reset()
	dubbo.attach(dubbo.dubboVersion)
	dubbo.attach(interfaceName)
	dubbo.attach(version)
	dubbo.attach(method)
	dubbo.attach(paramtypes)
	for _, arg := range args {
		dubbo.databuf.WriteString("\"")
		dubbo.databuf.Write(arg)
		dubbo.databuf.WriteString("\"\n")
	}
	dubbo.attachJSON([]byte("path"), interfaceName)
	dubbo.makeHeader(uint32(dubbo.databuf.Len()), requestID)
	dubbo.buffer.Write(dubbo.databuf.Bytes())
	return dubbo.buffer.Bytes()
}

func (dubbo *Dubbo) attach(bs []byte) {
	if bs == nil {
		dubbo.databuf.WriteString("null\n")
	} else {
		dubbo.databuf.WriteString("\"")
		dubbo.databuf.Write(bs)
		dubbo.databuf.WriteString("\"\n")
	}
}

func (dubbo *Dubbo) attachJSON(key []byte, val []byte) {
	dubbo.databuf.WriteString("{\"")
	dubbo.databuf.Write(key)
	dubbo.databuf.WriteString("\":")
	dubbo.databuf.WriteString("\"")
	dubbo.databuf.Write(val)
	dubbo.databuf.WriteString("\"}\n")
	//"{\"" + key + "\":" + "\"" + val + "\"}\n"
}

func (dubbo *Dubbo) makeHeader(dataLen uint32, requestID uint64) {

	// header length.
	//FLAG_EVENT := byte(0x20)

	// header.
	dubbo.buffer.Reset()
	// set magic number
	dubbo.header[0] = magicHigh
	dubbo.header[1] = magicLow
	// set request and serilization bit
	dubbo.header[2] = byte(flagRequest | flagTwoWay | serializationID)
	// set request id
	binary.BigEndian.PutUint64(dubbo.header[4:12], requestID)

	// encode request data.
	binary.BigEndian.PutUint32(dubbo.header[12:16], dataLen)
	dubbo.buffer.Write(dubbo.header)

}
