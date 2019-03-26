package main

import (
	"bytes"
	"flag"
	"github.com/jcoene/gologger"
    "github.com/beberlei/statsd-librato-go/statsd"
    "github.com/beberlei/statsd-librato-go/backend"
	"net"
	"os"
	"regexp"
	"sort"
	"strconv"
	"time"
)

type Packet struct {
	Bucket   string
	Value    string
	Modifier string
	Sampling float32
}

var log *logger.Logger
var sanitizeRegexp = regexp.MustCompile("[^a-zA-Z0-9\\-_\\.:\\|@]")
var packetRegexp = regexp.MustCompile("([a-zA-Z0-9_\\.]+):(\\-?[0-9\\.]+)\\|(c|ms|g)(\\|@([0-9\\.]+))?")
var statsdBackend statsd.DataBackend;

var (
	serviceAddress = flag.String("address", "0.0.0.0:8125", "UDP service address")
    httpAddress    = flag.String("http", "127.0.0.1:8126", "HTTP Interface to metrics")
	libratoUser    = flag.String("user", "", "Librato Username")
	libratoToken   = flag.String("token", "", "Librato API Token")
	source         = flag.String("source", "", "Librato Source")
	url 		   = flag.String("url", "https://metrics-api.librato.com/v1/metrics", "Librato Url")
	flushInterval  = flag.Int64("flush", 60, "Flush Interval (seconds)")
	debug          = flag.Bool("debug", false, "Enable Debugging")
)

var (
	In       = make(chan Packet, 10000)
	counters = make(map[string]int64)
	timers   = make(map[string][]float64)
	gauges   = make(map[string]float64)
)

func monitor() {
	t := time.NewTicker(time.Duration(*flushInterval) * time.Second)

	for {
		select {
		case <-t.C:
			if err := submit(); err != nil {
				log.Error("submit: %s", err)
			}
		case s := <-In:
            log.Debug("Recieved packet %s=%s", s.Bucket, s.Value);

			if s.Modifier == "ms" {
				_, ok := timers[s.Bucket]
				if !ok {
					var t []float64
					timers[s.Bucket] = t
				}
				floatValue, _ := strconv.ParseFloat(s.Value, 64)
				timers[s.Bucket] = append(timers[s.Bucket], floatValue)
			} else if s.Modifier == "g" {
				_, ok := gauges[s.Bucket]
				if !ok {
					gauges[s.Bucket] = float64(0)
				}
				floatValue, _ := strconv.ParseFloat(s.Value, 64)
				gauges[s.Bucket] = floatValue
			} else {
				_, ok := counters[s.Bucket]
				if !ok {
					counters[s.Bucket] = 0
				}
				floatValue, _ := strconv.ParseFloat(s.Value, 32)
				counters[s.Bucket] += int64(float32(floatValue) * (1 / s.Sampling))
			}
		}
	}
}

func prepareMeasurements(counters map[string]int64, timers map[string][]float64, gauges map[string]float64) statsd.Measurement {
    m := statsd.Measurement{
        Source: *source,
        Counters: make([]statsd.Counter, 0),
        Gauges: make([]interface{}, 0),
    }

    for k, v := range counters {
        c := statsd.Counter{Name: k, Value: v}
        m.Counters = append(m.Counters, c)
    }

    for k, v := range gauges {
        g := statsd.SimpleGauge{Name: k, Value: v}
        m.Gauges = append(m.Gauges, g)
    }

    for k, t := range timers {
        g := statsd.ComplexGauge{Name: k, Count: len(t)}

        if g.Count > 0 {
            sort.Float64s(t)
            g.Min = t[0]
            g.Max = t[len(t)-1]
            for _, v := range t {
                g.Sum += v
                g.SumSquares += (v * v)
            }
        }

        m.Gauges = append(m.Gauges, g)
    }

    return m
}

func submit() (err error) {
	m := prepareMeasurements(counters, timers, gauges)

	if m.Count() == 0 {
		log.Info("no new measurements in the last %d seconds", *flushInterval)
		return
	}

    err = statsdBackend.Submit(m)

    if err != nil {
        log.Warn("error sending %d measurements: %s", m.Count(), err)
        return
    }

	log.Info("%d measurements sent", m.Count())

	for k, _ := range timers {
		delete(timers, k)
	}

    for k, _ := range counters {
        counters[k] = 0
    }

	return
}

func handle(conn *net.UDPConn, remaddr net.Addr, buf *bytes.Buffer) {
	var packet Packet
	var value string
	s := sanitizeRegexp.ReplaceAllString(buf.String(), "")

	for _, item := range packetRegexp.FindAllStringSubmatch(s, -1) {
		value = item[2]
		if item[3] == "ms" {
			_, err := strconv.ParseFloat(item[2], 32)
			if err != nil {
				value = "0"
			}
		}

		sampleRate, err := strconv.ParseFloat(item[5], 32)
		if err != nil {
			sampleRate = 1
		}

		packet.Bucket = item[1]
		packet.Value = value
		packet.Modifier = item[3]
		packet.Sampling = float32(sampleRate)

		In <- packet
	}
}

func listen() {
	address, _ := net.ResolveUDPAddr("udp", *serviceAddress)
	listener, err := net.ListenUDP("udp", address)
	defer listener.Close()
	if err != nil {
		log.Fatal("unable to listen: %s", err)
		os.Exit(1)
	}

	log.Info("listening for events...")

	for {
		message := make([]byte, 512)
		n, remaddr, error := listener.ReadFrom(message)
		if error != nil {
			continue
		}
		buf := bytes.NewBuffer(message[0:n])
		go handle(listener, remaddr, buf)
	}
}

func main() {
	flag.Parse()

	if *debug {
		log = logger.NewLogger(logger.LOG_LEVEL_DEBUG, "statsd")
	} else {
		log = logger.NewLogger(logger.LOG_LEVEL_INFO, "statsd")
	}

    if (*libratoUser == "" || *libratoToken == "") {
        b := new(backend.Http)
        b.Address = *httpAddress
        statsdBackend = b
    } else {
        b := new(backend.Librato)
        b.User = *libratoUser
        b.Token = *libratoToken
        b.Url = *url
        statsdBackend = b
    }

    statsdBackend.Init(log)

	go listen()
	monitor()
}
