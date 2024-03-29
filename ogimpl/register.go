package ogimpl

import (
	"github.com/symphony09/ograph/global"
)

func init() {
	global.Factories.Add(Choose, ChooseClusterFactory)
	global.Factories.Add(Parallel, ParallelClusterFactory)
	global.Factories.Add(Race, RaceClusterFactory)

	global.Factories.Add(Async, AsyncWrapperFactory)
	global.Factories.Add(Condition, ConditionWrapperFactory)
	global.Factories.Add(Loop, LoopWrapperFactory)
	global.Factories.Add(Retry, RetryWrapperFactory)
	global.Factories.Add(Slient, SlientWrapperFactory)
	global.Factories.Add(Timeout, TimeoutWrapperFactory)
	global.Factories.Add(Trace, TraceWrapperFactory)
}
