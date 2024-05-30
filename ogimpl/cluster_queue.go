package ogimpl

import (
	"github.com/symphony09/ograph"
	"github.com/symphony09/ograph/ogcore"
)

var QueueClusterFactory = func() ogcore.Node {
	return &ograph.BaseCluster{}
}
