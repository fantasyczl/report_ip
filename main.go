package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/go-redis/redis/v8"
	"gopkg.in/yaml.v2"
)

var g_conf *Conf

const CONF_FILE = ".report_ip/config.yaml"

func main() {
	var isServer = flag.Bool("server", false, "run app as server")
	flag.Parse()

	confPath := getConfPath()
	var err error
	if g_conf, err = parseConf(confPath); err != nil {
		log.Fatalf("parse conf failed, %s", err)
	}

	if *isServer {
		serverMode()
	} else {
		clientMode()
	}
}

func getConfPath() string {
	home := os.Getenv("HOME")
	return filepath.Join(home, CONF_FILE)
}

func serverMode() {
	log.Println("server mode")

	ip, err := getLocalIP()
	if err != nil {
		log.Printf("getLocalIP failed, err:%s", err)
		return
	}

	log.Printf("Local IP:%s", ip.String())

	// set ip to content
	reportIP(g_conf, ip)
}

func clientMode() {
	fmt.Println("client mode, read ip from redis")

	ip, err := readIP(g_conf)
	if err != nil {
		log.Printf("read ip failed, err:%s\n", err)
	} else {
		log.Printf("Ret: %s\n", ip.String())
	}
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

type ReportIPRet struct {
	ReportTime int64  `json:"reportTime"`
	IP         string `json:"IP"`
}

func (r *ReportIPRet) String() string {
	t := time.Unix(r.ReportTime, 0).In(time.Local)

	return fmt.Sprintf("Time: %s\t\t\tIP: %s", t, r.IP)
}

func reportIP(conf *Conf, ip net.IP) {
	ret := ReportIPRet{
		ReportTime: time.Now().Unix(),
		IP:         ip.String(),
	}

	bs, err := json.Marshal(ret)
	if err != nil {
		log.Printf("json marshal failed")
		return
	}

	redisClient = getRedisClient(conf)

	var ctx = context.Background()
	if err = redisClient.Set(ctx, conf.IPKey, string(bs), 0).Err(); err != nil {
		log.Printf("redis set failed, %s\n", err)
		return
	}

	log.Printf("set to redis successfully")
}

func readIP(conf *Conf) (*ReportIPRet, error) {
	redisClient = getRedisClient(conf)

	var ctx = context.Background()
	ret, err := redisClient.Get(ctx, conf.IPKey).Result()
	if err != nil {
		log.Printf("redis get failed, %s\n", err)
		return nil, err
	}

	ipRet := ReportIPRet{}
	if err := json.Unmarshal([]byte(ret), &ipRet); err != nil {
		return nil, err
	}

	return &ipRet, nil
}
