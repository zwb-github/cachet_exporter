package client

import (
	"github.com/andygrunwald/cachet"
)

type Client interface {
	GetAllComponentsGroups() ([]cachet.ComponentGroup, error)
	GetAllIncidentsByStatus(status int) ([]cachet.Incident, error)
	Ping() (string, error)
}

type cachetClient struct {
	client *cachet.Client
}

func (c *cachetClient) GetAllComponentsGroups() ([]cachet.ComponentGroup, error) {
	groups, _, err := c.client.ComponentGroups.GetAll(&cachet.ComponentGroupsQueryParams{})

	if err != nil {
		return nil, err
	}

	return groups.ComponentGroups, nil
}

func (c *cachetClient) Ping() (string, error) {
	response, _, err := c.client.General.Ping()
	return response, err
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
