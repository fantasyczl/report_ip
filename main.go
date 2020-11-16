package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"

	"github.com/go-redis/redis/v8"
	"gopkg.in/yaml.v2"
)

var g_conf *Conf

const CONF_FILE = "./conf/config.yaml"

func main() {
	fmt.Println("vim-go")
	var err error

	if g_conf, err = parseConf(CONF_FILE); err != nil {
		log.Fatalf("parse conf failed, %s", err)
	}

	ip, err := getLocalIP()
	if err != nil {
		log.Printf("getLocalIP failed, err:%s", err)
		return
	}

	log.Printf("Local IP:%s", ip.String())

	// set ip to content
	reportIP(g_conf, ip)
}

// conf
type RedisConf struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

func (r *RedisConf) Address() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type Conf struct {
	Redis *RedisConf `yaml:"redis"`
	IPKey string     `yaml:"ip_key"`
}

func parseConf(filename string) (*Conf, error) {
	if ex, err := isExist(filename); err != nil {
		return nil, err
	} else if !ex {
		return nil, errors.New("filepath does not exist")
	}

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	conf := new(Conf)
	err = yaml.Unmarshal(bytes, &conf)
	return conf, err
}

func isExist(filepath string) (bool, error) {
	if filepath == "" {
		return false, nil
	}

	fileinfo, err := os.Stat(filepath)
	if err != nil {
		return false, err
	}

	if fileinfo.IsDir() {
		return false, nil
	}

	return true, nil
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

var redisClient *redis.Client

func getRedisClient(conf *Conf) *redis.Client {
	if redisClient != nil {
		return redisClient
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     conf.Redis.Address(),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return redisClient
}

func reportIP(conf *Conf, ip net.IP) {
	redisClient = getRedisClient(conf)

	var ctx = context.Background()
	err := redisClient.Set(ctx, conf.IPKey, ip.String(), 0).Err()
	if err != nil {
		log.Printf("redis set failed, %s\n", err)
		return
	}

	log.Printf("set to redis successfully")
}
