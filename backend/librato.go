package backend

import (
    "bytes"
    "errors"
    "encoding/json"
    "fmt"
    "github.com/beberlei/statsd-librato-go/statsd"
    "net/http"
    "io/ioutil"
)

type Librato struct {
    User string
    Token string
}

func (s Librato) Submit(m *statsd.Measurement) (err error) {
    payload, err := json.MarshalIndent(m, "", "  ")

    if err != nil {
        return err
    }

    req, err := http.NewRequest("POST", "https://metrics-api.librato.com/v1/metrics", bytes.NewBuffer(payload))
    if err != nil {
        return err
    }

    req.Header.Add("Content-Type", "application/json")
    req.Header.Set("User-Agent", "statsd/1.0")
    req.SetBasicAuth(s.User, s.Token)
    resp, err := http.DefaultClient.Do(req)
    if err == nil && resp.StatusCode != 200 {
        if err == nil {
            raw, _ := ioutil.ReadAll(resp.Body)
            err = errors.New(fmt.Sprintf("%s: %s", resp.Status, string(raw)))
        }
        return err
    }

    return nil
}
