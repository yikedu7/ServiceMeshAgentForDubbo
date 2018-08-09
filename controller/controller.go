package controller

import (
	"fmt"
	"math/rand"
	"net"
	"sync"

	"code.aliyun.com/runningguys/agent/registry"
	"github.com/coreos/etcd/clientv3"
)

var interfaceName = []byte("com.alibaba.dubbo.performance.demo.provider.IHelloService")

var rid uint64

var once sync.Once
var maxQueue = 200

type myconn struct {
	con net.Conn
	//full      chan struct{}
	power int
	pchan chan []byte

	active int
}

// StartAgent ...
func StartAgent(property registry.Property, cli *clientv3.Client) {
	if property.ServiceType == "provider" {
		startProviderAgent(property, cli)
	} else if property.ServiceType == "consumer" {
		startConsumerAgent(property, cli)
	} else {
		fmt.Println("invalid service type")
	}
}

/*func (ca *consumerAgent) sel() int {
	r := rand.Intn(ca.sum)
	n := len(ca.endpoints)
	down := 0
	up := 0
	for i := 0; i < n; i++ {
		up += ca.endpoints[i].Power
		if down <= r && r < up {
			return i
		}
		down = up
	}
	return n - 1
}*/

func (ca *consumerAgent) randSel() int {
	r := rand.Intn(ca.sum)
	for i := 0; i < ca.length; i++ {
		r -= ca.endpoints[i].Power
		if r < 0 {
			return i
		}
	}
	return ca.length - 1
}
func (ca *consumerAgent) randomSel() int {
	return rand.Intn(ca.length)
}

func (ca *consumerAgent) roundRobin(id uint64) int {
	return int(id % uint64(ca.length))
}

/*func (ca *consumerAgent) writeDatas(buf []byte) {
	for {
		index := ca.sel()
		select {
		case ca.pcons[index].pchan <- buf:
			return
		default:
		}
	}
}*/

func readn(conn net.Conn, buf []byte, l uint32) error {
	len := uint32(0)
	for {
		n, err := conn.Read(buf[len:])
		if n > 0 {
			len += uint32(n)
			if l <= len {
				break
			}
		}
		if err != nil {

			fmt.Printf("读错误\n")
			return err

		}
	}
	return nil
}
