package client

import (
	"github.com/andygrunwald/cachet"
)

// Client manages communication with the Cachet API.
type Client interface {
	GetAllComponentGroups() ([]cachet.ComponentGroup, error)
	GetAllIncidentsByStatus(status int) ([]cachet.Incident, error)
}

type cachetClient struct {
	client *cachet.Client
}

// NewCachetClient returns an initialized Cachet API client
func NewCachetClient(apiURL string) (Client, error) {
	client, err := cachet.NewClient(apiURL, nil)
	if err != nil {
		return nil, err
	}

	return &cachetClient{
		client: client,
	}, nil
}

func (c *cachetClient) GetAllComponentGroups() ([]cachet.ComponentGroup, error) {
	groups, _, err := c.client.ComponentGroups.GetAll(&cachet.ComponentGroupsQueryParams{})
	if err != nil {
		return nil, err
	}

	return groups.ComponentGroups, nil
}

func (c *cachetClient) GetAllIncidentsByStatus(status int) ([]cachet.Incident, error) {
	incidents, _, err := c.client.Incidents.GetAll(&cachet.IncidentsQueryParams{Status: status})
	if err != nil {
		return nil, err
	}

	return incidents.Incidents, nil
}
