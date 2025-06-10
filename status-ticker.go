package main

import "time"

type statusTicker struct {
	interval time.Duration
	done     chan bool
	statusFn func()
	ticker   *time.Ticker
}

func newStatusTicker(interval time.Duration, statusFn func()) *statusTicker {
	return &statusTicker{interval: interval, done: make(chan bool), statusFn: statusFn}
}

func (t *statusTicker) start() {
	t.ticker = time.NewTicker(t.interval)
	go func() {
		for {
			select {
			case <-t.done:
				return
			case <-t.ticker.C:
				t.statusFn()
			}
		}
	}()
}

func (t *statusTicker) stop() {
	t.ticker.Stop()
	t.done <- true
}
