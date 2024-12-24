package internal

import (
	"sync"
)

const (
	StatusTodo = iota
	StatusDoing
	StatusDone
)

type Graph[E any] struct {
	Vertices map[string]*GraphVertex[E]
	Edges    map[GraphEdge[E]]bool

	Heads        []*GraphVertex[E]
	VertexSlice  []*GraphVertex[E]
	SerialGroups map[string][]*GraphVertex[E]
	ScheduleNum  int

	doingCnt int

	sync.Mutex
}

type GraphEdge[E any] struct {
	From *GraphVertex[E]
	To   *GraphVertex[E]
}

type GraphVertex[E any] struct {
	Name   string
	Status int
	Wait   int
	Elem   E

	Dependencies []*GraphVertex[E]
	Next         []*GraphVertex[E]
}

func (graph *Graph[E]) AddVertex(name string, elem E) {
	graph.Vertices[name] = &GraphVertex[E]{
		Name: name,
		Elem: elem,

		Dependencies: make([]*GraphVertex[E], 0),
		Next:         make([]*GraphVertex[E], 0),
	}
}

func (graph *Graph[E]) AddEdge(from, to string) {
	if fromVertex, toVertex := graph.Vertices[from], graph.Vertices[to]; fromVertex != nil && toVertex != nil {
		fromVertex.Next = append(fromVertex.Next, toVertex)
		toVertex.Dependencies = append(toVertex.Dependencies, fromVertex)

		edge := GraphEdge[E]{
			From: fromVertex,
			To:   toVertex,
		}

		graph.Edges[edge] = true
	}
}

type Mapper[OE any, NE any] func(OE) (NE, error)

func MapToNewGraph[OE any, NE any](graph *Graph[OE], mapper Mapper[OE, NE]) (*Graph[NE], error) {
	var newGraph Graph[NE]
	newVertices := make(map[string]*GraphVertex[NE])
	newEdges := make(map[GraphEdge[NE]]bool)

	for _, vertex := range graph.Vertices {
		if newElem, err := mapper(vertex.Elem); err != nil {
			return &newGraph, err
		} else {
			newVertices[vertex.Name] = &GraphVertex[NE]{
				Name:   vertex.Name,
				Elem:   newElem,
				Status: vertex.Status,

				Dependencies: make([]*GraphVertex[NE], 0),
				Next:         make([]*GraphVertex[NE], 0),
			}
		}
	}

	for _, vertex := range graph.Vertices {
		newVertex := newVertices[vertex.Name]

		for _, dep := range vertex.Dependencies {
			newDep := newVertices[dep.Name]
			newVertex.Dependencies = append(newVertex.Dependencies, newDep)
		}

		for _, next := range vertex.Next {
			newNext := newVertices[next.Name]
			newVertex.Next = append(newVertex.Next, newNext)
		}
	}

	for edge, ok := range graph.Edges {
		if ok {
			newEdge := GraphEdge[NE]{
				From: newVertices[edge.From.Name],
				To:   newVertices[edge.To.Name],
			}

			newEdges[newEdge] = true
		}
	}

	newGraph.Vertices = newVertices
	newGraph.Edges = newEdges

	return &newGraph, nil
}

func NewGraph[E any]() *Graph[E] {
	return &Graph[E]{
		Edges:    make(map[GraphEdge[E]]bool),
		Vertices: make(map[string]*GraphVertex[E]),
	}
}
