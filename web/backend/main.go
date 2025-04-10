package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// 解析命令行参数
	port := flag.Int("port", 8080, "API 服务端口")
	mode := flag.String("mode", "debug", "运行模式 (debug, release)")
	flag.Parse()

	// 设置 Gin 模式
	if *mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建 Gin 引擎
	r := gin.Default()

	// 跨域中间件
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 注册路由
	setupRoutes(r)

	// 启动服务
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("API 服务启动在 %s\n", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("启动 API 服务失败: %v", err)
	}
}

func setupRoutes(r *gin.Engine) {
	// API 版本
	v1 := r.Group("/api/v1")

	// 用户认证
	auth := v1.Group("/auth")
	{
		auth.POST("/login", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "登录 API 尚未实现",
			})
		})
		auth.POST("/register", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "注册 API 尚未实现",
			})
		})
		auth.POST("/logout", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "登出 API 尚未实现",
			})
		})
	}

	// 设备管理
	devices := v1.Group("/devices")
	{
		devices.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "获取设备列表 API 尚未实现",
			})
		})
		devices.GET("/:id", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "获取设备详情 API 尚未实现",
				"id":      c.Param("id"),
			})
		})
		devices.POST("/", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "创建设备 API 尚未实现",
			})
		})
		devices.PUT("/:id", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "更新设备 API 尚未实现",
				"id":      c.Param("id"),
			})
		})
		devices.DELETE("/:id", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "删除设备 API 尚未实现",
				"id":      c.Param("id"),
			})
		})
	}

	// 应用管理
	apps := v1.Group("/apps")
	{
		apps.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "获取应用列表 API 尚未实现",
			})
		})
		apps.GET("/:id", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "获取应用详情 API 尚未实现",
				"id":      c.Param("id"),
			})
		})
		apps.POST("/", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "创建应用 API 尚未实现",
			})
		})
		apps.PUT("/:id", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "更新应用 API 尚未实现",
				"id":      c.Param("id"),
			})
		})
		apps.DELETE("/:id", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "删除应用 API 尚未实现",
				"id":      c.Param("id"),
			})
		})
	}

	// 端口转发管理
	forwards := v1.Group("/forwards")
	{
		forwards.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "获取转发规则列表 API 尚未实现",
			})
		})
		forwards.GET("/:id", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "获取转发规则详情 API 尚未实现",
				"id":      c.Param("id"),
			})
		})
		forwards.POST("/", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "创建转发规则 API 尚未实现",
			})
		})
		forwards.PUT("/:id", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "更新转发规则 API 尚未实现",
				"id":      c.Param("id"),
			})
		})
		forwards.DELETE("/:id", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "删除转发规则 API 尚未实现",
				"id":      c.Param("id"),
			})
		})
	}

	// 系统状态
	status := v1.Group("/status")
	{
		status.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "获取系统状态 API 尚未实现",
			})
		})
	}
}
