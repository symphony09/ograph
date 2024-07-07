package internal

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/symphony09/ograph/ogcore"
	"golang.org/x/sync/errgroup"
)

type Worker struct {
	graph *Graph[ogcore.Node]
}

type WorkParams struct {
	GorLimit         int
	ActionsBeforeRun map[string]ogcore.Action
	ActionsAfterRun  map[string]ogcore.Action
	Tracker          *ogcore.Tracker
}

func (worker *Worker) Work(ctx context.Context, state ogcore.State, params *WorkParams) error {
	todoCh, doneCh := worker.graph.Scheduling()
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
			if ctx.Err() != nil {
				return ctx.Err()
			}

			currentWorkName = work.Name
			node := work.Elem

			tracker := params.Tracker
			tracker.Record(currentWorkName, "ready", time.Now())

			if params.ActionsBeforeRun != nil {
				if action := params.ActionsBeforeRun[work.Name]; action != nil {
					if err := action(ctx, state); err != nil {
						return fmt.Errorf("%s failed, error: %w", work.Name, err)
					}
				}
			}

			tracker.Record(currentWorkName, "start", time.Now())

			if node != nil {
				if err := node.Run(ctx, state); err != nil {
					return fmt.Errorf("%s failed, error: %w", work.Name, err)
				}
			}

			tracker.Record(currentWorkName, "end", time.Now())

			if params.ActionsAfterRun != nil {
				if action := params.ActionsAfterRun[work.Name]; action != nil {
					if err := action(ctx, state); err != nil {
						return fmt.Errorf("%s failed, error: %w", work.Name, err)
					}
				}
			}

			tracker.Record(currentWorkName, "complete", time.Now())
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

func NewWorker(graph *Graph[ogcore.Node]) *Worker {
	return &Worker{
		graph: graph,
	}
}
