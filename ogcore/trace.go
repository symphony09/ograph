package ogcore

import "time"

type Tracker struct {
	TraceData []EventTrace
}

type EventTrace struct {
	NodeName  string
	Event     string
	Timestamp time.Time
}

func (tracker *Tracker) Record(nodeName string, event string, timestamp time.Time) {
	if tracker == nil {
		return
	}

	tracker.TraceData = append(tracker.TraceData, EventTrace{
		NodeName:  nodeName,
		Event:     event,
		Timestamp: timestamp,
	})
}
