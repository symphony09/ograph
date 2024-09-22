package internal

func (graph *Graph[E]) Scheduling() (todo <-chan []*GraphVertex[E], done chan<- []*GraphVertex[E]) {
	todoCh := make(chan []*GraphVertex[E], 1+graph.ScheduleNum/2)
	doneCh := make(chan []*GraphVertex[E], 1+graph.ScheduleNum/2)

	graph.Lock()
	graph.reset()

	go func() {
		defer graph.Unlock()

		if graph.SerialGroups == nil {
			graph.Optimize()
		}

		if len(graph.Heads) == 0 {
			close(todoCh)
			return
		}

		for _, vertex := range graph.Heads {
			group := graph.SerialGroups[vertex.Name]

			if group == nil {
				group = append(group, vertex)
			}

			for _, v := range group {
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
				v.Status = StatusDone
			}

			graph.doingCnt--

			nextGroups, allDone := graph.findTodo(group[len(group)-1])
			if allDone {
				close(todoCh)
				return
			} else {
				for i, group := range nextGroups {
					if len(group) == 0 {
						continue
					}

					for _, v := range group {
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
	for _, v := range graph.Heads {
		graph.resetVertexStatus(v)
	}
}

func (graph *Graph[E]) resetVertexStatus(vertex *GraphVertex[E]) {
	vertex.Status = StatusTodo

	for _, v := range vertex.Next {
		graph.resetVertexStatus(v)
	}
}

func (graph *Graph[E]) findTodo(doneVertex *GraphVertex[E]) ([][]*GraphVertex[E], bool) {
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

				if graph.SerialGroups != nil {
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
