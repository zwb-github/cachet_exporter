package main

import (
	"fmt"
	"net/http"

	"github.com/ContaAzul/cachet_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	version       = "dev"
	listenAddress = kingpin.Flag("web.listen-address", "Address to bind the server").Default(":9470").String()
	metricsPath   = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
	apiURL        = kingpin.Flag("cachet.api-url", "Cachet API URL").OverrideDefaultFromEnvar("CACHET_API_URL").String()
)

func main() {
	kingpin.Version(version)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	log.Info("Starting cachet_exporter ", version)

	if *apiURL == "" {
		log.Fatal("You must provide your Cachet API URL")
	}

	prometheus.MustRegister(collector.NewCachetCollector(*apiURL))

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w,
			`
			<html>
			<head><title>Cachet Exporter</title></head>
			<body>
				<h1>Cachet Exporter</h1>
				<p><a href="`+*metricsPath+`">Metrics</a></p>
			</body>
			</html>
			`)
	})

	log.Infof("Server listening on %s", *listenAddress)
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
