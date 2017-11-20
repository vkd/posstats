package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"
)

var (
	addr      = flag.String("addr", ":8080", "Address of server")
	redisAddr = flag.String("redis-addr", ":6378", "Address of redis server")

	configPath = flag.String("config", "config.js", "Path to config file")
)

// Config - config of service
type Config struct {
	Addr      string `json:"addr"`
	RedisAddr string `json:"redis-addr"`

	SeriesTimeoutSec int `json:"seriesTimeoutSec"`
	ResetTimeoutSec  int `json:"resetTimeoutSec"`
}

func main() {
	flag.Parse()

	var c = Config{
		Addr:      *addr,
		RedisAddr: *redisAddr,

		SeriesTimeoutSec: 5,
		ResetTimeoutSec:  600,
	}
	if *configPath != "" {
		f, err := os.OpenFile(*configPath, os.O_RDONLY, 0)
		if err != nil {
			fmt.Printf("Error on load config (%s): %v\n", *configPath, err)
			os.Exit(1)
		}
		defer f.Close() // TODO check error

		err = json.NewDecoder(f).Decode(&c)
		if err != nil {
			fmt.Printf("Error on unmarshal config (%s): %v", *configPath, err)
			os.Exit(1)
		}
	}

	redis, err := NewRedisClient(c.RedisAddr, time.Duration(c.SeriesTimeoutSec)*time.Second, time.Duration(c.ResetTimeoutSec)*time.Second)
	if err != nil {
		fmt.Printf("Error on connect to redis: %v\n", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", PosHandler(redis, redis))
	mux.HandleFunc("/stats", StatsHandler(redis))

	s := http.Server{
		Addr: c.Addr,
		// TODO set timeout
		// TODO graceful shutdown
		Handler: mux,
	}

	fmt.Printf("Starting server on %s ...", s.Addr)
	err = s.ListenAndServe()
	fmt.Printf("Server has stopped: %v", err)
}
