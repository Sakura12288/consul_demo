package discover

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

//服务实例结构体

type InstanceInfo struct {
	ID                string            `json:"ID"`
	Name              string            `json:"Name"`
	Service           string            `json:"Service,omitempty"`
	Tags              []string          `json:"Tags,omitempty"`
	Address           string            `json:"Address"`
	Port              int               `json:"Port"`
	Meta              map[string]string `json:"Meta,omitempty"`
	EnableTagOverride bool              `json:"EnableTagOverride"`
	Check             `json:"Check"`
	Weights           `json:"Weights"`
}

type Check struct {
	DeregisterCriticalServiceAfter string   `json:"DeregisterCriticalServiceAfter"`
	Args                           []string `json:"Args,omitempty"`
	HTTP                           string   `json:"HTTP"`
	Interval                       string   `json:"Interval,omitempty"`
	TTL                            string   `json:"TTL,omitempty"`
}
type Weights struct {
	Passing int `json:"Passing"`
	Warning int `json:"Warning"`
}

// consul 配置
type MyDiscoverClient struct {
	Host string
	Port int
}

// 满足DiscoveryClient接口
func (this *MyDiscoverClient) Register(serviceName, instanceID, healthCheckUrl, instanceHost string, instancePort int,
	meta map[string]string, logger *log.Logger) bool {
	// 1. 封装服务实例的元数据
	Info := &InstanceInfo{
		ID:                instanceID,
		Name:              serviceName,
		Address:           instanceHost,
		Port:              instancePort,
		Meta:              meta,
		EnableTagOverride: false,
		Check: Check{
			DeregisterCriticalServiceAfter: "30s",
			HTTP:                           "http://" + instanceHost + ":" + strconv.Itoa(instancePort) + healthCheckUrl,
			Interval:                       "15s",
		},
		Weights: Weights{
			Passing: 10,
			Warning: 1,
		},
	}

	//2. 向Consul发送服务注册的请求
	byteData, _ := json.Marshal(Info)
	req, err := http.NewRequest("PUT",
		"http://"+this.Host+":"+strconv.Itoa(this.Port)+
			"/v1/agent/service/register", bytes.NewReader(byteData))

	if err == nil {
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		client := http.Client{}
		response, err := client.Do(req)
		if err != nil {
			log.Printf("Register Service Error : %s", err.Error())
		} else {
			defer response.Body.Close()
			if response.StatusCode == 200 {
				log.Println("Register service success!")
			} else {
				log.Println("Register service error")
			}
		}
	}
	return true
}

func (this *MyDiscoverClient) DeRegister(instanceID string, logger *log.Logger) bool {
	req, err := http.NewRequest("PUT", fmt.Sprintf("http://%s:%d/v1/agent/service/deregister/%s",
		this.Host, this.Port, instanceID), nil)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Deregister service failed :", err)
	} else {
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			log.Println("Deregister service success ")
		} else {
			log.Println("Deregister service failed")
		}
	}
	return false
}

func (this *MyDiscoverClient) DiscoverServices(serviceName string, logger *log.Logger) []interface{} {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:%d/v1/health/service/%s",
		this.Host, this.Port, serviceName), nil)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Discover service failed :", err)
	} else if resp.StatusCode == 200 {
		defer resp.Body.Close()
		var serviceList []struct {
			Service InstanceInfo `json:"Service"`
		}
		err = json.NewDecoder(resp.Body).Decode(&serviceList)
		resp.Body.Close()
		if err == nil {
			instances := make([]interface{}, len(serviceList))
			for i, v := range serviceList {
				instances[i] = v.Service
			}
			return instances
		}
	}
	return nil
}

func NewMyDiscoverClient(consulHost string, consulPort int) (DiscoveryClient, error) {
	return &MyDiscoverClient{
		Host: consulHost,
		Port: consulPort,
	}, nil
}
