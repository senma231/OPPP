package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/senma231/p3/common/logger"
	"github.com/senma231/p3/server/api/middleware"
	"github.com/senma231/p3/server/app"
	"github.com/senma231/p3/server/auth"
	"github.com/senma231/p3/server/config"
	"github.com/senma231/p3/server/device"
	"github.com/senma231/p3/server/forward"
)

// Router API 路由
type Router struct {
	cfg            *config.Config
	authService    *auth.Service
	deviceService  *device.Service
	appService     *app.Service
	forwardService *forward.Service
}

// NewRouter 创建 API 路由
func NewRouter(cfg *config.Config, authService *auth.Service, deviceService *device.Service, appService *app.Service, forwardService *forward.Service) *Router {
	return &Router{
		cfg:            cfg,
		authService:    authService,
		deviceService:  deviceService,
		appService:     appService,
		forwardService: forwardService,
	}
}

// SetupRouter 设置路由
func SetupRouter(
	authService *auth.Service,
	deviceService *device.Service,
	appService *app.Service,
	forwardService *forward.Service,
) *gin.Engine {
	// 创建 Gin 引擎
	router := gin.New()

	// 使用中间件
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// API 版本
	v1 := router.Group("/api/v1")

	// 认证路由
	auth := v1.Group("/auth")
	{
		auth.POST("/register", Register)
		auth.POST("/login", Login)
		auth.POST("/refresh", RefreshToken)
		auth.POST("/logout", middleware.Auth(authService), Logout)
	}

	// 用户路由
	users := v1.Group("/users")
	users.Use(middleware.Auth(authService))
	{
		users.GET("/me", GetCurrentUser)
		users.PUT("/me", UpdateCurrentUser)
		users.PUT("/me/password", ChangePassword)
		users.POST("/me/2fa/enable", EnableTOTP)
		users.POST("/me/2fa/verify", VerifyTOTP)
		users.POST("/me/2fa/disable", DisableTOTP)
	}

	// 设备路由
	devices := v1.Group("/devices")
	devices.Use(middleware.Auth(authService))
	{
		devices.GET("", GetDevices)
		devices.POST("", CreateDevice)
		devices.GET("/:id", GetDevice)
		devices.PUT("/:id", UpdateDevice)
		devices.DELETE("/:id", DeleteDevice)
		devices.POST("/:id/token", RegenerateDeviceToken)
	}

	// 应用路由
	apps := v1.Group("/apps")
	apps.Use(middleware.Auth(authService))
	{
		apps.GET("", GetApps)
		apps.POST("", CreateApp)
		apps.GET("/:id", GetApp)
		apps.PUT("/:id", UpdateApp)
		apps.DELETE("/:id", DeleteApp)
		apps.POST("/:id/start", StartApp)
		apps.POST("/:id/stop", StopApp)
	}

	// 转发路由
	forwards := v1.Group("/forwards")
	forwards.Use(middleware.Auth(authService))
	{
		forwards.GET("", GetForwards)
		forwards.POST("", CreateForward)
		forwards.GET("/:id", GetForward)
		forwards.PUT("/:id", UpdateForward)
		forwards.DELETE("/:id", DeleteForward)
		forwards.POST("/:id/enable", EnableForward)
		forwards.POST("/:id/disable", DisableForward)
	}

	// 设备 API 路由
	deviceAPI := v1.Group("/device")
	deviceAPI.Use(middleware.DeviceAuth(deviceService))
	{
		deviceAPI.POST("/status", UpdateDeviceStatus)
		deviceAPI.GET("/apps", GetDeviceApps)
	}

	// 统计路由
	stats := v1.Group("/stats")
	stats.Use(middleware.Auth(authService))
	{
		stats.GET("/system", GetSystemStats)
		stats.GET("/user", GetUserStats)
	}

	logger.Info("API 路由设置完成")
	return router
}
