package connectors

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var printLock sync.Mutex // Lock for synchronizing log prints

var cnt atomic.Int32 // Counter for generating unique connect identifiers

// mockConnectMethod simulates a connection method, incrementing and logging a counter.
func mockConnectMethod() any {
	printLock.Lock()
	defer printLock.Unlock()

	c := cnt.Add(1) // Increment the counter for each new connection

	log.Printf("new connect %v\n", c) // Log the creation of a new connection
	return c                          // Return the counter value as the connection identifier
}

// mockCloseFunc simulates a function to close a connection, logging the connection that's being closed.
func mockCloseFunc(connect any) {
	printLock.Lock()
	defer printLock.Unlock()

	log.Printf("connect %v close", connect) // Log the closing of a connection
}

// mockDealPanicMethod simulates a panic handling method, logging the panic information.
func mockDealPanicMethod(panicInfo any) {
	printLock.Lock()
	defer printLock.Unlock()

	log.Printf("panic: %v\n", panicInfo) // Log the panic information
}

// printSize logs the size of the ConnectorSet.
func printSize(s ConnectorSet) {
	printLock.Lock()
	defer printLock.Unlock()

	fmt.Println(s.Size()) // Print the size of the ConnectorSet
}

// TestAutoClear tests the automatic clearing of connectors from the set after a specified duration.
func TestAutoClear(t *testing.T) {

	sc := time.Second // Second count for intervals

	var autoClearInterval, maxFreeTime *time.Duration
	autoClearInterval = &sc // Interval after which auto-clearing of connectors is triggered
	maxFreeTime = &sc       // Maximum time a connector can remain idle before being cleared

	mc := mockConnectMethod    // Mock connect method
	mcc := mockCloseFunc       // Mock close function
	mdp := mockDealPanicMethod // Mock panic handling method

	s := NewConnectorSet(autoClearInterval, maxFreeTime, &mcc, &mdp) // Initialize the ConnectorSet with auto-clearing

	var PoolSize = 10000  // Size of the connector pool to simulate
	var wg sync.WaitGroup // WaitGroup to manage goroutines

	wg.Add(PoolSize)

	for range PoolSize {
		go func() {
			defer wg.Done() // Signal completion upon goroutine exit

			s.AddConnector(&mc, &mdp) // Add a new connector

			time.Sleep(1 * time.Second) // Simulate work by sleeping

			c := s.GetFreeConnector() // Attempt to get a free connector
			if c == nil {
				return // Exit if no free connector is available
			}

			c.StartTimingWork(1 * time.Second) // Start timing work for the connector

			f := func(a any) {
				panic(a) // Simulate a panic
			}

			c.Do(&f, &mdp) // Execute a function that panics, using the mock deal panic method
		}()
	}

	printSize(s)                // Print the size of the connector set before waiting
	wg.Wait()                   // Wait for all goroutines to complete
	printSize(s)                // Print the size of the connector set after waiting
	time.Sleep(3 * time.Second) // Wait for auto-clear to potentially run
	printSize(s)                // Print the final size of the connector set
}
