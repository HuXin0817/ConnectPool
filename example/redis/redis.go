package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	connectpool "github.com/HuXin0817/ConnectPool"
	"github.com/go-redis/redis"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Redis Redis `yaml:"redis"`
}

type Redis struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

var DBConfig Config

const ConfigFilePath = "./config.yaml"

func init() {
	data, err := ioutil.ReadFile(ConfigFilePath)
	if err != nil {
		log.Panicf("error: %v", err)
	}

	err = yaml.Unmarshal(data, &DBConfig)
	if err != nil {
		log.Panicf("error: %v", err)
	}
}

const PoolSize = 1000

var (
	connectId atomic.Int64
	closeId   atomic.Int64

	t atomic.Int64
)

var (
	port = DBConfig.Redis.Port
	host = DBConfig.Redis.Host

	uri = fmt.Sprintf("%s:%d", host, port)

	r = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func connectMethod() any {
	connectId.Add(1)

	rdb := redis.NewClient(&redis.Options{
		Addr:     uri,
		Password: "",
		DB:       0,
	})

	if rdb == nil {
		log.Panicf("connect error! ConnectId: %d\n", connectId.Load())
	}

	return rdb
}

func closeMethod(db any) {
	closeId.Add(1)

	err := db.(*redis.Client).Close()
	if err != nil {
		panic(err)
	}
}

var pool = connectpool.NewConnectPool(PoolSize, connectMethod)

func printInfo() {
	fmt.Printf("WorkingNumber: %d\n", pool.WorkingNumber())
	fmt.Printf("ConnectId: %d\n", connectId.Load())
	fmt.Printf("CloseId: %d\n", closeId.Load())
	fmt.Printf("RegisterCount: %d\n\n", t.Load())

	time.Sleep(time.Second / 2)
}

func main() {
	go func() {
		for {
			printInfo()
		}
	}()

	pool.SetCloseMethod(closeMethod)

	const turn = PoolSize * 5

	var wq sync.WaitGroup
	wq.Add(turn)

	for range turn {
		go func() {
			defer wq.Done()

			time.Sleep(time.Second * time.Duration(r.Int63()%5))

			c, cancel := pool.Register()
			defer cancel()

			rdb := c.(*redis.Client)
			rdb.Ping()

			t.Add(1)

			time.Sleep(time.Second * time.Duration(r.Int63()%5))
		}()
	}

	wq.Wait()

	for pool.WorkingNumber() > 0 {
		runtime.Gosched()
	}

	for pool.PoolSize() > 0 {
		runtime.Gosched()
	}

	printInfo()
}
