package internal

import (
	"slices"
)

func (graph *Graph[E]) Optimize() {
	graph.Heads = make([]*GraphVertex[E], 0)
	graph.VertexSlice = make([]*GraphVertex[E], 0, len(graph.Vertices))
	complexVertices := make(map[string]bool, 0)

	priorityCmpFn := func(v1 *GraphVertex[E], v2 *GraphVertex[E]) int {
		if v1.Priority < v2.Priority {
			return 1
		} else if v1.Priority > v2.Priority {
			return -1
		} else {
			return 0
		}
	}

	for _, v := range graph.Vertices {
		graph.VertexSlice = append(graph.VertexSlice, v)
		if len(v.Dependencies) == 0 {
			graph.Heads = append(graph.Heads, v)
		}
		if len(v.Dependencies) > 1 || len(v.Next) > 1 {
			complexVertices[v.Name] = true
		}

		slices.SortFunc(v.Next, priorityCmpFn)
	}

	var zipped int

	for _, v := range graph.Vertices {
		v.Group = []*GraphVertex[E]{v} // default group of vertex only contain it self

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
				group[0].Group = group // head vertex hold the group
				zipped += len(group) - 1
			}
		}
	}

	slices.SortFunc(graph.Heads, priorityCmpFn)

	graph.ScheduleNum = len(graph.Vertices) - zipped
	graph.optimized = true
}
