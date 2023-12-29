package internal

type GraphMarshaler[E any] struct {
	Vertices map[string]E
	Edges    [][2]string
}

func (marshaler GraphMarshaler[E]) GenerateGraph() *Graph[E] {
	graph := NewGraph[E]()

	for name, v := range marshaler.Vertices {
		graph.AddVertex(name, v)
	}

	for _, e := range marshaler.Edges {
		graph.AddEdge(e[0], e[1])
	}

	return graph
}

func NewGraphMarshaler[E any](graph *Graph[E]) *GraphMarshaler[E] {
	graph.Lock()
	defer graph.Unlock()

	marshaler := new(GraphMarshaler[E])

	marshaler.Vertices = make(map[string]E)
	marshaler.Edges = make([][2]string, 0)

	for _, v := range graph.Vertices {
		marshaler.Vertices[v.Name] = v.Elem
	}

	for e := range graph.Edges {
		marshaler.Edges = append(marshaler.Edges, [2]string{
			e.From.Name, e.To.Name,
		})
	}

	return marshaler
}
