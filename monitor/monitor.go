package monitor

var (
	registry = make([]MonitorDataCollector, 0)
)

func Register(collector ...MonitorDataCollector) {
	registry = append(registry, collector...)
}

func GetMonitorData() []MetricData {
	metrics := make([]MetricData, 0)
	for _, collector := range registry {
		metrics = append(metrics, collector.Monitor()...)
	}
	return metrics
}

type ILogger interface {
	Errorf(template string, args ...interface{})
}

type MonitorDataCollector interface {
	Monitor() []MetricData
}

type Type string

const (
	Counter Type = "Counter"
	Gauge   Type = "Gauge"
)

type MetricData struct {
	Type   Type
	Metric string
	Labels map[string]string
	Data   float64
}

func NewMetricData(Type Type, metric string, labels map[string]string, data float64) MetricData {
	return MetricData{Type: Type, Metric: metric, Labels: labels, Data: data}
}
func NewGaugeMetricData(metric string, labels map[string]string, data float64) MetricData {
	return MetricData{Type: Gauge, Metric: metric, Labels: labels, Data: data}
}
func NewCounterMetricData(metric string, labels map[string]string, data float64) MetricData {
	return MetricData{Type: Counter, Metric: metric, Labels: labels, Data: data}
}
