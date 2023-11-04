package discover

import (
	"fmt"
	"github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"log"
	"sync"
)

type KitDiscoverClient struct {
	Host   string
	Port   int
	client consul.Client
	//连接consul的配置
	config *api.Config
	mutex  sync.Mutex
	//服务实例缓存字段
	instanceMap sync.Map
}

func (this *KitDiscoverClient) Register(serviceName, instanceID, healthCheckUrl, instanceHost string, instancePort int,
	meta map[string]string, logger *log.Logger) bool {
	serviceInfo := &api.AgentServiceRegistration{
		Name:    serviceName,
		ID:      instanceID,
		Meta:    meta,
		Address: instanceHost,
		Port:    instancePort,
		Check: &api.AgentServiceCheck{
			DeregisterCriticalServiceAfter: "30s",
			HTTP:                           fmt.Sprintf("http://%s:%d/%s", instanceHost, instancePort, healthCheckUrl),
			Interval:                       "15s",
		},
	}
	err := this.client.Register(serviceInfo)

	if err != nil {
		log.Println("KitRegister service fail", err)
		return false
	}
	log.Println("KitRegister service success")
	return true
}

func (this *KitDiscoverClient) DeRegister(instanceID string, logger *log.Logger) bool {
	serviceInfo := &api.AgentServiceRegistration{ID: instanceID}
	err := this.client.Deregister(serviceInfo)

	if err != nil {
		log.Println("KitDeregister service fail", err)
		return false
	}
	log.Println("KitDeregister service success")
	return true
}

func (this *KitDiscoverClient) DiscoverServices(serviceName string, logger *log.Logger) []interface{} {

	//1.该服务已被监控缓存
	instancesList, ok := this.instanceMap.Load(serviceName)
	if ok {
		return instancesList.([]interface{})
	}

	//2.申请锁，保护并发,类似于singlefight
	this.mutex.Lock()
	//再次检查是否监控(第一个会false，其余会true)
	instancesList, ok = this.instanceMap.Load(serviceName)
	if ok {
		return instancesList.([]interface{})
	} else {
		//注册监控
		go func() {
			params := make(map[string]interface{})
			params["type"] = "service"
			params["service"] = serviceName
			plan, _ := watch.Parse(params)
			plan.Handler = func(u uint64, i interface{}) {
				if i == nil {
					return
				}
				v, ok := i.([]*api.ServiceEntry)
				if !ok {
					return
				}
				if len(v) == 0 {
					this.instanceMap.Store(serviceName, []interface{}{})
				}
				var healthServices []interface{}
				for _, v := range v {
					if v.Checks.AggregatedStatus() == api.HealthPassing {
						healthServices = append(healthServices, v.Service)
					}
				}
				this.instanceMap.Store(serviceName, healthServices)
			}
			defer plan.Stop()
			plan.Run(this.config.Address)
		}()
	}
	defer this.mutex.Unlock()

	entries, _, err := this.client.Service(serviceName, "", false, nil)
	if err != nil {
		this.instanceMap.Store(serviceName, []interface{}{})
		log.Println("KitDiscover service fail", err)
		return nil
	}
	log.Println("KitDiscover service success")
	instances := make([]interface{}, len(entries))
	for i, v := range entries {
		instances[i] = v.Service
	}
	this.instanceMap.Store(serviceName, instances)
	return instances
}

func NewKitDiscoverClient(host string, port int) (DiscoveryClient, error) {
	consulConfig := api.DefaultConfig()
	consulConfig.Address = fmt.Sprintf("%s:%d", host, port)
	apiClient, err := api.NewClient(consulConfig)

	if err != nil {
		return nil, err
	}

	client := consul.NewClient(apiClient)
	return &KitDiscoverClient{
		Host:   host,
		Port:   port,
		client: client,
		config: consulConfig,
	}, nil

}
