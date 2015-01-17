package statsd

type Measurement struct {
    Counters []Counter     `json:"counters"`
    Gauges   []interface{} `json:"gauges"`
    Source   string        `json:"source"`
}

func (m *Measurement) Count() int {
    return (len(m.Counters) + len(m.Gauges))
}

type Counter struct {
    Name  string `json:"name"`
    Value int64  `json:"value"`
}

type SimpleGauge struct {
    Name  string  `json:"name"`
    Value float64 `json:"value"`
}

type ComplexGauge struct {
    Name       string  `json:"name"`
    Count      int     `json:"count"`
    Sum        float64 `json:"sum"`
    Min        float64 `json:"min"`
    Max        float64 `json:"max"`
    SumSquares float64 `json:"sum_squares"`
}

type DataBackend interface {
    Submit(m *Measurement) (err error)
}
