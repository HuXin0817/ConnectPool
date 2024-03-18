package connectpool

import (
	"sync/atomic"
	"time"
)

type connector interface {
	GetConnect() any                             // Get the Connector's connection variable
	SinceLastWorkingTime() time.Duration         // Get the time since the Connector last worked
	IsFree() bool                                // Determine if the Connector is free
	StartWorking()                               // Begin working
	StopWorking()                                // End working
	StartTimingWork(time.Duration)               // Start working for a specified duration
	Do(f *func(any), dealPanicMethod *func(any)) // Invoke an external method and handle any potential Panic
}

type atomicConnector struct {
	connect         any           // Connection variable
	isWorking       atomic.Bool   // Working state
	lastWorkingTime atomic.Value  // Last work time, stored as time.Time
	waitCloseState  atomic.Bool   // State of waiting to automatically stop working
	stopSignalChan  chan struct{} // Channel for transmitting work stop signals
}

// newConnector creates a new connector with connect as the connection variable
func newConnector(connectMethod *func() any, dealPanicMethod *func(any)) connector {

	c := &atomicConnector{
		stopSignalChan: make(chan struct{}, 1), // Allocate a buffer of length 1 for stopSignalChan
	}

	c.updateLastWorkingTime() // Update the working time to the most recent

	func() {
		defer func() {
			// If dealPanicMethod is not nil, invoke dealPanicMethod to handle any possible panic
			if r := recover(); r != nil && dealPanicMethod != nil && *dealPanicMethod != nil {
				(*dealPanicMethod)(r)
			}
		}()

		// If the connection strategy is nil, abandon this connection attempt
		if connectMethod == nil || *connectMethod == nil {
			return
		}

		// Store the connection variable in c.connect
		c.connect = (*connectMethod)()
	}()

	return c
}

func (c *atomicConnector) GetConnect() any {
	return c.connect
}

func (c *atomicConnector) StartWorking() {
	c.isWorking.Store(true)
}

func (c *atomicConnector) StopWorking() {
	c.isWorking.Store(false)  // Update the working state
	c.updateLastWorkingTime() // Update the last working time

	// If in waitCloseState, send an end signal to stopSignalChan
	if c.waitCloseState.Load() {
		c.stopSignalChan <- struct{}{}
	}
}

// updateLastWorkingTime updates the working time to the most recent
func (c *atomicConnector) updateLastWorkingTime() {
	c.lastWorkingTime.Store(time.Now())
}

// endTimingWork ends TimingWork
func (c *atomicConnector) endTimingWork() {
	c.waitCloseState.Store(false) // End the connector's waitCloseState
	c.isWorking.Store(false)
	c.updateLastWorkingTime()
}

func (c *atomicConnector) StartTimingWork(deadline time.Duration) {
	// Start a new goroutine, asynchronously wait and end work
	go func() {
		c.waitCloseState.Store(true) // Make the connector enter waitCloseState

		c.StartWorking()

		timer := time.NewTimer(deadline) // Set a timer with a deadline duration

		// Exit TimingWork upon meeting one of the conditions
		select {
		case <-timer.C: // Time reached the deadline
			c.endTimingWork()

		case <-c.stopSignalChan: // External force actively ended TimingWork
			c.endTimingWork()
		}
	}()
}

func (c *atomicConnector) IsFree() bool {
	return !c.isWorking.Load()
}

func (c *atomicConnector) SinceLastWorkingTime() time.Duration {
	// If the connector is working, return 0
	if !c.IsFree() {
		return 0
	}

	t := c.lastWorkingTime.Load().(time.Time)
	return time.Since(t)
}

func (c *atomicConnector) Do(f *func(any), dealPanicMethod *func(any)) {
	defer func() {
		// Handle any panic that occurs during work
		if r := recover(); r != nil && dealPanicMethod != nil && *dealPanicMethod != nil {
			(*dealPanicMethod)(r)
		}
	}()

	// If the function is nil, abandon executing it
	if f == nil || *f == nil {
		return
	}

	(*f)(c.connect)
}
