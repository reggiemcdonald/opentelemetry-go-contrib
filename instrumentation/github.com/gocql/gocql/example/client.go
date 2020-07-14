// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// The Cassandra docker container conatains the
// "gocql-integration-example" keyspace and a single table
// with the following schema:
// gocql_integration_example.book
//   id UUID
//   title text
//   author_first_name text
//   author_last_name text
//   PRIMARY KEY(id)
// The example will insert fictional books into the database.

package main

import (
	"context"
	"github.com/gocql/gocql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"

	otelGocql "go.opentelemetry.io/contrib/github.com/gocql/gocql"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/exporters/metric/prometheus"
	zipkintrace "go.opentelemetry.io/otel/exporters/trace/zipkin"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var logger = log.New(os.Stderr, "zipkin-example", log.Ldate|log.Ltime|log.Llongfile)
var wg sync.WaitGroup

func initMetrics() {
	// Start prometheus
	metricExporter, err := prometheus.NewExportPipeline(prometheus.Config{})
	if err != nil {
		logger.Fatalf("failed to install metric exporter, %v", err)
	}
	server := http.Server{Addr: ":2222"}
	http.HandleFunc("/", metricExporter.ServeHTTP)
	go func() {
		defer wg.Done()
		wg.Add(1)
		log.Print(server.ListenAndServe())
	}()

	// ctrl+c will stop the server gracefully
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)
	go func() {
		<-shutdown
		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("problem shutting down server, %v", err)
		} else {
			log.Print("gracefully shutting down server")
		}
	}()

	otelGocql.InstrumentWithProvider(metricExporter.Provider())
}

func initTracer() {
	traceExporter, err := zipkintrace.NewExporter(
		"http://localhost:9411/api/v2/spans",
		"zipkin-example",
		zipkintrace.WithLogger(logger),
	)
	if err != nil {
		log.Fatalf("failed to create span traceExporter, %v", err)
	}

	provider, err := sdktrace.NewProvider(
		sdktrace.WithBatcher(
			traceExporter,
			sdktrace.WithBatchTimeout(5),
			sdktrace.WithMaxExportBatchSize(10),
		),
		sdktrace.WithConfig(sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
	)
	if err != nil {
		log.Fatalf("failed to create trace provider, %v", err)
	}

	global.SetTraceProvider(provider)
}

func getCluster() *gocql.ClusterConfig {
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = "gocql_integration_example"
	cluster.Consistency = gocql.LocalQuorum
	cluster.ProtoVersion = 3
	return cluster
}

func main() {
	initMetrics()
	initTracer()

	ctx, span := global.Tracer(
		"go.opentelemetry.io/contrib/github.com/gocql/gocql/example",
	).Start(context.Background(), "begin example")

	cluster := getCluster()
	// Create a session to begin making queries
	session, err := otelGocql.NewSessionWithTracing(
		cluster,
	)
	if err != nil {
		log.Fatalf("failed to create a session, %v", err)
	}
	defer session.Close()

	id := gocql.TimeUUID()
	if err := session.Query(
		"INSERT INTO book (id, title, author_first_name, author_last_name) VALUES (?, ?, ?, ?)",
		id,
		"Example Book 1",
		"firstName",
		"lastName",
	).WithContext(ctx).Exec(); err != nil {
		log.Printf("failed to insert data, %v", err)
	}

	res := session.Query(
		"SELECT title, author_first_name, author_last_name from book WHERE id = ?",
		id,
	).WithContext(ctx).Iter()

	var (
		title     string
		firstName string
		lastName  string
	)

	res.Scan(&title, &firstName, &lastName)

	log.Printf("Found Book {id: %s, title: %s, Name: %s, %s}", id, title, lastName, firstName)

	res.Close()

	if err = session.Query("DELETE FROM book WHERE id = ?", id).WithContext(ctx).Exec(); err != nil {
		log.Printf("failed to delete data, %v", err)
	}

	span.End()

	wg.Wait()
}
