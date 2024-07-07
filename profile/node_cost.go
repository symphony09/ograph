package profile

import (
	"sort"
	"time"

	"github.com/symphony09/ograph/ogcore"
)

type NodeCost struct {
	NodeName      string
	TotalCost     time.Duration
	SelfCost      time.Duration
	BeforeRunCost time.Duration
	AfterRunCost  time.Duration
}

func GetNodeCostMap(traceData []ogcore.EventTrace) map[string]NodeCost {
	result := make(map[string]NodeCost)

	nodeCostList := ListNodeCost(traceData, -1)

	for i := range nodeCostList {
		result[nodeCostList[i].NodeName] = nodeCostList[i]
	}

	return result
}

type NodeCostList []NodeCost

func (a NodeCostList) Len() int           { return len(a) }
func (a NodeCostList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a NodeCostList) Less(i, j int) bool { return a[i].TotalCost > a[j].TotalCost }

func ListNodeCost(traceData []ogcore.EventTrace, limitN int) []NodeCost {
	var result NodeCostList
	nodeTsGroup := make(map[string]*[4]*time.Time)

	for i, trace := range traceData {
		tsGroup := nodeTsGroup[trace.NodeName]
		if tsGroup == nil {
			tsGroup = new([4]*time.Time)
			nodeTsGroup[trace.NodeName] = tsGroup
		}

		ts := &traceData[i].Timestamp

		switch trace.Event {
		case "ready":
			tsGroup[0] = ts
		case "start":
			tsGroup[1] = ts
		case "end":
			tsGroup[2] = ts
		case "complete":
			tsGroup[3] = ts
		}
	}

	for nodeName, tsGroup := range nodeTsGroup {
		var nodeCost NodeCost
		nodeCost.NodeName = nodeName
		if tsGroup[0] != nil && tsGroup[3] != nil {
			nodeCost.TotalCost = (*tsGroup[3]).Sub(*tsGroup[0])
		}
		if tsGroup[1] != nil && tsGroup[2] != nil {
			nodeCost.SelfCost = (*tsGroup[2]).Sub(*tsGroup[1])
		}
		if tsGroup[0] != nil && tsGroup[1] != nil {
			nodeCost.BeforeRunCost = (*tsGroup[1]).Sub(*tsGroup[0])
		}
		if tsGroup[2] != nil && tsGroup[3] != nil {
			nodeCost.AfterRunCost = (*tsGroup[3]).Sub(*tsGroup[2])
		}

		result = append(result, nodeCost)
	}

	sort.Sort(result)

	if limitN > 0 && limitN < len(result) {
		result = result[:limitN-1]
	}

	return result
}
