package api

import (
	"github.com/gin-gonic/gin"
	"github.com/senma231/p3/server/app"
	"github.com/senma231/p3/server/auth"
	"github.com/senma231/p3/server/device"
	"github.com/senma231/p3/server/forward"
	"github.com/senma231/p3/server/monitor"
)

// SetupRouter 设置路由
func SetupRouter(
	authService *auth.Service,
	deviceService *device.Service,
	appService *app.Service,
	forwardService *forward.Service,
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
	forwardController := NewForwardController(forwardService)
	
	// 创建监控器
	monitorInstance := monitor.NewMonitor()
	monitorInstance.Start()
	wsHandler := NewWSHandler(monitorInstance)
	
	// 创建权限管理器
	permMgr := auth.NewPermissionManager()
	permController := NewPermissionController(permMgr)
	
	// 创建分组控制器
	groupController := NewGroupController(deviceService)
	
	// 创建批量操作控制器
	batchController := NewBatchController(deviceService, appService, forwardService)

	// API 版本
	v1 := r.Group("/api/v1")

	// 认证路由
	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/register", authController.Register)
		authGroup.POST("/login", authController.Login)
		authGroup.POST("/logout", authController.Logout)
	}

	// 需要认证的路由
	authorized := v1.Group("/")
	authorized.Use(AuthMiddleware(authService))
	{
		// 当前用户
		authorized.GET("/user", authController.GetCurrentUser)
		authorized.PUT("/user", authController.UpdateUser)

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
		
		// 转发规则管理
		forwards := authorized.Group("/forwards")
		{
			forwards.GET("/", forwardController.GetForwards)
			forwards.GET("/:id", forwardController.GetForward)
			forwards.POST("/", forwardController.CreateForward)
			forwards.PUT("/:id", forwardController.UpdateForward)
			forwards.DELETE("/:id", forwardController.DeleteForward)
			forwards.POST("/:id/enable", forwardController.EnableForward)
			forwards.POST("/:id/disable", forwardController.DisableForward)
			forwards.GET("/:id/stats", forwardController.GetForwardStats)
		}
		
		// 系统状态
		authorized.GET("/status", deviceController.GetSystemStatus)
		
		// WebSocket
		authorized.GET("/ws", wsHandler.HandleWS)
		
		// 权限管理
		permissions := authorized.Group("/permissions")
		{
			permissions.GET("/users/:id/role", permController.GetUserRole)
			permissions.PUT("/users/:id/role", permController.SetUserRole)
			permissions.GET("/users/:id/permissions", permController.GetUserPermissions)
			permissions.POST("/check", permController.CheckPermission)
		}
		
		// 分组管理
		groups := authorized.Group("/groups")
		{
			groups.GET("/", groupController.GetGroups)
			groups.POST("/", groupController.CreateGroup)
			groups.GET("/:id", groupController.GetGroup)
			groups.PUT("/:id", groupController.UpdateGroup)
			groups.DELETE("/:id", groupController.DeleteGroup)
			groups.POST("/:id/devices/:deviceId", groupController.AddDeviceToGroup)
			groups.DELETE("/:id/devices/:deviceId", groupController.RemoveDeviceFromGroup)
			groups.GET("/:id/devices", groupController.GetDevicesInGroup)
		}
		
		// 批量操作
		batch := authorized.Group("/batch")
		{
			batch.POST("/devices", batchController.BatchDeviceOperation)
			batch.POST("/apps", batchController.BatchAppOperation)
			batch.POST("/forwards", batchController.BatchForwardOperation)
		}
	}

	return r
}
