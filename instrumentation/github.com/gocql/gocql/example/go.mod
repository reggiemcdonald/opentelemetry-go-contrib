module go.opentelemetry.io/contrib/github.com/gocql/gocql/example

go 1.14

require (
	github.com/gocql/gocql v0.0.0-20200624222514-34081eda590e
	go.opentelemetry.io/contrib/github.com/gocql/gocql v0.0.0
	go.opentelemetry.io/otel v0.8.0
	go.opentelemetry.io/otel/exporters/metric/prometheus v0.8.0
	go.opentelemetry.io/otel/exporters/trace/zipkin v0.8.0
)

// TODO: remove
replace go.opentelemetry.io/contrib/github.com/gocql/gocql => ../
