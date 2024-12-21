package internal

import (
	"iter"
)

func (graph *Graph[E]) Scheduling(interrupts iter.Seq[string]) (todo <-chan []*GraphVertex[E], done chan<- []*GraphVertex[E]) {
	todoCh := make(chan []*GraphVertex[E], 1+graph.ScheduleNum/2)
	doneCh := make(chan []*GraphVertex[E], 1+graph.ScheduleNum/2)

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

		if enableSerialGroup && graph.SerialGroups == nil {
			graph.Optimize()
		}

		if len(graph.Heads) == 0 {
			close(todoCh)
			return
		}

		for _, vertex := range graph.Heads {
			var group []*GraphVertex[E]

			if enableSerialGroup {
				group = graph.SerialGroups[vertex.Name]
			}

			if group == nil {
				group = append(group, vertex)
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
			}

			graph.doingCnt--

			nextGroups, allDone := graph.findTodo(group[len(group)-1], enableSerialGroup)
			if allDone {
				close(todoCh)
				return
			} else {
				for i, group := range nextGroups {
					if len(group) == 0 {
						continue
					}

					for _, v := range group {
						if doInterrupt && (interruptAt == v.Name+":start" || interruptAt == "*:start" || interruptAt == "*") {
							interruptAt, doInterrupt = nextInterrupt()
						}

						v.Status = StatusDoing
					}

					graph.doingCnt++

					todoCh <- nextGroups[i]
				}
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
		}
	} else {
		for _, v := range graph.Heads {
			graph.resetVertexStatus(v)
		}
	}
}

func (graph *Graph[E]) resetVertexStatus(vertex *GraphVertex[E]) {
	vertex.Status = StatusTodo

	for _, v := range vertex.Next {
		graph.resetVertexStatus(v)
	}
}

func (graph *Graph[E]) findTodo(doneVertex *GraphVertex[E], enableSerialGroup bool) ([][]*GraphVertex[E], bool) {
	var vertexGroups [][]*GraphVertex[E]

	if doneVertex != nil {
		for i, next := range doneVertex.Next {
			if next.Status == StatusDoing {
				continue
			}

			var notReady bool

			for _, dep := range next.Dependencies {
				if dep.Status != StatusDone {
					notReady = true
				}
			}

			if !notReady {
				var serialGroup []*GraphVertex[E]

				if enableSerialGroup && graph.SerialGroups != nil {
					serialGroup = graph.SerialGroups[doneVertex.Next[i].Name]
				}

				if serialGroup != nil {
					vertexGroups = append(vertexGroups, serialGroup)
				} else {
					vertexGroups = append(vertexGroups, []*GraphVertex[E]{doneVertex.Next[i]})
				}
			}
		}
	}

	if len(vertexGroups) == 0 && graph.doingCnt == 0 {
		return vertexGroups, true
	}

	return vertexGroups, false
}
