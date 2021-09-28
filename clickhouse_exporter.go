package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/ClickHouse/clickhouse_exporter/exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/log"
	"gopkg.in/yaml.v2"
)

var (
	listeningAddress    = flag.String("telemetry.address", ":9116", "Address on which to expose metrics.")
	metricsEndpoint     = flag.String("telemetry.endpoint", "/metrics", "Path under which to expose metrics.")
	clickhouseScrapeURI = flag.String("scrape_uri", "http://localhost:8123/", "URI to clickhouse http endpoint")
	clickhouseOnly      = flag.Bool("clickhouse_only", true, "Expose only Clickhouse metrics, not metrics from the exporter itself")
	insecure            = flag.Bool("insecure", true, "Ignore server certificate if using https")
	user                = os.Getenv("CLICKHOUSE_USER")
	password            = os.Getenv("CLICKHOUSE_PASSWORD")
)

func main() {
	flag.Parse()
	log.Println("clickhouse metrics exporter for Prometheus monitoring")
	http.HandleFunc(*metricsEndpoint, handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Clickhouse Exporter</title></head>
			<body>
			<h1>Clickhouse Exporter</h1>
			<p><a href="` + *metricsEndpoint + `">Metrics</a></p>
			</body>
			</html>`))
	})

	log.Fatal(http.ListenAndServe(*listeningAddress, nil))
}

func handler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		target := query.Get("target")
		module := query.Get("module")
		conf, _ := loadConfig()

		if target == "" {
			buf := "uri error, not found target"
			w.Write([]byte(buf))
			return
		}

		if module == "" {
			buf := "uri error, not found module"
			w.Write([]byte(buf))
			return
		}

		var user string
		var password string
		var isBreak bool
		for i:=0; i < len(conf.Modules); i++ {
			if conf.Modules[i].Name == module {
				user = conf.Modules[i].User
				password = conf.Modules[i].Password
				isBreak = true
			}
		}
		if ! isBreak {
			buf := "not found module in conf.yml"
			w.Write([]byte(buf))
			return
		}

		*clickhouseScrapeURI = fmt.Sprintf("http://%v", target)

		uri, err := url.Parse(*clickhouseScrapeURI)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Scraping %s", *clickhouseScrapeURI)

		exporter := exporter.NewExporter(*uri, *insecure, user, password)
		registerer := prometheus.NewRegistry()
		registerer.MustRegister(exporter)

		gatherers := prometheus.Gatherers{}
		gatherers = append(gatherers, registerer)
		if ! *clickhouseOnly {
			gatherers = append(gatherers, prometheus.DefaultGatherer)
		}

		// Delegate http serving to Prometheus client library, which will call collector.Collect.
		h := promhttp.HandlerFor(gatherers, promhttp.HandlerOpts{
			ErrorHandling: promhttp.ContinueOnError,
		})

		h.ServeHTTP(w, r)
	}
}

func loadConfig() (*Config, error)  {
	path,_ := os.Getwd()
	path = filepath.Join(path, "conf/conf.yml")
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.New("read conf.yml fail, path: " + path)
	}
	conf := new(Config)
	err = yaml.Unmarshal(data, conf)
	if err != nil {
		return nil, errors.New("unmarshal conf.yml fail")
	}
	return conf, err
}
