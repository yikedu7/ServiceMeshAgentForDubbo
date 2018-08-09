package registry

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/coreos/etcd/clientv3"
)

// IRegistry ...
type IRegistry interface {
	Register(serviceName string, port int) error
	Find(serviceName string) error
}

const (
	etcdRootPath = "dubbomesh"
)

// Property ...
type Property struct {
	ServiceType string
	ServiceName string
	Port        string
	DestPort    string
	Power       string
}

// EndPoint ...
type EndPoint struct {
	Host  string
	Port  string
	Power int
}

// EtcdRegistry ...
type EtcdRegistry struct {
	logger log.Logger
	client *clientv3.Client
	key    string
}

// GetHost ...
func GetHost() string {
	name, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	addrs, err := net.LookupHost(name)
	if err != nil {
		panic(err)
	}
	return addrs[0]

	//return "127.0.0.1"
}

// Init ...
func (etcdReg *EtcdRegistry) Init(cli *clientv3.Client, property Property) {
	etcdReg.client = cli
	etcdReg.key = ""
	go etcdReg.KeepAlive()

	if property.ServiceType == "provider" {
		etcdReg.Register(property.ServiceName, property.Port, property.Power)
	}
}

// Register ...
func (etcdReg *EtcdRegistry) Register(serviceName string, port string, power string) {
	etcdReg.key = fmt.Sprintf("%s/%s/%s:%s", etcdRootPath, serviceName, GetHost(), port)
	_, err := etcdReg.client.Put(context.Background(), etcdReg.key, power)
	if err != nil {
		panic(err)
	}

	fmt.Println("Register a new service at:" + etcdReg.key)
}

// KeepAlive ...
func (etcdReg *EtcdRegistry) KeepAlive() {
	rch := etcdReg.client.Watch(context.Background(), etcdReg.key, clientv3.WithPrefix())
	for resp := range rch {
		for _, ev := range resp.Events {
			fmt.Printf("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
		}
	}
}

// Find ...
func (etcdReg *EtcdRegistry) Find(serviceName string) []EndPoint {
	keystr := fmt.Sprintf("%s/%s", etcdRootPath, serviceName)
	fmt.Println(keystr)
	key := keystr
	resp, err := etcdReg.client.Get(context.Background(), key, clientv3.WithPrefix())
	if err != nil {
		panic(err)
	}

	endpoints := make([]EndPoint, 0)

	for _, kv := range resp.Kvs {
		s := string(kv.Key[:])
		power, err := strconv.Atoi(string(kv.Value[:]))
		if err != nil {
			panic(err)
		}
		index := strings.LastIndex(s, "/")
		endpointStr := s[index+1:]
		hostp := strings.Split(endpointStr, ":")
		endpoints = append(endpoints, EndPoint{hostp[0], hostp[1], power})

	}
	return endpoints
}
