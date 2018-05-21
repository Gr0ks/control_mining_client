package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
)

type MinerStatus struct {
	Version string `json:"version"`
	Time    int64  `json:"time"`

	EHashrate      int64   `json:"e_hashrate"`
	EValidShares   int64   `json:"e_valid_shares"`
	EInvalidShares int64   `json:"e_invalid_shares"`
	EGpuHashrate   []int64 `json:"e_gpu_hashrate"`

	DHashrate      int64   `json:"d_hashrate"`
	DValidShares   int64   `json:"d_valid_shares"`
	DInvalidShares int64   `json:"d_invalid_shares"`
	DGpuHashrate   []int64 `json:"d_gpu_hashrate"`

	Pools string    `json:"pools"`
	Gpus  [][]int64 `json:"gpus"`
}

type Miner struct {
	Name   string       `json:"name"`
	Addr   string       `json:"addr"`
	Port   int64        `json:"port"`
	Status *MinerStatus `json:"status"`
	Online bool         `json:"online"`
}

func (miner *Miner) GetStatus() {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", miner.Addr, miner.Port))
	if err != nil {
		if miner.Online {
			log.Printf("%s %s disconnected", miner.Name, miner.Addr)
		}
		miner.Online = false
		return
	}
	if !miner.Online {
		log.Printf("%s %s connected", miner.Name, miner.Addr)
		miner.Online = true
	}

	text := `{"id":0,"jsonrpc":"2.0","method":"miner_getstat1"}`
	fmt.Fprintf(conn, text+"\n")
	message, _ := bufio.NewReader(conn).ReadString('\n')
	if len(message) > 0 {
		data := make(map[string]interface{})
		json.Unmarshal([]byte(message), &data)
		result := data["result"].([]interface{})
		stat := MinerStatus{}
		stat.Version = result[0].(string)
		stat.Time, _ = strconv.ParseInt(result[1].(string), 10, 64)

		p := getInts(result[2].(string))
		stat.EHashrate = p[0]
		stat.EValidShares = p[1]
		stat.EInvalidShares = p[2]
		stat.EGpuHashrate = getInts(result[3].(string))

		p = getInts(result[4].(string))
		stat.DHashrate = p[0]
		stat.DValidShares = p[1]
		stat.DInvalidShares = p[2]
		stat.DGpuHashrate = getInts(result[5].(string))

		temps := getInts(result[6].(string))
		var gpus [][]int64
		ov := int64(0)
		for i, t := range temps {
			if i%2 != 0 {
				var arr []int64
				arr = append(arr, ov)
				arr = append(arr, t)
				gpus = append(gpus, arr)
			} else {
				ov = t
			}
		}

		stat.Gpus = gpus
		stat.Pools = result[7].(string)

		miner.Status = &stat
	}
	conn.Close()

}
