package main

import (
	"fmt"
	"log"
	"net"
)

var conf *Conf

func main() {
	fmt.Println("vim-go")

	parseConf()

	ip, err := getLocalIP()
	if err != nil {
		log.Printf("getLocalIP failed, err:%s", err)
		return
	}

	log.Printf("Local IP:%s", ip.String())
}

// conf
type RedisConf struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type Conf struct {
	Redis *RedisConf `yaml:"redis"`
}

func parseConf() (*Conf, error) {
	conf := new(Conf)

	// TODO

	return conf, nil
}

func getLocalIP() (ip net.IP, err error) {
	var ifaces []net.Interface
	ifaces, err = net.Interfaces()
	if err != nil {
		return
	}

	for _, i := range ifaces {
		var addrs []net.Addr
		addrs, err = i.Addrs()
		if err != nil {
			return
		}

		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip.IsLoopback() {
				continue
			}

			if ip.To4() == nil {
				continue
			}

			return
		}
	}

	return
}

func reportIP() {
}
