package backend

import (
    "encoding/json"
    "fmt"
    "github.com/jcoene/gologger"
    "github.com/beberlei/statsd-librato-go/statsd"
    "net/http"
    "strings"
)

type Http struct {
    Address string
    staticHandler *http.ServeMux
    lastMeasurement statsd.Measurement
    log *logger.Logger
}

func (s *Http) urlHandler(w http.ResponseWriter, r *http.Request) {
    if (strings.Contains(r.URL.Path, "/data.json")) {
        s.log.Debug("Obj %+v", s.lastMeasurement)
        payload, err := json.MarshalIndent(s.lastMeasurement, "", "  ")

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
    go s.serve()
    return nil
}

func (s *Http) Submit(m statsd.Measurement) (err error) {
    fmt.Print("Have measurements %+v", s.lastMeasurement)
    fmt.Print("Receiving measurements %+v", m)
    s.lastMeasurement = m
    return
}
