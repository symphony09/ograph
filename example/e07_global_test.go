package example

import (
	"context"
	"testing"

	"github.com/symphony09/ograph"
	"github.com/symphony09/ograph/global"
)

var XFactory = "X"

func init() {
	global.Factories.Add(XFactory, NewProductionX)
}

func TestGlobalFactory(t *testing.T) {
	pipeline := ograph.NewPipeline()

	pipeline.Register(ograph.NewElement("x1").UseFactory(XFactory))

	if err := pipeline.Run(context.TODO(), nil); err != nil {
		t.Error(err)
	}
}
