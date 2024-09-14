// /home/krylon/go/src/krylib/semaphor.go
// -*- mode: go; coding: utf-8; -*-
// Created on 17. 07. 2017 by Benjamin Walkenhorst
// (c) 2017 Benjamin Walkenhorst
// Time-stamp: <2017-07-18 19:59:12 krylon>

package krylib

import "sync"

// Semaphore implements a classic among synchronization
// facilities.
type Semaphore struct {
	lock  sync.Mutex
	max   int
	cur   int
	empty *sync.Cond
}

// NewSemaphore creates, uh, a new semaphore with a maximum
// counter value of max.
func NewSemaphore(max int) *Semaphore {
	var sem = &Semaphore{
		max: max,
		cur: max,
	}

	sem.empty = sync.NewCond(&sem.lock)

	return sem
} // func NewSemaphore(max int) *Semaphore

// Dec decreases the counter by one, atomically.
// If the counter is currently zero, it blocks until another goroutine
// calls Inc() on the same semaphore.
func (s *Semaphore) Dec() {
	s.lock.Lock()

CHECK:
	if s.cur > 0 {
		s.cur--
		s.lock.Unlock()
	} else {
		s.empty.Wait()
		goto CHECK
	}
} // func (s *Semaphore) Dec()

// Inc increases the counter. If the counter was previously zero (0),
// it will signal the next goroutine waiting on the Semaphore.
func (s *Semaphore) Inc() {
	s.lock.Lock()

	var zero = s.cur == 0

	s.cur++

	if zero {
		s.empty.Broadcast()
	}

	s.lock.Unlock()
} // func (s *Semaphore) Inc()
