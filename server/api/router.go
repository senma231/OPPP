package api

import (
	"github.com/gin-gonic/gin"
	"github.com/senma231/p3/server/app"
	"github.com/senma231/p3/server/auth"
	"github.com/senma231/p3/server/device"
)

// SetupRouter 设置路由
func SetupRouter(
	authService *auth.Service,
	deviceService *device.Service,
	appService *app.Service,
) *gin.Engine {
	// 创建 Gin 引擎
	r := gin.Default()

	// 使用中间件
	r.Use(CORSMiddleware())
	r.Use(LoggerMiddleware())
	r.Use(RecoveryMiddleware())

	// 创建控制器
	authController := NewAuthController(authService)
	deviceController := NewDeviceController(deviceService)
	appController := NewAppController(appService)

	// API 版本
	v1 := r.Group("/api/v1")

	// 认证路由
	auth := v1.Group("/auth")
	{
		auth.POST("/register", authController.Register)
		auth.POST("/login", authController.Login)
		auth.POST("/logout", authController.Logout)
	}

	// 需要认证的路由
	authorized := v1.Group("/")
	authorized.Use(AuthMiddleware(authService))
	{
		// 当前用户
		authorized.GET("/user", authController.GetCurrentUser)

		// 设备管理
		devices := authorized.Group("/devices")
		{
			devices.GET("/", deviceController.GetDevices)
			devices.GET("/:id", deviceController.GetDevice)
			devices.POST("/", deviceController.CreateDevice)
			devices.PUT("/:id", deviceController.UpdateDevice)
			devices.DELETE("/:id", deviceController.DeleteDevice)
			devices.GET("/:id/stats", deviceController.GetDeviceStats)
		}

		// 应用管理
		apps := authorized.Group("/apps")
		{
			apps.GET("/", appController.GetApps)
			apps.GET("/:id", appController.GetApp)
			apps.POST("/", appController.CreateApp)
			apps.PUT("/:id", appController.UpdateApp)
			apps.DELETE("/:id", appController.DeleteApp)
			apps.POST("/:id/start", appController.StartApp)
			apps.POST("/:id/stop", appController.StopApp)
			apps.GET("/:id/stats", appController.GetAppStats)
		}
	}

	return r
}
