package client

import (
	"github.com/andygrunwald/cachet"
)

type Client interface {
	GetAllComponentGroups() ([]cachet.ComponentGroup, error)
	GetAllIncidentsByStatus(status int) ([]cachet.Incident, error)
	Ping() (float64, error)
}

type cachetClient struct {
	client *cachet.Client
}

func (c *cachetClient) GetAllComponentGroups() ([]cachet.ComponentGroup, error) {
	groups, _, err := c.client.ComponentGroups.GetAll(&cachet.ComponentGroupsQueryParams{})

	if err != nil {
		return nil, err
	}

	return groups.ComponentGroups, nil
}

func (c *cachetClient) Ping() (float64, error) {
	_, _, err := c.client.General.Ping()
	return 1, err
}

func NewCachetClient(apiURL string) (Client, error) {
	client, err := cachet.NewClient(apiURL, nil)
	if err != nil {
		return nil, err
	}

	return &cachetClient{
		client: client,
	}, nil
}

func (c *cachetClient) GetAllIncidentsByStatus(status int) ([]cachet.Incident, error) {
	incidents, _, err := c.client.Incidents.GetAll(&cachet.IncidentsQueryParams{Status: status})

	if err != nil {
		return nil, err
	}

	return incidents.Incidents, nil
}
