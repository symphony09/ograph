package internal

import (
	"context"
	"fmt"
	"iter"
	"runtime"
	"sync"
	"time"

	"github.com/symphony09/ograph/ogcore"
	"golang.org/x/sync/errgroup"
)

type Worker struct {
	graph *Graph[ogcore.Node]
}

type WorkParams struct {
	GorLimit   int
	Tracker    *ogcore.Tracker
	Interrupts iter.Seq[string]

	Pause        bool
	ContinueCond *sync.Cond
}

func (worker *Worker) Work(ctx context.Context, state ogcore.State, params *WorkParams) error {
	tracker := params.Tracker

	// opt for graph that can be fully serialized
	if worker.graph.ScheduleNum == 1 {
		headNode := worker.graph.Heads[0]
		var works []*GraphVertex[ogcore.Node]

		if headNode.Group != nil {
			works = headNode.Group
		} else {
			works = append(works, headNode)
		}

		for _, work := range works {
			if params.ContinueCond != nil {
				waitContinue(params)
			}

			if ctx.Err() != nil {
				return ctx.Err()
			}

			currentWorkName := work.Name
			node := work.Elem

			if tracker != nil {
				tracker.Record(currentWorkName, "ready", time.Now())
			}

			if tracker != nil {
				tracker.Record(currentWorkName, "start", time.Now())
			}

			if node != nil {
				if err := node.Run(ctx, state); err != nil {
					return fmt.Errorf("%s failed, error: %w", work.Name, err)
				}
			}

			if tracker != nil {
				tracker.Record(currentWorkName, "end", time.Now())
			}

			if tracker != nil {
				tracker.Record(currentWorkName, "complete", time.Now())
			}
		}
		return nil
	}

	// schedule as normal
	todoCh, doneCh := worker.graph.Scheduling(params.Interrupts, params.GorLimit)
	defer close(doneCh)

	doWorks := func(works []*GraphVertex[ogcore.Node]) (err error) {
		var currentWorkName string

		defer func() {
			if info := recover(); info != nil {
				err = fmt.Errorf("worker panic on %s, info: %v", currentWorkName, info)
			}

			doneCh <- works
		}()

		for _, work := range works {
			if params.ContinueCond != nil {
				waitContinue(params)
			}

			if ctx.Err() != nil {
				return ctx.Err()
			}

			currentWorkName = work.Name
			node := work.Elem

			if tracker != nil {
				tracker.Record(currentWorkName, "ready", time.Now())
			}

			if tracker != nil {
				tracker.Record(currentWorkName, "start", time.Now())
			}

			if node != nil {
				if err := node.Run(ctx, state); err != nil {
					return fmt.Errorf("%s failed, error: %w", work.Name, err)
				}
			}

			if tracker != nil {
				tracker.Record(currentWorkName, "end", time.Now())
			}

			if tracker != nil {
				tracker.Record(currentWorkName, "complete", time.Now())
			}
		}

		return nil
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(params.GorLimit)

	pn := runtime.GOMAXPROCS(0)

	if params.GorLimit > 0 && params.GorLimit <= pn {
		for i := 0; i < params.GorLimit; i++ {
			g.Go(func() (err error) {
				for works := range todoCh {
					if err := doWorks(works); err != nil {
						return err
					}
				}

				return nil
			})
		}
	} else {
		if worker.graph.ScheduleNum > pn {
			for i := 0; i < pn; i++ {
				g.Go(func() (err error) {
					for works := range todoCh {
						if err := doWorks(works); err != nil {
							return err
						}
					}

					return nil
				})
			}
		}

		for works := range todoCh {
			works := works // https://golang.org/doc/faq#closures_and_goroutines

			g.Go(func() (err error) {
				return doWorks(works)
			})
		}
	}

	err := g.Wait()

	return err
}

func waitContinue(params *WorkParams) {
	params.ContinueCond.L.Lock()

	for params.Pause {
		params.ContinueCond.Wait()
	}

	params.ContinueCond.L.Unlock()
}

func NewWorker(graph *Graph[ogcore.Node]) *Worker {
	return &Worker{
		graph: graph,
	}
}
