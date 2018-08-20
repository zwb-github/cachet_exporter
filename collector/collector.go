package collector

import (
	"strconv"
	"sync"
	"time"

	"github.com/andygrunwald/cachet"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const namespace = "cachet"

type metricConfig struct {
	Name   string            `yaml:"metric_name"`
	Labels map[string]string `yaml:"labels"`
}

// Config is the struct used to load the configurations from yaml file
type Config struct {
	Metrics []metricConfig `yaml:"metrics"`
}

type cachetCollector struct {
	mutex  sync.RWMutex
	client *cachet.Client
	config Config

	up                   *prometheus.Desc
	scrapeDuration       *prometheus.Desc
	cachetIncidentsTotal *prometheus.Desc
}

// NewCachetCollector returns a prometheus collector which exports
// metrics from a Cachet status page.
func NewCachetCollector(apiURL string, config Config) prometheus.Collector {
	client, err := cachet.NewClient(apiURL, nil)
	if err != nil {
		log.With("apiURL", apiURL).Fatal("Failed to create a new Cachet client")
	}

	return &cachetCollector{
		config: config,
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
		cachetIncidentsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "cachet_incidents_total"),
			"Blah",
			[]string{"status", "group_name", "component_name"},
			nil,
		),
	}
}

// Describe describes all the metrics exported by the Cachet exporter.
// It implements prometheus.Collector.
func (c *cachetCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up
	ch <- c.scrapeDuration
	ch <- c.cachetIncidentsTotal
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

	groups, _, err := c.client.ComponentGroups.GetAll(&cachet.ComponentGroupsQueryParams{})

	if err != nil {
		log.With("error", err).Error("Failed to scrape Gropu Components")
	}

	incidents := map[int][]cachet.Incident{
		1: getIncidentsByStatus(c, 1),
		2: getIncidentsByStatus(c, 2),
		3: getIncidentsByStatus(c, 3),
	}
	for _, group := range groups.ComponentGroups {
		createIncidentsTotalMetric(c, group, incidents, ch)
	}

	ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 1)
	ch <- prometheus.MustNewConstMetric(c.scrapeDuration, prometheus.GaugeValue, time.Since(start).Seconds())
}

func getIncidentsByStatus(c *cachetCollector, status int) []cachet.Incident {
	incidents, _, err := c.client.Incidents.GetAll(&cachet.IncidentsQueryParams{Status: 1})
	if err != nil {
		log.With("error", err).Error("Failed to scrape Gropu Components")
	}
	return incidents.Incidents
}

func createIncidentsTotalMetric(c *cachetCollector, group cachet.ComponentGroup, incidents map[int][]cachet.Incident, ch chan<- prometheus.Metric) {
	for _, component := range group.EnabledComponents {
		createIncidentsTotalMetricByComponent(c, group, component, incidents, ch)
	}
}

func createIncidentsTotalMetricByComponent(c *cachetCollector, group cachet.ComponentGroup, component cachet.Component, incidents map[int][]cachet.Incident, ch chan<- prometheus.Metric) {
	for status, allIncidents := range incidents {
		componentIncidents := make([]cachet.Incident, 0)
		for _, incident := range allIncidents {
			if incident.ComponentID == component.ID && status == incident.Status {
				componentIncidents = append(componentIncidents, incident)
			}
		}

		ch <- prometheus.MustNewConstMetric(c.cachetIncidentsTotal, prometheus.GaugeValue, float64(len(componentIncidents)), strconv.Itoa(status), group.Name, component.Name)
	}

}
