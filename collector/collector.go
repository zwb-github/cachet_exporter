package collector

import (
	"sync"
	"time"

	"github.com/ContaAzul/cachet_exporter/client"
	"github.com/andygrunwald/cachet"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const namespace = "cachet"

var incidentStatus = map[int]string{
	0: "Scheduled",
	1: "Investigating",
	2: "Identified",
	3: "Watching",
	4: "Fixed",
}

var componentStatus = map[int]string{
	0: "Unknown",
	1: "Operational",
	2: "Performance Issues",
	3: "Partial Outage",
	4: "Major Outage",
}

type cachetCollector struct {
	mutex  sync.RWMutex
	client client.Client

	up             *prometheus.Desc
	scrapeDuration *prometheus.Desc
	incidents      *prometheus.Desc
	components     *prometheus.Desc
}

// NewCachetCollector returns a prometheus collector which exports
// metrics from a Cachet status page.
func NewCachetCollector(client client.Client) prometheus.Collector {
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
		incidents: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "incidents"),
			"Number of incidents by status",
			[]string{"status", "group_name", "component_name"},
			nil,
		),
		components: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "components"),
			"Number of components by status",
			[]string{"status", "group_name"},
			nil,
		),
	}
}

// Describe describes all the metrics exported by the Cachet exporter.
// It implements prometheus.Collector.
func (c *cachetCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up
	ch <- c.scrapeDuration
	ch <- c.incidents
	ch <- c.components
}

// Collect fetches the metrics data from the Cachet application and
// delivers them as Prometheus metrics. It implements prometheus.Collector.
func (c *cachetCollector) Collect(ch chan<- prometheus.Metric) {
	// To protect metrics from concurrent collects.
	c.mutex.Lock()
	defer c.mutex.Unlock()

	start := time.Now()
	log.Info("Collecting metrics from Cachet")

	groups, err := c.client.GetAllComponentGroups()
	if err != nil {
		// If fails to get all components groups all metrics will be wrong,
		// in this case it's better to stop here
		ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 0)
		log.With("error", err).Error("failed to scrape Group Components")
		return
	}

	up := 1
	incidents := map[int][]cachet.Incident{
		0: c.getIncidentsByStatus(0, &up),
		1: c.getIncidentsByStatus(1, &up),
		2: c.getIncidentsByStatus(2, &up),
		3: c.getIncidentsByStatus(3, &up),
		4: c.getIncidentsByStatus(4, &up),
	}

	for _, group := range groups {
		c.createComponentsMetric(group, ch)
		c.createIncidentsTotalMetric(group, incidents, ch)
	}

	ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, float64(up))
	ch <- prometheus.MustNewConstMetric(c.scrapeDuration, prometheus.GaugeValue, time.Since(start).Seconds())
}

func (c *cachetCollector) createComponentsMetric(group cachet.ComponentGroup, ch chan<- prometheus.Metric) {
	componentsByStatus := map[int][]cachet.Component{
		0: make([]cachet.Component, 0),
		1: make([]cachet.Component, 0),
		2: make([]cachet.Component, 0),
		3: make([]cachet.Component, 0),
		4: make([]cachet.Component, 0),
	}
	for _, component := range group.EnabledComponents {
		components := append(componentsByStatus[component.Status], component)
		componentsByStatus[component.Status] = components
	}

	for status, components := range componentsByStatus {
		ch <- prometheus.MustNewConstMetric(c.components, prometheus.GaugeValue, float64(len(components)), componentStatus[status], group.Name)
	}
}

func (c *cachetCollector) getIncidentsByStatus(status int, up *int) []cachet.Incident {
	incidents, err := c.client.GetAllIncidentsByStatus(status)
	if err != nil {
		log.With("error", err).Error("failed to scrape Group Components")
		*up = 0
	}
	return incidents
}

func (c *cachetCollector) createIncidentsTotalMetric(group cachet.ComponentGroup, incidents map[int][]cachet.Incident, ch chan<- prometheus.Metric) {
	for _, component := range group.EnabledComponents {
		c.createIncidentsTotalMetricByComponent(group, component, incidents, ch)
	}
}

func (c *cachetCollector) createIncidentsTotalMetricByComponent(group cachet.ComponentGroup, component cachet.Component, incidents map[int][]cachet.Incident, ch chan<- prometheus.Metric) {
	for status, allIncidents := range incidents {
		componentIncidents := make([]cachet.Incident, 0)
		for _, incident := range allIncidents {
			if incident.ComponentID == component.ID && status == incident.Status {
				componentIncidents = append(componentIncidents, incident)
			}
		}
		ch <- prometheus.MustNewConstMetric(c.incidents, prometheus.GaugeValue, float64(len(componentIncidents)), incidentStatus[status], group.Name, component.Name)
	}

}
