package ograph

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
)

func (pipeline *Pipeline) DumpDOT() ([]byte, error) {
	buf := &bytes.Buffer{}

	buf.WriteString(fmt.Sprintf("digraph %s {\n", pipeline.name))

	for _, v := range pipeline.graph.Vertices {
		if err := pipeline.dumpDOTNode(buf, v.Elem, "\t"); err != nil {
			return nil, err
		}
	}

	for edge := range pipeline.graph.Edges {
		if err := pipeline.dumpDOTEdge(buf, edge.From.Elem, edge.To.Elem, "\t"); err != nil {
			return nil, err
		}
	}

	buf.WriteString("}\n")

	return buf.Bytes(), nil
}

func (pipeline *Pipeline) dumpDOTNode(buf *bytes.Buffer, elem *Element, indent string) error {
	var tags []string

	tags = append(tags, fmt.Sprintf("Name: %s", elem.Name))

	if elem.Virtual {
		tags = append(tags, "Virtual Node")

		if len(elem.ImplElements) > 0 {
			var impls []string

			for _, e := range elem.ImplElements {
				impls = append(impls, e.Name)
			}

			tags = append(tags, "Implements: "+strings.Join(impls, ","))
		}
	} else if elem.FactoryName != "" {
		tags = append(tags, fmt.Sprintf("Factory: %s", elem.FactoryName))

		if len(elem.SubElements) > 0 {
			var group []string

			for _, e := range elem.SubElements {
				group = append(group, e.Name)
			}

			tags = append(tags, "Include: "+strings.Join(group, ","))
		}
	} else if elem.Singleton != nil {
		tags = append(tags, fmt.Sprintf("Type: %v", reflect.TypeOf(elem.Singleton)))
	}

	label := fmt.Sprintf("{%s}", strings.Join(tags, "|"))

	buf.WriteString(fmt.Sprintf("%s%s [shape=record, label=\"%s\"];\n", indent, elem.Name, label))
	return nil
}

func (pipeline *Pipeline) dumpDOTEdge(buf *bytes.Buffer, from, to *Element, indent string) error {
	buf.WriteString(fmt.Sprintf("%s%s -> %s;\n", indent, from.Name, to.Name))
	return nil
}
