package connectpool

import "time"

type Option func(*connectPool)

func WithCap(cap int) Option {
	return func(pool *connectPool) {
		pool.cap = cap
	}
}

func WithMaxFreeTime(maxFreeTime time.Duration) Option {
	return func(pool *connectPool) {
		pool.maxFreeTime = maxFreeTime
	}
}

func WithAutoClearInterval(autoClearInterval time.Duration) Option {
	return func(pool *connectPool) {
		pool.autoClearInterval = autoClearInterval
	}
}

func WithDealPanicMethod(dealPanicMethod func(panicInfo any)) Option {
	return func(pool *connectPool) {
		pool.dealPanicMethod = dealPanicMethod
	}
}

func WithCloseMethod(closeMethod func(connect any)) Option {
	return func(pool *connectPool) {
		pool.closeMethod = closeMethod
	}
}
