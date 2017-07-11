package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"code.cloudfoundry.org/localip"

	"github.com/ryanmoran/viron"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/ifrit/sigmon"
)

type Environment struct {
	VCAPServices struct {
		ServiceRegistry []struct {
			Credentials struct {
				RegistryURI    string `json:"uri"`
				ClientSecret   string `json:"client_secret"`
				ClientID       string `json:"client_id"`
				AccessTokenURI string `json:"access_token_uri"`
			} `json:"credentials"`
		} `json:"p-service-registry"`
	} `env:"VCAP_SERVICES" env-required:"true"`

	VCAPApplication struct {
		ApplicationName string `json:"application_name"`
		InstanceIndex   int    `json:"instance_index"`
	} `env:"VCAP_APPLICATION" env-required:"true"`

	Port               string `env:"PORT"               env-required:"true"`
	StartPort          string `env:"START_PORT"         env-required:"false"`
	ListenPorts        string `env:"LISTEN_PORTS"       env-required:"false"`
	RegistryTTLSeconds string `env:"REGISTRY_TTL_SECONDS"       env-required:"false"`
}

func main() {
	if err := mainWithError(); err != nil {
		log.Printf("%s", err)
		os.Exit(1)
	}
}

func mainWithError() error {
	var env Environment
	err := viron.Parse(&env)
	if err != nil {
		return fmt.Errorf("unable to parse environment: %s", err)
	}

	var startPort int
	if env.StartPort != "" {
		startPort, err = strconv.Atoi(env.StartPort)
		if err != nil {
			return fmt.Errorf("invalid env var START_PORT: %s", err)
		}
	}

	var listenPorts int
	if env.ListenPorts != "" {
		listenPorts, err = strconv.Atoi(env.ListenPorts)
		if err != nil {
			return fmt.Errorf("invalid env var LISTEN_PORTS: %s", err)
		}
	}

	var ttlSeconds int
	if env.RegistryTTLSeconds != "" {
		ttlSeconds, err = strconv.Atoi(env.RegistryTTLSeconds)
		if err != nil {
			return fmt.Errorf("invalid env var REGISTRY_TTL_SECONDS: %s", err)
		}
	}
	if ttlSeconds < 10 {
		fmt.Printf("Setting TTL to 10s because min TTL of registry is 10 seconds")
		ttlSeconds = 10
	}

	localIP, err := localip.LocalIP()
	if err != nil {
		return fmt.Errorf("unable to discover local ip: %s", err)
	}

	var serviceInstances []ServiceInstance
	for i := 0; i < listenPorts; i++ {
		serviceInstances = append(serviceInstances, ServiceInstance{
			Name:     env.VCAPApplication.ApplicationName,
			Instance: env.VCAPApplication.InstanceIndex,
			IP:       localIP,
			Port:     startPort + i,
		})
	}

	serviceCredentials := env.VCAPServices.ServiceRegistry[0].Credentials

	uaaClient := &UAAClient{
		BaseURL: serviceCredentials.AccessTokenURI,
		Name:    serviceCredentials.ClientID,
		Secret:  serviceCredentials.ClientSecret,
	}

	eurekaClient := &EurekaClient{
		BaseURL:          serviceCredentials.RegistryURI,
		HttpClient:       http.DefaultClient,
		UAAClient:        uaaClient,
		ServiceInstances: serviceInstances,
	}

	pollInterval := time.Duration(ttlSeconds*1000*1/4) * time.Millisecond // we can fail twice and not lose presence in the registry
	fmt.Printf("ttl is %d seconds, polling interval is %v\n", ttlSeconds, pollInterval)
	poller := &Poller{
		PollInterval: pollInterval,
		Action:       eurekaClient.RegisterAll,
	}

	infoHandler := &InfoHandler{
		InfoData: env.VCAPApplication,
	}

	servers := []ifrit.Runner{http_server.New(fmt.Sprintf("0.0.0.0:%s", env.Port), infoHandler)}
	for i := 0; i < listenPorts; i++ {
		servers = append(servers, http_server.New(fmt.Sprintf("0.0.0.0:%d", startPort+i), infoHandler))
	}

	members := []grouper.Member{}
	for i, server := range servers {
		members = append(members, grouper.Member{fmt.Sprintf("http_server_%d", i), server})
	}

	// poller goes at the end, so that registration happens after all servers start
	members = append(members, grouper.Member{"registration_poller", poller})

	monitor := ifrit.Invoke(sigmon.New(grouper.NewOrdered(os.Interrupt, members)))

	err = <-monitor.Wait()
	if err != nil {
		return fmt.Errorf("ifrit monitor: %s", err)
	}

	return nil
}

