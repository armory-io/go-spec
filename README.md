# Notice

This project is deprecated, use [Armory Go Commons](https://github.com/armory-io/go-commons)

# go-spec

`go-spec` aims to provide a framework for making our Go applications behave like Spinnaker's
Spring based applications from an operational standpoint. The goal is to provide the least amount
of boilerplate needed to bootstrap an application.

This framework provides developers with:
1. A preconfigured logger to use throughout your application
2. Metrics, exposed in the way we need.
3. A web server that uses a common TLS configuration.

Libraries used by this framework:
1. `gorilla/mux`
2. `go-eden/slf4go`
3. `armon/go-metrics`


## Components

### Logger

TODO

### Metrics

A universal metrics interface is supplied by `armon/go-metrics` and the framework handles surfacing them as necessary.

### Web Server

Applications using this framework will be supplied with a web server (provided by `armory/go-yaml-tools/server`) that
will be started and managed by the application. Users can configure the application's host/port via the following config
block:

```yaml
server:
  host: 0.0.0.0
  port: 3000
``` 

Any routes attached to the context's router will be instrumented with HTTP request metrics by default.

## Example

Example usage can be found in the `examples` directory.

## TODO

[] Settle on log package. Temporary one is in place for now.
[] Add more common, useful metrics
