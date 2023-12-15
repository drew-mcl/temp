package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"go.etcd.io/etcd/clientv3"
)

type Client struct {
	endpoint string
	client   *clientv3.Client
}

type Suite struct {
	Apps []string `json:"apps"`
}

type HealthCheck struct {
	Type string `json:"type"`
}

func NewClient(endpoint string) (*Client, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{endpoint},
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		endpoint: endpoint,
		client:   cli,
	}, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

func (c *Client) GetSuite(key string) (*Suite, error) {
	resp, err := c.client.Get(context.Background(), key)
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	var suite Suite
	err = json.Unmarshal(resp.Kvs[0].Value, &suite)
	if err != nil {
		return nil, err
	}

	return &suite, nil
}

func (c *Client) GetHealthCheck(key string) (*HealthCheck, error) {
	resp, err := c.client.Get(context.Background(), key)
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	var healthCheck HealthCheck
	err = json.Unmarshal(resp.Kvs[0].Value, &healthCheck)
	if err != nil {
		return nil, err
	}

	return &healthCheck, nil
}
