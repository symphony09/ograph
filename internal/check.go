package internal

import "fmt"

func (graph *Graph[E]) Check() error {
	_, left := graph.Steps()

	if len(left) != 0 {
		return fmt.Errorf("found cycle between vertices: %v", left)
	} else {
		return nil
	}
}

func (graph *Graph[E]) Steps() ([][]string, []string) {
	steps := make([][]string, 0)

	graph.Lock()
	graph.reset()

	for {
		var names []string

		for name, vertex := range graph.Vertices {
			// vertex is todo and all dep vertices are done, so it can be processed
			if vertex.Status == StatusTodo {
				ready := true

				for _, dep := range vertex.Dependencies {
					if dep.Status != StatusDone {
						ready = false
					}
				}

				if ready {
					names = append(names, name)
				}
			}
		}

		if len(names) == 0 {
			break
		}

		// set vertex status to be done
		for _, name := range names {
			graph.Vertices[name].Status = StatusDone
		}

		steps = append(steps, names)
	}

	// found vertices which are never done
	left := make([]string, 0)
	for _, vertex := range graph.Vertices {
		if vertex.Status != StatusDone {
			left = append(left, vertex.Name)
		}
	}

	graph.Unlock()

	return steps, left
}