type InfoHandler struct {
	InfoData interface{}
}

func (h *InfoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(h.InfoData)
}

type EurekaClient struct {
	BaseURL          string
	HttpClient       *http.Client
	UAAClient        *UAAClient
	ServiceInstances []ServiceInstance
}

func (e *EurekaClient) RegisterAll() error {
	for _, s := range e.ServiceInstances {
		err := e.Register(s)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *EurekaClient) Register(serviceInstance ServiceInstance) error {
	token, err := e.UAAClient.GetToken()
	if err != nil {
		return err
	}

	postBody := map[string]interface{}{
		"instance": map[string]interface{}{
			"hostName": fmt.Sprintf("%s-%d-%d", serviceInstance.Name, serviceInstance.Instance, serviceInstance.Port),
			"app":      serviceInstance.Name,
			"ipAddr":   serviceInstance.IP,
			"status":   "UP",
			"port": map[string]interface{}{
				"$":        fmt.Sprintf("%d", serviceInstance.Port),
				"@enabled": "true",
			},
			"dataCenterInfo": map[string]interface{}{
				"@class": "com.netflix.appinfo.InstanceInfo$DefaultDataCenterInfo",
				"name":   "MyOwn",
			},
		},
	}
	reqBytes, err := json.Marshal(postBody)
	if err != nil {
		return err
	}

	url, err := e.createURL(fmt.Sprintf("/eureka/apps/%s", serviceInstance.Name))
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("bearer %s", token))

	resp, err := e.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected response code: %d: %s", resp.StatusCode, respBytes)
	}

	return nil
}

func (e *EurekaClient) createURL(route string) (string, error) {
	u, err := url.Parse(e.BaseURL)
	if err != nil {
		return "", fmt.Errorf("unable to parse base url: %s", err)
	}
	u.Path = path.Join(u.Path, route)
	return u.String(), nil
}

type ServiceInstance struct {
	Name     string
	Instance int
	IP       string
	Port     int
}

type UAAClient struct {
	BaseURL string
	Name    string
	Secret  string
}

func (c *UAAClient) GetToken() (string, error) {
	bodyString := "grant_type=client_credentials"
	request, err := http.NewRequest("POST", c.BaseURL, strings.NewReader(bodyString))
	request.SetBasicAuth(c.Name, c.Secret)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	type getTokenResponse struct {
		AccessToken string `json:"access_token"`
	}
	response := &getTokenResponse{}
	err = c.makeRequest(request, response)
	if err != nil {
		return "", err
	}
	return response.AccessToken, nil
}

func (c *UAAClient) makeRequest(request *http.Request, response interface{}) error {
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("http client: %s", err)
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %s", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("bad uaa response, code %d, msg %s", resp.StatusCode, string(respBytes))
	}

	err = json.Unmarshal(respBytes, &response)
	if err != nil {
		return fmt.Errorf("unmarshal json: %s", err)
	}
	return nil
}

type Poller struct {
	PollInterval time.Duration
	Action       func() error
}

func (m *Poller) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	err := m.Action()
	if err != nil {
		return err
	}

	close(ready)

	for {
		select {
		case <-signals:
			return nil
		case <-time.After(m.PollInterval):
			err = m.Action()
			if err != nil {
				log.Printf("%s", err)
				continue
			}
		}
	}
}
