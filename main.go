package main

import (
	"fmt"
	_ "net/http/pprof"
	"os"
	"strconv"
	"time"

	"code.aliyun.com/runningguys/agent/controller"
	"code.aliyun.com/runningguys/agent/registry"
	"github.com/coreos/etcd/clientv3"
)

var (
	dialTimeout    = 5 * time.Second
	requestTimeout = 2 * time.Second
	endpoints      = []string{"172.17.0.1:2379"}
	serviceName    = "com.alibaba.dubbo.performance.demo.provider.IHelloService"
)

func main() {
	if len(os.Args) != 6 {
		fmt.Printf("usage: %s <service type> <srcPort> <destPort> <etcd-url> <power>\n", os.Args[0])
		return
	}
	serType := os.Args[1]
	srcPort := os.Args[2]
	destPort := os.Args[3]
	endpoints[0] = os.Args[4]
	power := os.Args[5]
	_, err := strconv.Atoi(power)
	if err != nil {
		fmt.Printf("invalid power")
		return
	}
	if serType != "provider" && serType != "consumer" {
		fmt.Printf("invalid service type")
		return
	}

	if _, err := strconv.Atoi(srcPort); err != nil {
		fmt.Printf("invalid srcPort")
		return
	}

	if _, err := strconv.Atoi(destPort); err != nil {
		fmt.Printf("invalid destPort")
	}
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		panic(err)
	}
	defer cli.Close()
	property := registry.Property{
		serType,
		serviceName,
		srcPort,
		destPort,
		power,
	}
	/*go func() {
		log.Println(http.ListenAndServe(":6060", nil))
	}()*/
	controller.StartAgent(property, cli)

}
