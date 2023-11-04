package discover

import "log"

// consul客户端
type DiscoveryClient interface {
	/**
	* 服务注册接口
	* @param serviceName 服务名
	* @param instanceID 服务实例ID
	* @param instancePort 服务实例端口
	* @param healthCheckUrl 健康检查地址
	* @param instanceHost 服务实例地址
	* @param meta 服务实例元数据
	 */
	Register(serviceName, instanceID, healthCheckUrl, instanceHost string, instancePort int,
		meta map[string]string, logger *log.Logger) bool

	/**
	* 服务注销端口
	* @param instanceID 服务实例ID
	 */
	DeRegister(instanceID string, logger *log.Logger) bool

	/**
	* 服务发现接口
	* @param serviceName 服务名
	 */
	DiscoverServices(serviceName string, logger *log.Logger) []interface{}
}
