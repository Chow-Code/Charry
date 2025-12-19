package consul

import (
	"github.com/charry/config"
	"github.com/charry/logger"
)

// GracefulShutdown 优雅关闭时注销服务
func (c *Client) GracefulShutdown(appConfig *config.AppConfig) {
	if err := c.DeregisterService(appConfig); err != nil {
		logger.Errorf("注销服务失败: %v", err)
	} else {
		logger.Infof("服务注销成功: %s-%s-%d",
			appConfig.Type, appConfig.Environment, appConfig.Id)
	}
}
