package controller

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/valyala/fasthttp"

	"code.aliyun.com/runningguys/agent/dubbo"
	"code.aliyun.com/runningguys/agent/registry"
	"github.com/coreos/etcd/clientv3"
)

var weight = []int{14, 25, 25}
var delay = time.Millisecond * 1
var mu sync.Mutex
var maxconn = uint64(1024)
var dubboVersion = []byte("2.6.1")
var sep = []byte("\n")
var maxdatalen = 20

type consumerAgent struct {
	port      string
	reg       *registry.EtcdRegistry
	endpoints []registry.EndPoint
	pcons     []myconn
	resp      []chan []byte
	sum       int
	length    int
	reqChan   chan []byte
	done      chan *myconn
	dubboBuf  []*dubbo.Dubbo
	resbuf    [][]byte
}

func startConsumerAgent(property registry.Property, cli *clientv3.Client) {
	reg := new(registry.EtcdRegistry)
	reg.Init(cli, property)
	ca := &consumerAgent{property.Port, reg, nil, nil, nil, 0, 0, nil, nil, nil, nil}
	ca.endpoints = ca.reg.Find(string(interfaceName))
	ca.length = len(ca.endpoints)
	ca.reqChan = make(chan []byte, ca.length*maxQueue)
	ca.done = make(chan *myconn, 1)
	ca.dubboBuf = make([]*dubbo.Dubbo, maxconn)
	ca.resbuf = make([][]byte, maxconn)
	for i := 0; i < ca.length; i++ {
		ca.endpoints[i].Power = weight[ca.endpoints[i].Power-1]
		ca.sum += ca.endpoints[i].Power
	}
	ca.pcons = make([]myconn, ca.length)
	ca.resp = make([]chan []byte, maxconn)
	defer ca.closePcon()
	ca.listenAndServe()
}

func (ca *consumerAgent) Connect() {
	for i := 0; i < len(ca.endpoints); i++ {
		ep := ca.endpoints[i]
		pcon, err := net.Dial("tcp", ep.Host+":"+ep.Port)
		if err != nil {
			panic(err)
		}
		mcon := myconn{pcon, ep.Power, make(chan []byte, maxQueue), 0}
		// 5: 2886, 6: 2940
		ca.pcons[i] = mcon
		go ca.readPcon(&ca.pcons[i])
		go ca.writePcon(&ca.pcons[i])
		fmt.Println("create conn")
	}
	go ca.balance()
}

func (ca *consumerAgent) writePcon(mcon *myconn) {
	pcon := mcon.con
	for {
		buf := <-mcon.pchan
		_, err := pcon.Write(buf)
		if err != nil {
			fmt.Println("writePcon error")
			break
		}
	}
}
func (ca *consumerAgent) readPcon(mcon *myconn) {
	pcon := mcon.con
	header := make([]byte, 16)
	//mdata := make([]byte, 20)

	for {

		err := readn(pcon, header, 16)
		if err != nil {
			fmt.Println("readResponse error")
			break
			//return err
		}
		dataLen := binary.BigEndian.Uint32(header[12:16])
		id := binary.BigEndian.Uint64(header[4:12])

		ch := ca.resp[id]
		data := ca.resbuf[id]
		if ch != nil && data != nil {
			data = data[:dataLen]
			err = readn(pcon, data, dataLen)
			if err != nil {
				fmt.Println("readResponse error")
				//return err
				break
			}

			resb := bytes.Split(data, sep)[1]
			//resbType := string(resbs[0][:])
			ch <- resb
			ca.done <- mcon
		} else {
			fmt.Println("id 不存在")
		}
	}
}

func (ca *consumerAgent) balance() {
	for {
		select {
		case buf := <-ca.reqChan:
			index := ca.leastActive()
			ca.pcons[index].pchan <- buf
			ca.pcons[index].active++
		case pcon := <-ca.done:
			pcon.active--
		}
	}
}
func (ca *consumerAgent) closePcon() {
	for _, con := range ca.pcons {
		con.con.Close()
	}
}
func (ca *consumerAgent) httpHandle(ctx *fasthttp.RequestCtx) {
	once.Do(ca.Connect)
	// := (ctx.ConnID() << 32) ^ ctx.ID()
	id := ctx.ConnID() % maxconn
	if ca.resp[id] == nil {
		ca.resp[id] = make(chan []byte)
	}
	if ca.dubboBuf[id] == nil {
		ca.dubboBuf[id] = dubbo.NewDubbo(dubboVersion)
	}
	if ca.resbuf[id] == nil {
		ca.resbuf[id] = make([]byte, maxdatalen)
	}
	postVal := ctx.PostArgs()
	method := postVal.Peek("method")
	paramTypes := postVal.Peek("parameterTypesString")
	params := postVal.Peek("parameter")
	ca.reqChan <- ca.dubboBuf[id].Encode(id, interfaceName, nil, method, paramTypes, [][]byte{params})
	ctx.SetStatusCode(200)
	ctx.SetContentType("application/json")
	ctx.Write(<-ca.resp[id])
}

func (ca *consumerAgent) sel() int {
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
}

/*func (ca *consumerAgent) selFromID(id uint64) int {
	r := int(id % uint64(ca.sum))
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
}
*/

func (ca *consumerAgent) leastActive() int {
	index := 0
	min := ca.pcons[0].active / ca.pcons[0].power
	for i := 1; i < ca.length; i++ {
		active := ca.pcons[i].active / ca.pcons[i].power
		if active < min {
			min = active
			index = i
		}
	}
	/*power := ca.pcons[index].power
	for i := 0; i < ca.length; i++ {
		if ca.pcons[i].active == min {
			if power < ca.pcons[i].power {
				power = ca.pcons[i].power
				index = i
			}
		}
	}*/
	return index
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

func (ca *consumerAgent) listenAndServe() {

	if err := fasthttp.ListenAndServe("0.0.0.0:"+ca.port, ca.httpHandle); err != nil {
		fmt.Println("start fasthttp fail:", err.Error())
	}
}
