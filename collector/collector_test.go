package collector

import (
	"testing"

	"github.com/andygrunwald/cachet"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/stretchr/testify/assert"
)

type dummyClient struct{}

func (d *dummyClient) GetAllComponentsGroups() ([]cachet.ComponentGroup, error) {
	return make([]cachet.ComponentGroup, 0), nil
}

func (d *dummyClient) GetAllIncidentsByStatus(status int) ([]cachet.Incident, error) {
	return make([]cachet.Incident, 0), nil
}

func (d *dummyClient) Ping() (string, error) {
	return "up", nil
}

func TestDescribe(t *testing.T) {
	client := &dummyClient{}
	config := Config{}
	collector := NewCachetCollector(client, config)

	ch := make(chan *prometheus.Desc, 3)
	collector.Describe(ch)

	up := <-ch
	scrapeDuration := <-ch
	incidentsTotal := <-ch
	assert.Contains(t, up.String(), "Cachet API is up and accepting requests")
	assert.Contains(t, scrapeDuration.String(), "Time Cachet scrape took in seconds")
	assert.Contains(t, incidentsTotal.String(), "Total of incidents by status")
}

func TestCollect(t *testing.T) {
	// client := &cachet.Client{}
	// config := Config{}
	// collector := NewCachetCollector(client, config)

	// ch := make(chan prometheus.Metric, 3)
	// collector.Collect(ch)

	// up := <-ch
	// scrapeDuration := <-ch
	// incidentsTotal := <-ch
}
