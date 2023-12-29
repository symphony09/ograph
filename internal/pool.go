package internal

import (
	"sync"
)

type WorkerPool struct {
	cache WorkerCache
	lower sync.Pool
}

func (pool *WorkerPool) Get() (*Worker, bool) {
	if worker, ok := pool.cache.Get(); ok && worker != nil {
		return worker, true
	} else if worker, ok = pool.lower.Get().(*Worker); ok && worker != nil {
		return worker, true
	} else {
		return nil, false
	}
}

func (pool *WorkerPool) Put(worker *Worker) {
	if ok := pool.cache.Put(worker); ok {
		return
	} else {
		pool.lower.Put(worker)
	}
}

type WorkerCache struct {
	MaxSize int

	workers []*Worker

	sync.Mutex
}

func (cache *WorkerCache) Get() (*Worker, bool) {
	if cache.MaxSize <= 0 {
		return nil, false
	}

	cache.Lock()
	defer cache.Unlock()

	if len(cache.workers) == 0 {
		return nil, false
	}

	worker := cache.workers[len(cache.workers)-1]
	cache.workers[len(cache.workers)-1] = nil
	cache.workers = cache.workers[0 : len(cache.workers)-1]

	return worker, true
}

func (cache *WorkerCache) Put(worker *Worker) bool {
	if cache.MaxSize <= 0 {
		return false
	}

	cache.Lock()
	defer cache.Unlock()

	if len(cache.workers) >= cache.MaxSize {
		return false
	}

	cache.workers = append(cache.workers, worker)

	return true
}

func NewWorkerPool(cacheSize int) *WorkerPool {
	pool := new(WorkerPool)
	pool.cache = WorkerCache{
		MaxSize: cacheSize,
	}

	return pool
}
