package mesh

import (
	"bytes"
	"encoding/binary"
)

/*
	协议格式：
	byte 0-3 : 字段长度
	byte 4-11 : httpID
	字段1\n
	字段2\n
	...
*/

// Encode ...
func Encode(httpID uint64, interfaceName string, method string, version string, paramTyps string, params []string) *bytes.Buffer {
	interfaceName = checkNullString(interfaceName)
	method = checkNullString(method)
	version = checkNullString(version)
	paramTyps = checkNullString(paramTyps)
	datas := interfaceName + "\n" + method + "\n" + version + "\n" + paramTyps + "\n"
	for _, s := range params {
		datas += s + "\n"
	}
	len := uint32(len(datas))
	bs := make([]byte, 4)
	idbs := make([]byte, 8)
	binary.BigEndian.PutUint32(bs, len)
	binary.BigEndian.PutUint64(idbs, httpID)
	buff := bytes.NewBuffer(bs)
	buff.Write(idbs)
	_, err := buff.WriteString(datas)
	if err != nil {
		panic(err)
	}
	//fmt.Printf("%s", buff.Bytes())
	return buff
}

func checkNullString(str string) string {
	if str == "" {
		return "null"
	}
	return str
}
