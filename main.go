package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type Config struct {
	Miner  map[string]map[string]*Miner `json:"miner"`
	Listen string                       `json:"listen"`
}

type Server struct {
	Config *Config
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

func setHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-cache")
}

func (srv *Server) MinersHandler(w http.ResponseWriter, r *http.Request) {
	setHeaders(w)
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(srv.Config.Miner)
	if err != nil {
		log.Println("Error serializing /miner: ", err)
	}
}

func main() {
	cfg, err := LoadConfigFile()
	if err != nil {
		log.Panic(err)
		return
	}

	server := Server{}
	server.Config = cfg

	r := mux.NewRouter()

	r.HandleFunc("/miner", server.MinersHandler)

	go func() {
		l, err := net.Listen("tcp", server.Config.Listen)
		if err != nil {
			log.Panic(err)
		} else {
			if err := http.Serve(l, r); err != nil {
				log.Panic(err)
			}
		}
	}()

	if err != nil {
		log.Panic(err)
		return
	}

	intv := time.Duration(time.Second * 10)
	timer := time.NewTimer(intv)
	for _, miner := range cfg.Miner {
		for _, worker := range miner {
			go worker.GetStatus()
		}
	}
	for {
		select {
		case <-timer.C:
			for _, miner := range cfg.Miner {
				for _, worker := range miner {
					go worker.GetStatus()
				}
			}
			timer.Reset(intv)
		}
	}
}
