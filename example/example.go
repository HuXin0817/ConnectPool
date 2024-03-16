package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	connectpool "github.com/HuXin0817/ConnectPool" // Import the external package for connection pooling.
	"github.com/go-redis/redis"                    // Import the Redis client package.
	"gopkg.in/yaml.v3"                             // Import the YAML package for configuration parsing.
)

// Config struct to map the configuration file structure.
type Config struct {
	Redis Redis `yaml:"redis"` // Redis configuration section.
}

// Redis struct to hold Redis-specific configurations.
type Redis struct {
	Host string `yaml:"host"` // Redis server host.
	Port int    `yaml:"port"` // Redis server port.
}

var DBConfig Config // Global variable to hold database configuration.

const ConfigFilePath = "./config.yaml" // Path to the configuration file.

// init function runs before main to load and parse the configuration file.
func init() {
	data, err := ioutil.ReadFile(ConfigFilePath) // Read the configuration file.
	if err != nil {
		log.Panicf("error: %v", err) // Log and panic on read error.
	}

	err = yaml.Unmarshal(data, &DBConfig) // Parse the YAML into the Config struct.
	if err != nil {
		log.Panicf("error: %v", err) // Log and panic on parsing error.
	}
}

const PoolSize = 1000 // Size of the connection pool.

var (
	connectId atomic.Int64 // Counter for the number of connections made.
	closeId   atomic.Int64 // Counter for the number of connections closed.

	t atomic.Int64 // General purpose atomic counter.
)

var (
	port = DBConfig.Redis.Port // Redis port from config.
	host = DBConfig.Redis.Host // Redis host from config.

	uri = fmt.Sprintf("%s:%d", host, port) // Construct the Redis URI.

	r = rand.New(rand.NewSource(time.Now().UnixNano())) // Initialize a random number generator.
)

// connectMethod defines how to establish a new connection to Redis.
func connectMethod() any {
	connectId.Add(1) // Increment the connection ID counter.

	rdb := redis.NewClient(&redis.Options{
		Addr:     uri,
		Password: "", // Assuming no password is set.
		DB:       0,  // Default database.
	})

	if rdb == nil {
		log.Panicf("connect error! ConnectId: %d\n", connectId.Load()) // Panic if connection fails.
	}

	return rdb // Return the Redis client.
}

// closeMethod defines how to close a Redis connection.
func closeMethod(db any) {
	closeId.Add(1) // Increment the close ID counter.

	err := db.(*redis.Client).Close() // Close the Redis connection.
	if err != nil {
		panic(err) // Panic on close error.
	}
}

var pool = connectpool.NewConnectPool(PoolSize, connectMethod) // Initialize the connection pool with the connect method.

// printInfo prints out connection pool statistics.
func printInfo() {
	fmt.Printf("WorkingNumber: %d\n", pool.WorkingNumber()) // Number of connections currently in use.
	fmt.Printf("ConnectId: %d\n", connectId.Load())         // Total number of connections established.
	fmt.Printf("CloseId: %d\n", closeId.Load())             // Total number of connections closed.
	fmt.Printf("RegisterCount: %d\n\n", t.Load())           // Total number of successful registrations.

	time.Sleep(time.Second / 2) // Pause for half a second.
}

func main() {
	go func() {
		for {
			printInfo() // Continuously print pool information.
		}
	}()

	pool.SetCloseMethod(closeMethod) // Set the method to close a connection in the pool.

	const turn = PoolSize * 5 // Define how many goroutines to spawn.

	var wq sync.WaitGroup
	wq.Add(turn) // Set the wait group counter.

	for i := 0; i < turn; i++ { // Iterate and spawn goroutines.
		go func() {
			defer wq.Done() // Signal done on goroutine completion.

			time.Sleep(time.Second * time.Duration(r.Int63()%5)) // Random sleep to simulate work.

			c, cancel := pool.Register() // Register for a connection from the pool.
			defer cancel()               // Ensure the connection is released.

			rdb := c.(*redis.Client) // Type assert to a Redis client.
			rdb.Ping()               // Ping the Redis server.

			t.Add(1) // Increment the general counter.

			time.Sleep(time.Second * time.Duration(r.Int63()%5)) // Random sleep to simulate more work.
		}()
	}

	wq.Wait() // Wait for all goroutines to complete.

	for pool.PoolSize() > 0 {
		// Wait for the pool to empty.
	}

	printInfo() // Print final pool information.
}
