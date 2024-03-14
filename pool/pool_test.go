package pool

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var printLock sync.Mutex // Lock used to synchronize print operations to prevent interleaved log outputs

var cnt atomic.Int32 // Global counter used to assign unique identifiers to connections

// mockConnectMethod simulates a connection creation method, incrementing and logging the counter.
func mockConnectMethod() any {
	printLock.Lock()
	defer printLock.Unlock()

	c := cnt.Add(1) // Increment the global counter

	log.Printf("new connect %v\n", c) // Log the creation of a new connection
	return c                           // Return the incremented counter as the connection identifier
}

// mockCloseFunc simulates a connection close function, logging the identifier of the connection being closed.
func mockCloseFunc(connect any) {
	printLock.Lock()
	defer printLock.Unlock()

	log.Printf("connect %v close", connect) // Log the closure of a connection
}

// mockDealPanicMethod simulates a panic handling function, logging the panic information.
func mockDealPanicMethod(panicInfo any) {
	printLock.Lock()
	defer printLock.Unlock()

	log.Printf("panic: %v\n", panicInfo) // Log the panic information
}

// TestPool tests the connection pool's functionality, particularly the registration and deregistration of connections.
func TestPool(t *testing.T) {
	const Turn = 1000000 // Define the number of iterations (and thus goroutines) to simulate

	pool := NewConnectPool(Turn/2, mockConnectMethod) // Initialize a new connection pool with half of Turn as its size
	pool.SetCloseMethod(mockCloseFunc)                // Set the method to be called when a connection is closed
	pool.SetDealPanicMethod(mockDealPanicMethod)      // Set the method to handle panic situations

	var wq sync.WaitGroup // Use a WaitGroup to wait for all goroutines to complete
	wq.Add(Turn)          // Add the total number of turns to the WaitGroup counter

	for range Turn {
		go func() {
			defer wq.Done() // Decrement the WaitGroup counter when the goroutine completes

			_, cancel := pool.Register() // Register a new connection in the pool
			time.Sleep(1 * time.Second)  // Simulate some work by sleeping for a second
			cancel()                     // Deregister the connection by calling the cancel function
		}()
	}

	wq.Wait() // Wait for all goroutines to finish

	fmt.Println(pool.WorkingNumber()) // Print the number of active connections in the pool after all goroutines have finished
	time.Sleep(5 * time.Second)       // Wait for any delayed operations to complete
	fmt.Println(pool.WorkingNumber()) // Print the number of active connections in the pool after the delay, to check for any changes
}
