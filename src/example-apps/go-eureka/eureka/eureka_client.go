package eureka

import "net"

type EurekaClient struct {
	UAAClient     UAAClient
	EurekaBaseURL string
}

func (c *EurekaClient) Register(myAppName string, myAppID string) error {}

func (c *EurekaClient) List(appName string) ([]net.IP, error) {}
