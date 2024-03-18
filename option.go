package connectpool

import "time"

type option func(*connectPool)

func WithCap(cap int) option {
	return func(pool *connectPool) {
		pool.cap = cap
	}
}

func WithMaxFreeTime(maxFreeTime time.Duration) option {
	return func(pool *connectPool) {
		pool.maxFreeTime = maxFreeTime
	}
}

func WithAutoClearInterval(autoClearInterval time.Duration) option {
	return func(pool *connectPool) {
		pool.autoClearInterval = autoClearInterval
	}
}

func WithDealPanicMethod(dealPanicMethod func(panicInfo any)) option {
	return func(pool *connectPool) {
		pool.dealPanicMethod = dealPanicMethod
	}
}

func WithCloseMethod(closeMethod func(connect any)) option {
	return func(pool *connectPool) {
		pool.closeMethod = closeMethod
	}
}
