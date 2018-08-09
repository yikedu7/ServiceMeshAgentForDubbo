package controller

import (
	"fmt"
	"io"
	"net"

	"code.aliyun.com/runningguys/agent/registry"
	"github.com/coreos/etcd/clientv3"
)

type providerAgent struct {
	port     string
	destPort string
}

func startProviderAgent(property registry.Property, cli *clientv3.Client) {
	reg := new(registry.EtcdRegistry)
	reg.Init(cli, property)
	pa := &providerAgent{property.Port, property.DestPort}
	pa.listenAndServe()
}

func (pa *providerAgent) listenAndServe() {
	ln, err := net.Listen("tcp", ":"+pa.port)
	if err != nil {
		panic(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}
		go pa.serve(conn)
	}
}

func (pa *providerAgent) serve(conn net.Conn) {
	defer conn.Close()

	pcon, err := net.Dial("tcp", "127.0.0.1:"+pa.destPort)
	if err != nil {
		panic(err)
	}
	go copy(conn, pcon)

	_, err = io.Copy(pcon, conn)
	if err != nil {
		fmt.Println(err)
	}

}
func copy(conn net.Conn, pcon net.Conn) {
	defer pcon.Close()

	_, err := io.Copy(conn, pcon)

	if err != nil {
		fmt.Println(err)
	}
}
