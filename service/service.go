package service

import (
	"context"
	"errors"
	"service/config"
	"service/discover"
)

var (
	ErrServiceNotExist = errors.New("服务不存在")
)

type Service interface {
	//健康检查接口
	HealthCheck() bool
	//打招呼接口
	SayHello() string
	//从consul通过服务名发现接口
	DiscoverServices(ctx context.Context, serviceName string) ([]interface{}, error)
}

type DiscoveryServiceImpl struct {
	discoveryClient discover.DiscoveryClient
}

//创建

func NewDiscoveryServiceImpl(discoveryClient discover.DiscoveryClient) *DiscoveryServiceImpl {
	return &DiscoveryServiceImpl{
		discoveryClient: discoveryClient,
	}
}

//简单实现service接口

func (service *DiscoveryServiceImpl) HealthCheck() bool {
	return true
}

func (service *DiscoveryServiceImpl) SayHello() string {
	return "hello"
}

func (service *DiscoveryServiceImpl) DiscoverServices(ctx context.Context, serviceName string) ([]interface{}, error) {
	//从consul根据服务名获取服务实例列表
	instances := service.discoveryClient.DiscoverServices(serviceName, config.Logger)
	if instances == nil || len(instances) == 0 {
		return nil, ErrServiceNotExist
	}
	return instances, nil
}
