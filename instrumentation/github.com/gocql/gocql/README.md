## `go.opentelemetry.io/contrib/instrumentation/github.com/gocql/gocql`

This package provides tracing and metrics to the golang cassandra client `github.com/gocql/gocql` using the `ConnectObserver`, `QueryObserver` and `BatchObserver` interfaces. 

To enable tracing in your application: 

```go
package main

import (
        "context"
        
        "github.com/gocql/gocql"
        otelGocql "go.opentelemetry.io/contrib/github.com/gocql/gocql"
)

func main() {
        // Create a cluster
        host := "localhost"
        cluster := gocql.NewCluster(host)

        // Create a session with tracing
        session, err := otelGocql.NewSessionWithTracing(
                context.Background(),
                cluster,
                // Include any options here
        )

        // Begin using the session
        
}
```

You can customize the tracing session by passing any of the following options to `NewSessionWithTracing`:

| Function | Description |
| -------- | ----------- |
| `WithQueryObserver(gocql.QueryObserver)` | Specify an additional QueryObserver to be called. |
| `WithBatchObserver(gocql.BatchObserver)` | Specify an additional BatchObserver to be called. |
| `WithConnectObserver(gocql.ConnectObserver)` | Specify an additional ConnectObserver to be called. |
| `WithTracer(trace.Tracer)` | The tracer to be used to create spans for the gocql session. If not specified, `global.Tracer("github.com/gocql/gocql")` will be used. |
| `WithQueryInstrumentation(bool)` | To enable/disable tracing and metrics for queries. |
| `WithBatchInstrumentation(bool)` | To enable/disable tracing and metrics for batch queries. |
| `WithConnectInstrumentation(bool)` | To enable/disable tracing and metrics for new connections. |

In addition to using the convenience function, you may also manually set the obsevers: 

```go
host := "localhost"
cluster := gocql.NewCluster(host)
cluster.QueryObserver = otelGocql.NewQueryObserver(nil, &OtelConfig{
        tracer: global.Tracer("github.com/gocql/gocql"),
})
session, err := cluster.CreateSession()
```