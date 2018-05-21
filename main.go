package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Miner  map[string]map[string]*Miner `json:"miner"`
	Listen string                       `json:"listen"`
}

func getInts(s string) []int64 {
	parts := strings.Split(s, ";")
	var arr []int64
	for _, p := range parts {
		i, _ := strconv.ParseInt(p, 10, 64)
		arr = append(arr, i)
	}
	return arr
}

func LoadConfigFile() (*Config, error) {
	file, err := os.Open("config.json")
	if err != nil {
		return nil, err
	}
	cfg := Config{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

const (
	tcpProtocol = "tcp"
)

var connectAddr = &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8989}

func sendKey(cfg *Config) {
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:8989"))
	if err != nil {
		log.Panic(err)
		return
	}
	defer func() {
		fmt.Println("Closing connection...")
		conn.Close()
	}()
	log.Println("Connect to ", conn.RemoteAddr)
	message, err := json.Marshal(cfg)
	if err != nil {
		log.Panic(err)
		return
	}
	text := string(message)
	fmt.Fprintf(conn, text+"\n")

}

func main() {
	cfg, err := LoadConfigFile()
	if err != nil {
		log.Panic(err)
		return
	}
	//log.Println(cfg)
	//cfg.Miner содержит инфу о майнере
	//cfg.Server - сервер сборщик статистики

	stop := make(chan bool)
	intv := time.Duration(time.Second * 10)
	timer := time.NewTimer(intv)
	var miner map[string]*Miner
	var worker *Miner
	for _, miner = range cfg.Miner {
		for _, worker = range miner {
			go worker.GetStatus()

		}
	}

	for {
		select {
		case <-timer.C:
			for _, miner = range cfg.Miner {
				for _, worker = range miner {
					go worker.GetStatus()
					sendKey(cfg)
					log.Println("Send: ", worker.Status)

				}
			}
			timer.Reset(intv)
		}
	}

	stop <- true
}
