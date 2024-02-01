package event

// NewEvent represents the data structure for an event that is not yet created.
type NewEvent struct {
	NewAttributes `json:"attributes"`
}

// ExistingProfile represents the data structure for an existing profile for event structure.
type ExistingProfile struct {
	ID string `json:"id"`
}

// ExistingEvent represents the data structure for an existing event.
type ExistingEvent struct {
	ID         string `json:"id"`
	EventType  string `json:"type"`
	Attributes Attributes
}

// NewAttributes represents the data structure for an attributes of event that is not yet created.
type NewAttributes struct {
	Time       string            `json:"time"`
	Value      float64           `json:"value"`
	Properties map[string]string `json:"properties"`
	Profile    interface{}       `json:"profile"`
	Metric     interface{}       `json:"metric"`
}

// Attributes represents the data structure for an existing attributes.
type Attributes struct {
	Timestamp       int64                  `json:"timestamp"`
	Datetime        string                 `json:"Datetime"`
	UUID            string                 `json:"uuid"`
	EventProperties map[string]interface{} `json:"event_properties"`
}

// NewMetric represents the data structure for a metric that is not yet created.
type NewMetric struct {
	Attributes MetricAttributes `json:"attributes"`
}

// MetricAttributes represents the data structure for a metric attributes.
type MetricAttributes struct {
	Name string `json:"name"`
}
