package connectpool

import (
	"log"
	"runtime"
	"sync/atomic"
	"time"
)

const (
	defaultMaxFreeTime       = 3 * time.Second // Default maximum idle wait time
	defaultAutoCleanInterval = 2 * time.Second // Default auto-clean cycle execution
	defaultCap               = 1000            // Default pool cap
)

var defaultDealPanicMethod = func(panicInfo any) {
	log.Println(panicInfo) // Default method for handling panic by logging the panicInfo
}

type ConnectPool interface {
	Register() (newConnect any, cancelFunc func())                                    // Registers a connection
	RegisterWithTimeLimit(deadLine time.Duration) (newConnect any, cancelFunc func()) // Registers a connection with a deadline
	WorkingNumber() int                                                               // Gets the number of active connections
	Size() int                                                                        // Gets the pool's cap
	Cap() int                                                                         // Gets the pool's maximum size
	MaxFreeTime() time.Duration                                                       // Gets the maximum idle time for connectors
	AutoClearInterval() time.Duration                                                 // Gets the interval for auto-clearing
	Close()                                                                           // Closes the pool
}

type connectPool struct {
	autoClearInterval time.Duration       // Interval for auto-clearing cycles
	maxFreeTime       time.Duration       // Maximum idle wait time
	cap               int                 // Maximum number of connections
	pool              connectorSet        // Pool of connectors
	connectMethod     func() any          // Method for creating connections
	dealPanicMethod   func(panicInfo any) // Method for handling panic
	closeMethod       func(connect any)   // Method to execute before closing a connection
}

// NewConnectPool creates a new connection pool with a specified maximum size and connection creation method.
func NewConnectPool(connectMethod func() any, options ...option) ConnectPool {
	// Initially use default values, which can be modified using Set methods
	pool := &connectPool{
		connectMethod:     connectMethod,
		autoClearInterval: defaultAutoCleanInterval,
		maxFreeTime:       defaultMaxFreeTime,
		cap:               defaultCap,
		dealPanicMethod:   defaultDealPanicMethod,
	}

	for _, op := range options {
		op(pool)
	}

	pool.pool = newConnectorSet(&pool.autoClearInterval, &pool.maxFreeTime, &pool.closeMethod, &pool.dealPanicMethod)
	return pool
}

// searchConnector finds a connector in the connectPool.
func (p *connectPool) searchConnector() (Connect connector) {

	freeConnect := p.pool.GetFreeConnector() // Try to get a free connector from the existing pool
	if freeConnect != nil {
		Connect = freeConnect // If there is a free connector in the pool, use it directly
	}

	for {
		// If Connect is not nil, return it
		if Connect != nil {
			return
		}

		maxSize := p.Cap() // Get the maximum number of connections in the pool

		// Check if the pool has reached its maximum size, if not, create a new Connector
		if p.Size() < maxSize {
			return p.pool.AddConnector(&p.connectMethod, &p.dealPanicMethod) // Create and return a new Connector in the pool
		}

		runtime.Gosched() // Yield the processor to allow other goroutines to run
	}
}

func (p *connectPool) Register() (newConnect any, cancelFunc func()) {
	c := p.searchConnector()
	if c == nil {
		return nil, nil
	}

	c.StartWorking()
	return c.GetConnect(), c.StopWorking
}

func (p *connectPool) RegisterWithTimeLimit(deadLine time.Duration) (newConnect any, cancelFunc func()) {
	c := p.searchConnector()
	if c == nil {
		return nil, nil
	}

	c.StartTimingWork(deadLine)
	return c.GetConnect(), c.StopWorking
}

func (p *connectPool) WorkingNumber() int {
	return int(p.pool.WorkingNumber())
}

func (p *connectPool) Cap() int {
	return p.cap
}

func (p *connectPool) MaxFreeTime() time.Duration {
	return time.Duration(atomic.LoadInt64((*int64)(&p.maxFreeTime)))
}

func (p *connectPool) AutoClearInterval() time.Duration {
	return time.Duration(atomic.LoadInt64((*int64)(&p.autoClearInterval)))
}

func (p *connectPool) Size() int {
	return p.pool.Size()
}

func (p *connectPool) Close() {
	p.pool.Close() // Close the pool
}
