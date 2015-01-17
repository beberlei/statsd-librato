package backend

import (
    "encoding/json"
    "github.com/jcoene/gologger"
    "github.com/beberlei/statsd-librato-go/statsd"
    "net/http"
    "strings"
    "reflect"
)

type Http struct {
    Address string
    staticHandler *http.ServeMux
    lastMeasurement statsd.Measurement
    log *logger.Logger
    Values map[string]float64
}

func (s *Http) urlHandler(w http.ResponseWriter, r *http.Request) {
    if (strings.Contains(r.URL.Path, "/data.json")) {
        payload, err := json.MarshalIndent(s.Values, "", "  ")

        if (err == nil) {
            w.Write(payload)
        } else {
            s.log.Warn("Unable to marshal last measurement: %s", err)
        }
    } else {
        s.staticHandler.ServeHTTP(w, r)
    }   
}

func (s *Http) serve() (err error) {
    s.staticHandler.Handle("/", http.FileServer(http.Dir("./htdocs")))

    http.HandleFunc("/", s.urlHandler)
    return http.ListenAndServe(s.Address, nil)
}

func (s *Http) Init(log *logger.Logger) (err error) {
    s.staticHandler = http.NewServeMux()
    s.log = log
    s.Values = make(map[string]float64)
    go s.serve()
    return nil
}

func (s *Http) Submit(m statsd.Measurement) (err error) {
    s.lastMeasurement = m

    for _, counter := range m.Counters {
        if _, ok := s.Values[counter.Name]; ok {
            s.Values[counter.Name] += (float64)(counter.Value)
        } else {
            s.Values[counter.Name] = (float64)(counter.Value)
        }
    }

    for _, gauge := range m.Gauges {
        name := reflect.ValueOf(gauge).FieldByName("Name")
        value := reflect.ValueOf(gauge).FieldByName("Value")
        count := reflect.ValueOf(gauge).FieldByName("Count")
        sum   := reflect.ValueOf(gauge).FieldByName("Sum")

        if (value.IsValid()) {
            s.Values[name.String()] = value.Float();
        } else if (count.IsValid()) {
            s.Values[name.String()] = (sum.Float() / (float64)(count.Int()))
        }
    }

    return
}
