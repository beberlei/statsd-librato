# StatsD Server for Librato

This is an implementation of Etsy's StatsD written in Go that submits data to
Librato Metrics or to a local HTTP server.

This was forked from
[jbuchbinder/statsd-go](https://github.com/jbuchbinder/statsd-go) and altered
to provide support for Librato as a submission backend.  Besides the Librato
backend it now also contains a backend for locally displaying the data
from a HTTP server. This is helpful for development environments that make use
of statsd.

# Usage

```
Usage of statsd:
  -address="0.0.0.0:8125": UDP service address
  -debug=false: Enable Debugging
  -flush=30: Flush Interval (seconds)
  -token="": Librato API Token
  -user="": Librato Username
  -httpAddress="127.0.0.1:8126": Used for starting a HTTP server in local mode
```

# Local HTTP Server for StatsD

![Local StatsD](https://raw.githubusercontent.com/beberlei/statsd-librato/master/statsd-local-http.gif)

## License

MIT License, see LICENSE for details.
