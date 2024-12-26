package internal

import (
	"iter"
	"runtime"
)

func (graph *Graph[E]) Scheduling(interrupts iter.Seq[string], parallelismLimit int) (todo <-chan []*GraphVertex[E], done chan<- []*GraphVertex[E]) {
	scheduleChanSize := 1 + graph.ScheduleNum/2

	if parallelismLimit <= 0 {
		scheduleChanSize = min(scheduleChanSize, runtime.GOMAXPROCS(0))
	}

	todoCh := make(chan []*GraphVertex[E], scheduleChanSize)
	doneCh := make(chan []*GraphVertex[E], scheduleChanSize)

	graph.Lock()
	graph.reset()

	go func() {
		defer graph.Unlock()

		var nextInterrupt func() (string, bool)
		var stopInterrupt func()
		var interruptAt string
		var doInterrupt bool

		if interrupts != nil {
			nextInterrupt, stopInterrupt = iter.Pull(interrupts)
			defer stopInterrupt()

			interruptAt, doInterrupt = nextInterrupt()
		}

		enableSerialGroup := interrupts == nil

		if !graph.optimized {
			graph.Optimize()
		}

		if len(graph.Heads) == 0 {
			close(todoCh)
			return
		}

		for _, vertex := range graph.Heads {
			var group []*GraphVertex[E] = vertex.Group

			if !enableSerialGroup && len(group) > 1 {
				group = group[:1]
			}

			for _, v := range group {
				if doInterrupt && (interruptAt == vertex.Name+":start" || interruptAt == "*:start" || interruptAt == "*") {
					interruptAt, doInterrupt = nextInterrupt()
				}

				v.Status = StatusDoing
			}

			graph.doingCnt++
			todoCh <- group
		}

		for group := range doneCh {
			if len(group) == 0 {
				continue
			}

			for _, v := range group {
				if doInterrupt && (interruptAt == v.Name+":end" || interruptAt == "*:end" || interruptAt == "*") {
					interruptAt, doInterrupt = nextInterrupt()
				}

				v.Status = StatusDone

				for _, next := range v.Next {
					next.Wait--
				}
			}

			graph.doingCnt--

			handleTodo := func(group []*GraphVertex[E]) {
				if len(group) == 0 {
					return
				}

				for _, v := range group {
					if doInterrupt && (interruptAt == v.Name+":start" || interruptAt == "*:start" || interruptAt == "*") {
						interruptAt, doInterrupt = nextInterrupt()
					}

					v.Status = StatusDoing
				}

				graph.doingCnt++

				todoCh <- group
			}

			allDone := graph.findTodo(group[len(group)-1], enableSerialGroup, handleTodo)
			if allDone {
				close(todoCh)
				return
			}
		}

		close(todoCh)
	}()

	return todoCh, doneCh
}

func (graph *Graph[E]) reset() {
	if graph.VertexSlice != nil {
		for _, v := range graph.VertexSlice {
			v.Status = StatusTodo
			v.Wait = len(v.Dependencies)
		}
	} else {
		for _, v := range graph.Heads {
			graph.resetVertexStatus(v)
		}
	}
}

func (graph *Graph[E]) resetVertexStatus(vertex *GraphVertex[E]) {
	vertex.Status = StatusTodo
	vertex.Wait = len(vertex.Dependencies)

	for _, v := range vertex.Next {
		graph.resetVertexStatus(v)
	}
}

func (graph *Graph[E]) findTodo(doneVertex *GraphVertex[E], enableSerialGroup bool, then func(group []*GraphVertex[E])) bool {
	var todoCnt int

	if doneVertex != nil {
		for i, next := range doneVertex.Next {
			if next.Status == StatusDoing {
				continue
			}

			if next.Wait == 0 {
				group := doneVertex.Next[i].Group

				if !enableSerialGroup && len(group) > 1 {
					group = group[:1]
				}

				todoCnt++
				then(group)
			}
		}
	}

	if todoCnt == 0 && graph.doingCnt == 0 {
		return true
	}

	return false
}
