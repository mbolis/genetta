package workerpool

import (
	"errors"
	"iter"
	"math"
	"runtime"
	"sync"
)

type Pool[W any] interface {
	All() iter.Seq2[int, W]
	Status() Status

	Offer(int)
	Wait()
	Resume()
	Close()
}

type Status uint

const (
	StatusRunning Status = iota
	StatusIdle
)

type pool[W any] struct {
	workers []W
	queue   chan int
	resumeQ chan struct{}

	status Status
	wgroup sync.WaitGroup
}

const stopSignal = math.MinInt

var (
	ErrEmptyPool   = errors.New("empty pool")
	ErrEmptyBuffer = errors.New("empty buffer")
)

func New[W any](nWorkers, bufferSize int, work func(*W, int)) (Pool[W], error) {
	if nWorkers == 0 {
		return nil, ErrEmptyPool
	}
	if bufferSize <= 0 {
		return nil, ErrEmptyBuffer
	}

	var emptyCtx W

	p := &pool[W]{
		workers: make([]W, nWorkers),
		queue:   make(chan int, bufferSize),
		resumeQ: make(chan struct{}, nWorkers),
	}

	p.wgroup.Add(nWorkers)

	for i := range nWorkers {
		go func(i int) {
		running:
			for in := range p.queue {
				if in == stopSignal {
					p.wgroup.Done()
					goto parked
				}

				work(&p.workers[i], in)
				runtime.Gosched()
			}
		parked:
			for range p.resumeQ {
				p.wgroup.Add(1)
				p.workers[i] = emptyCtx
				goto running
			}
		}(i)
	}

	return p, nil
}

func (p *pool[W]) All() iter.Seq2[int, W] {
	return p.iterate
}
func (p *pool[W]) iterate(yield func(int, W) bool) {
	for i, w := range p.workers {
		if !yield(i, w) {
			break
		}
	}
}

func (p *pool[W]) Status() Status {
	return p.status
}

func (p *pool[W]) Offer(input int) {
	p.queue <- input
}

func (p *pool[W]) Wait() {
	if p.status == StatusIdle {
		return
	}

	for range p.workers {
		p.queue <- stopSignal
	}

	p.wgroup.Wait()
	p.status = StatusIdle
}

func (p *pool[W]) Resume() {
	if p.status != StatusIdle {
		panic("cannot resume running pool")
	}
	p.status = StatusRunning

	for range p.workers {
		p.resumeQ <- struct{}{}
	}
}

func (p *pool[W]) Close() {
	close(p.queue)
	close(p.resumeQ)
}
