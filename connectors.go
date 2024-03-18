package connectpool

import (
	"sync"
	"sync/atomic"
	"time"
)

type connectorSet interface {
	AddConnector(connectMethod *func() any, dealPanicMethod *func(panicInfo any)) (newConnector connector)       // Adds a new Connector
	GetFreeConnector() connector                                                                                 // Retrieves a free Connector
	Size() int                                                                                                   // Returns the size of the connector set
	WorkingNumber() int64                                                                                        // Returns the count of the Working Connector
	Close()                                                                                                      // Closes the ConnectorSet, terminating the Set's AutoClear
	Clear(maxFreeTime *time.Duration, closeMethod *func(any), dealPanicMethod *func(any))                        // Actively performs a cleanup
	autoClear(autoClearInterval, maxFreeTime *time.Duration, closeMethod *func(any), dealPanicMethod *func(any)) // Asynchronously performs the auto-cleanup function
}

type autoClearConnectorSet struct {
	token               atomic.Uint64        // An internally incremented Token for encoding Connectors
	closed              atomic.Bool          // Indicates whether it's closed
	connectorSet        map[uint64]connector // Collection of Connectors
	connectorSetRWMutex sync.RWMutex         // Read-write lock protecting the connector collection
}

func newConnectorSet(autoClearInterval, maxFreeTime *time.Duration, closeMethod *func(any), dealPanicMethod *func(any)) (NewConnectorSet connectorSet) {
	NewConnectorSet = &autoClearConnectorSet{
		connectorSet: make(map[uint64]connector),
	}

	go NewConnectorSet.autoClear(autoClearInterval, maxFreeTime, closeMethod, dealPanicMethod) // Starts a new goroutine to periodically clean up Connectors
	return NewConnectorSet
}

func (s *autoClearConnectorSet) Clear(maxFreeTime *time.Duration, closeMethod *func(any), dealPanicMethod *func(any)) {

	var RemoveList []uint64

	// Finds all Connectors to be removed under a read lock
	s.connectorSetRWMutex.RLock()

	for key, value := range s.connectorSet {
		// Actively cleans up the Connector if a nil Connector is found
		if value == nil || value.GetConnect() == nil {
			RemoveList = append(RemoveList, key)
			continue
		}

		if value.SinceLastWorkingTime() > *maxFreeTime {
			RemoveList = append(RemoveList, key)

			// Executes the respective closeMethod before removal
			value.Do(closeMethod, dealPanicMethod)
		}
	}

	s.connectorSetRWMutex.RUnlock()

	if len(RemoveList) > 0 {

		// Removes the Connectors listed in RemoveList under a write lock
		s.connectorSetRWMutex.Lock()
		defer s.connectorSetRWMutex.Unlock()

		for _, key := range RemoveList {
			delete(s.connectorSet, key)
		}
	}
}

func (s *autoClearConnectorSet) autoClear(autoClearInterval, maxFreeTime *time.Duration, closeMethod *func(any), dealPanicMethod *func(any)) {
	for {

		// Determines AutoClearInterval; uses defaultAutoCleanInterval if autoClearInterval is nil
		AutoClearInterval := defaultAutoCleanInterval
		if autoClearInterval != nil {
			AutoClearInterval = *autoClearInterval
		}

		// Creates a timer with a length of AutoClearInterval
		timer := time.NewTimer(AutoClearInterval)

		// Determines MaxFreeTime; uses defaultMaxFreeTime if maxFreeTime is nil
		MaxFreeTime := defaultMaxFreeTime
		if maxFreeTime != nil {
			MaxFreeTime = *maxFreeTime
		}

		s.Clear(&MaxFreeTime, closeMethod, dealPanicMethod) // Automatically performs a cleanup

		// Terminates the cleanup thread if the Set is closed
		if s.closed.Load() {
			return
		}

		<-timer.C // Waits for the timer to expire
	}
}

func (s *autoClearConnectorSet) registerToken() uint64 {
	return s.token.Add(1) // Increment token, ensuring a unique token value each time
}

func (s *autoClearConnectorSet) AddConnector(connectMethod *func() any, dealPanicMethod *func(panicInfo any)) (NewConnector connector) {

	var contains bool
	var connectorToken uint64

	s.connectorSetRWMutex.RLock()

	// Finds an unused Token in the connectorSet
	for {
		// Registers a Token
		connectorToken = s.registerToken()

		// Checks if the newToken already exists in the connectorSet
		_, contains = s.connectorSet[connectorToken]

		// If not, uses this Token
		if !contains {
			break
		}
	}

	s.connectorSetRWMutex.RUnlock()

	// Obtains a new Connector
	NewConnector = newConnector(connectMethod, dealPanicMethod)

	s.connectorSetRWMutex.Lock()
	// Inserts connectorToken and NewConnector into the dictionary
	s.connectorSet[connectorToken] = NewConnector
	s.connectorSetRWMutex.Unlock()

	return
}

func (s *autoClearConnectorSet) GetFreeConnector() connector {

	// Uses a write lock to ensure the retrieved FreeConnector is only used by one owner
	s.connectorSetRWMutex.Lock()
	defer s.connectorSetRWMutex.Unlock()

	for _, v := range s.connectorSet {
		if v.IsFree() {
			v.StartWorking() // Marks the retrieved FreeConnector as busy to avoid reuse
			return v
		}
	}

	return nil
}

func (s *autoClearConnectorSet) Size() (size int) {
	s.connectorSetRWMutex.RLock()
	defer s.connectorSetRWMutex.RUnlock()

	size = len(s.connectorSet)
	return
}

func (s *autoClearConnectorSet) Close() {
	s.connectorSetRWMutex.Lock()
	defer s.connectorSetRWMutex.Unlock()

	s.closed.Store(true)  // Signals the autoClear coroutine to terminate
	clear(s.connectorSet) // Cleans up the connectorSet to avoid memory usage
}

func (s *autoClearConnectorSet) WorkingNumber() int64 {
	s.connectorSetRWMutex.RLock()
	defer s.connectorSetRWMutex.RUnlock()

	cnt := int64(0)
	for _, v := range s.connectorSet {
		if !v.IsFree() {
			cnt++
		}
	}

	return cnt
}
