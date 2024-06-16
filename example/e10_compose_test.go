package example

import (
	"context"
	"fmt"
	"testing"

	"github.com/symphony09/ograph"
)

func TestComposePipeline(t *testing.T) {
	study := newSubPipeline("LearnProgramming", "LearnEnglish")
	relax := newSubPipeline("PlayGame", "Sleep")

	studyThings := ograph.NewElement("StudyThings").UseNode(study)
	relaxThings := ograph.NewElement("RelaxThings").UseNode(relax)

	day := ograph.NewPipeline()
	day.Register(studyThings).Register(relaxThings, ograph.Rely(studyThings))

	if err := day.Run(context.TODO(), nil); err != nil {
		t.Error(err)
	}
}

func newSubPipeline(things ...string) *ograph.Pipeline {
	pipeline := ograph.NewPipeline()

	for _, thing := range things {
		thing := thing
		pipeline.Register(ograph.NewElement(thing).UseFn(func() error {
			fmt.Printf("->%s", thing)
			return nil
		}))
	}

	return pipeline
}
