package internal

import (
	"context"
	"fmt"
	"iter"
	"runtime"
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
}

func (worker *Worker) Work(ctx context.Context, state ogcore.State, params *WorkParams) error {
	todoCh, doneCh := worker.graph.Scheduling()
	defer close(doneCh)

	var nextInterrupt func() (string, bool)
	var stopInterrupt func()
	var interruptAt string
	var doInterrupt bool

	if params.Interrupts != nil {
		nextInterrupt, stopInterrupt = iter.Pull(params.Interrupts)
		defer stopInterrupt()

		interruptAt, doInterrupt = nextInterrupt()
	}

	tracker := params.Tracker

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

			if tracker != nil {
				tracker.Record(currentWorkName, "ready", time.Now())
			}

			if nextInterrupt != nil {
				if doInterrupt && interruptAt == currentWorkName+":start" {
					interruptAt, doInterrupt = nextInterrupt()
				}
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

			if nextInterrupt != nil {
				if doInterrupt && interruptAt == work.Name+":end" {
					interruptAt, doInterrupt = nextInterrupt()
				}
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

func NewWorker(graph *Graph[ogcore.Node]) *Worker {
	return &Worker{
		graph: graph,
	}
}
