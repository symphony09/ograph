package ogcore

import "context"

type VirtualNode struct {
	VirtualName string
	DefaultImpl string
	Implements  map[string]Node
}

func (vn *VirtualNode) Name() string {
	return vn.VirtualName
}

func (vn *VirtualNode) Run(ctx context.Context, state State) error {
	if len(vn.Implements) == 0 {
		return nil
	}

	name := GetImplNodeName(ctx, vn.Name())

	if name == "" {
		name = vn.DefaultImpl
	}

	if node := vn.Implements[name]; node != nil {
		return node.Run(ctx, state)
	}

	return nil
}

func GetImplNodeName(ctx context.Context, vname string) string {
	if v := ctx.Value(ImplQueryKey(vname)); v != nil {
		if nodeName, ok := v.(string); ok {
			return nodeName
		}
	}

	return ""
}
