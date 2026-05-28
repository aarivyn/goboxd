package queue

import "runtime"

// Queue limits how many jobs run at the same time.
// If all slots are full, the next job waits — it never fails.
type Queue struct {
	sem chan struct{}
}

// New creates a queue. If limit is 0 it uses number of CPU cores.
func New(limit int) *Queue {
	if limit <= 0 {
		limit = runtime.NumCPU()
	}
	return &Queue{
		sem: make(chan struct{}, limit),
	}
}

// Run submits a job. If queue is full it WAITS until a slot is free.
func (q *Queue) Run(job func()) {
	// Take a slot — blocks if full
	q.sem <- struct{}{}
	// Release slot when job is done
	defer func() { <-q.sem }()
	job()
}

// InFlight returns how many jobs are currently running.
func (q *Queue) InFlight() int {
	return len(q.sem)
}