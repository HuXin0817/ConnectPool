package pool

import (
	"log"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/HuXin0817/ConnectPool/connector"
	"github.com/HuXin0817/ConnectPool/connectors"
)

const (
	DefaultMaxFreeTime       = connectors.DefaultMaxFreeTime       // Default maximum idle wait time
	DefaultAutoCleanInterval = connectors.DefaultAutoCleanInterval // Default auto-clean cycle execution period
)

var DefaultDealPanicMethod = func(panicInfo any) {
	log.Println(panicInfo) // Default method for handling panic by logging the panicInfo
}

type ConnectPool interface {
	Register() (newConnect any, cancelFunc func())                                    // Registers a connection
	RegisterWithTimeLimit(deadLine time.Duration) (newConnect any, cancelFunc func()) // Registers a connection with a deadline
	WorkingNumber() int64                                                             // Gets the number of active connections
	MaxSize() int64                                                                   // Gets the pool's maximum size
	SetMaxSize(size int64)                                                            // Sets the pool's maximum size
	MaxFreeTime() time.Duration                                                       // Gets the maximum idle time for connectors
	SetMaxFreeTime(time.Duration)                                                     // Sets the maximum idle time for connectors
	AutoClearInterval() time.Duration                                                 // Gets the interval for auto-clearing
	SetAutoClearInterval(time.Duration)                                               // Sets the interval for auto-clearing
	SetDealPanicMethod(func(panicInfo any))                                           // Sets the method for handling panic
	SetCloseMethod(func(any))                                                         // Sets the method to execute before closing a connection
	Close()                                                                           // Closes the pool
}

type connectPool struct {
	autoClearInterval time.Duration           // Interval for auto-clearing cycles
	maxFreeTime       time.Duration           // Maximum idle wait time
	maxSize           atomic.Int64            // Maximum number of connections
	pool              connectors.ConnectorSet // Pool of connectors
	connectMethod     func() any              // Method for creating connections
	dealPanicMethod   func(panicInfo any)     // Method for handling panic
	closeMethod       func(connect any)       // Method to execute before closing a connection
}

// NewConnectPool creates a new connection pool with a specified maximum size and connection creation method.
func NewConnectPool(maxSize int, connectMethod func() any) ConnectPool {
	// Initially use default values, which can be modified using Set methods
	pool := &connectPool{
		connectMethod:     connectMethod,
		autoClearInterval: DefaultAutoCleanInterval,
		maxFreeTime:       DefaultMaxFreeTime,
		dealPanicMethod:   DefaultDealPanicMethod,
	}

	pool.maxSize.Store(int64(maxSize))
	pool.pool = connectors.NewConnectorSet(&pool.autoClearInterval, &pool.maxFreeTime, &pool.closeMethod, &pool.dealPanicMethod)
	return pool
}

// searchConnector finds a connector in the connectPool.
func (p *connectPool) searchConnector() (Connect connector.Connector) {

	freeConnect := p.pool.GetFreeConnector() // Try to get a free connector from the existing pool
	if freeConnect != nil {
		Connect = freeConnect // If there is a free connector in the pool, use it directly
	}

	for {
		// If Connect is not nil, return it
		if Connect != nil {
			return
		}

		maxSize := p.MaxSize() // Get the maximum number of connections in the pool

		// Check if the pool has reached its maximum size, if not, create a new Connector
		if p.WorkingNumber() < maxSize {
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

func (p *connectPool) WorkingNumber() int64 {
	return int64(p.pool.Size())
}

func (p *connectPool) MaxSize() int64 {
	return p.maxSize.Load()
}

func (p *connectPool) SetMaxSize(size int64) {
	p.maxSize.Store(size)
}

func (p *connectPool) SetDealPanicMethod(dealPanicMethod func(panicInfo any)) {
	p.dealPanicMethod = dealPanicMethod
}

func (p *connectPool) MaxFreeTime() time.Duration {
	return time.Duration(atomic.LoadInt64((*int64)(&p.maxFreeTime)))
}

func (p *connectPool) SetMaxFreeTime(maxFreeTime time.Duration) {
	atomic.StoreInt64((*int64)(&p.maxFreeTime), int64(maxFreeTime))
}

func (p *connectPool) AutoClearInterval() time.Duration {
	return time.Duration(atomic.LoadInt64((*int64)(&p.autoClearInterval)))
}

func (p *connectPool) SetAutoClearInterval(autoCleanInterval time.Duration) {
	atomic.StoreInt64((*int64)(&p.autoClearInterval), int64(autoCleanInterval))
}

func (p *connectPool) SetCloseMethod(closeMethod func(any)) {
	p.closeMethod = closeMethod
}

func (p *connectPool) Close() {
	p.pool.Close() // Close the pool
}
