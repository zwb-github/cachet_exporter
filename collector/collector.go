package collector

import (
	"sync"
	"time"

	"github.com/andygrunwald/cachet"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const namespace = "cachet"

type cachetCollector struct {
	mutex  sync.RWMutex
	client *cachet.Client

	up             *prometheus.Desc
	scrapeDuration *prometheus.Desc
}

// NewCachetCollector returns a prometheus collector which exports
// metrics from a Cachet status page.
func NewCachetCollector(apiURL string) prometheus.Collector {
	client, err := cachet.NewClient(apiURL, nil)
	if err != nil {
		log.With("apiURL", apiURL).Fatal("Failed to create a new Cachet client")
	}

	return &cachetCollector{
		client: client,
		up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "up"),
			"Cachet API is up and accepting requests",
			nil,
			nil,
		),
		scrapeDuration: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "scrape_duration_seconds"),
			"Time Cachet scrape took in seconds",
			nil,
			nil,
		),
	}
}

// Describe describes all the metrics exported by the Cachet exporter.
// It implements prometheus.Collector.
func (c *cachetCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up
	ch <- c.scrapeDuration
}

// Collect fetches the metrics data from the Cachet application and
// delivers them as Prometheus metrics. It implements prometheus.Collector.
func (c *cachetCollector) Collect(ch chan<- prometheus.Metric) {
	// To protect metrics from concurrent collects.
	c.mutex.Lock()
	defer c.mutex.Unlock()

	start := time.Now()
	log.Info("Collecting metrics from Cachet")
	_, _, err := c.client.General.Ping()
	if err != nil {
		ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 0)
		log.With("error", err).Error("Failed to scrape Cachet")
		return
	}

	ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 1)
	ch <- prometheus.MustNewConstMetric(c.scrapeDuration, prometheus.GaugeValue, time.Since(start).Seconds())
}
