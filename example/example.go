package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	connectpool "github.com/HuXin0817/ConnectPool" // Import the external package for connection pooling.
)

const PoolSize = 1000 // Size of the connection pool.

var (
	connectId atomic.Int64 // Counter for the number of connections made.
	closeId   atomic.Int64 // Counter for the number of connections closed.

	t atomic.Int64 // General purpose atomic counter.
)

var (
	r = rand.New(rand.NewSource(time.Now().UnixNano())) // Initialize a random number generator.
)

// connectMethod defines how to establish a new connection to Redis.
func connectMethod() any {
	connectId.Add(1)        // Increment the connection ID counter.
	return connectId.Load() // Return the Redis client.
}

// closeMethod defines how to close a Redis connection.
func closeMethod(db any) {
	closeId.Add(1) // Increment the close ID counter.
}

var pool = connectpool.NewConnectPool(connectMethod, connectpool.WithCap(PoolSize), connectpool.WithCloseMethod(closeMethod)) // Initialize the connection pool with the connect method.

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

	const turn = PoolSize * 5 // Define how many goroutines to spawn.

	var wq sync.WaitGroup
	wq.Add(turn) // Set the wait group counter.

	for i := 0; i < turn; i++ { // Iterate and spawn goroutines.
		go func() {
			defer wq.Done() // Signal done on goroutine completion.

			time.Sleep(time.Second * time.Duration(r.Int63()%5)) // Random sleep to simulate work.

			c, cancel := pool.RegisterWithTimeLimit(time.Second * time.Duration(r.Int63()%5)) // Register for a connection from the pool.
			defer cancel()                                                                    // Ensure the connection is released.

			_ = c

			t.Add(1) // Increment the general counter.

			time.Sleep(time.Second * time.Duration(r.Int63()%5)) // Random sleep to simulate more work.
		}()
	}

	wq.Wait() // Wait for all goroutines to complete.

	for pool.Size() > 0 {
		// Wait for the pool to empty.
		runtime.Gosched()
	}

	printInfo() // Print final pool information.
}
