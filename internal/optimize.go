package internal

func (graph *Graph[E]) Optimize() {
	graph.Heads = make([]*GraphVertex[E], 0)
	graph.SerialGroups = make(map[string][]*GraphVertex[E])
	complexVertices := make(map[string]bool, 0)

	for _, v := range graph.Vertices {
		if len(v.Dependencies) == 0 {
			graph.Heads = append(graph.Heads, v)
		}
		if len(v.Dependencies) > 1 || len(v.Next) > 1 {
			complexVertices[v.Name] = true
		}
	}

	for _, v := range graph.Vertices {
		if !complexVertices[v.Name] {
			var group []*GraphVertex[E]

			if len(v.Dependencies) == 0 || complexVertices[v.Dependencies[0].Name] {
				for {
					group = append(group, v)

					if len(v.Next) == 1 {
						v = v.Next[0]
					} else {
						break
					}

					if complexVertices[v.Name] {
						break
					}
				}
			}

			if len(group) > 1 {
				graph.SerialGroups[group[0].Name] = group
			}
		}
	}

	var zipped int
	for _, group := range graph.SerialGroups {
		if len(group) > 1 {
			zipped += len(group) - 1
		}
	}

	graph.ScheduleNum = len(graph.Vertices) - zipped
}
