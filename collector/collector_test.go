package collector

import (
	"regexp"
	"testing"

	"github.com/andygrunwald/cachet"
	"github.com/prometheus/client_golang/prometheus"

	dto "github.com/prometheus/client_model/go"

	assert "github.com/stretchr/testify/require"
)

type dummyClient struct {
	IncidentsTotal int
}

func (d *dummyClient) GetAllComponentGroups() ([]cachet.ComponentGroup, error) {
	components := []cachet.Component{cachet.Component{
		ID:   1,
		Name: "Component",
	}}
	return []cachet.ComponentGroup{cachet.ComponentGroup{EnabledComponents: components}}, nil
}

func (d *dummyClient) GetAllIncidentsByStatus(status int) ([]cachet.Incident, error) {
	incidents := make([]cachet.Incident, 0)
	for i := 0; i < d.IncidentsTotal; i++ {
		incidents = append(incidents, cachet.Incident{Status: 1, ComponentID: 1})
	}
	return incidents, nil
}

func (d *dummyClient) Ping() (float64, error) {
	return 1, nil
}

func TestDescribe(t *testing.T) {
	client := &dummyClient{}
	collector := NewCachetCollector(client)

	ch := make(chan *prometheus.Desc, 3)
	collector.Describe(ch)

	up := <-ch
	scrapeDuration := <-ch
	incidentsTotal := <-ch
	assert.Contains(t, up.String(), "Cachet API is up and accepting requests")
	assert.Contains(t, scrapeDuration.String(), "Time Cachet scrape took in seconds")
	assert.Contains(t, incidentsTotal.String(), "Total of incidents by status")
}

func TestCollectCachetUp(t *testing.T) {
	client := &dummyClient{}
	collector := NewCachetCollector(client)
	ch := make(chan prometheus.Metric)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()
	metric := getMetrics("cachet_up", ch)[0]

	assert.NotNil(t, metric)
	assert.Equal(t, float64(1), *metric.GetGauge().Value)
}

func TestCollectCachetInsidentsByStatus(t *testing.T) {
	client := &dummyClient{
		IncidentsTotal: 10,
	}
	collector := NewCachetCollector(client)
	ch := make(chan prometheus.Metric)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()
	metric := getMetrics("cachet_incidents_total", ch)[0]

	assert.NotNil(t, metric)
	assert.Equal(t, float64(10), *metric.GetGauge().Value)
}

func getMetrics(key string, ch <-chan prometheus.Metric) []*dto.Metric {
	result := make([]*dto.Metric, 0)
	reg := regexp.MustCompile("fqName:\\s\"(.+?)\",")
	for m := range ch {
		metric := &dto.Metric{}
		m.Write(metric)
		matches := reg.FindStringSubmatch(m.Desc().String())

		if len(matches) > 0 && matches[1] == key {
			result = append(result, metric)
		}
	}
	return result
}
