package ograph

type Op func(pipeline *Pipeline, node *Element)

var DependOn = func(dependencies ...*Element) Op {
	return func(pipeline *Pipeline, element *Element) {
		for _, dep := range dependencies {
			if pipeline.elements[dep.Name] == nil {
				pipeline.Register(dep)
			}

			if pipeline.elements[dep.Name] == dep {
				pipeline.graph.AddEdge(dep.Name, element.Name)
			}
		}
	}
}

var Then = func(nextElements ...*Element) Op {
	return func(pipeline *Pipeline, element *Element) {
		for _, next := range nextElements {
			if pipeline.elements[next.Name] == nil {
				pipeline.Register(next)
			}

			if pipeline.elements[next.Name] == next {
				pipeline.graph.AddEdge(element.Name, next.Name)
			}
		}
	}
}
