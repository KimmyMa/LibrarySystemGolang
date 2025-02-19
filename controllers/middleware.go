package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

// AuthMiddleware 检查用户是否已登录，并验证角色
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查登录状态
		loginStatus, err := c.Cookie("loginStatus")
		if err != nil || loginStatus != "true" {
			c.Redirect(http.StatusFound, "/login")
			c.Abort() // 停止后续处理
			return
		}

		// 获取用户角色
		isAdmin, _ := c.Cookie("isAdmin")
		readerID, _ := c.Cookie("readercard")

		// 获取请求路径
		requestPath := c.Request.URL.Path

		// 检查路径是否符合角色权限
		if isAdmin == "true" {
			// 管理员只能访问以 "/admin" 开头的路径
			if !strings.HasPrefix(requestPath, "/admin") {
				c.Redirect(http.StatusFound, "/admin")
				c.Abort()
				return
			}
		} else if readerID != "" {
			// 读者只能访问以 "/reader" 开头的路径
			if !strings.HasPrefix(requestPath, "/reader") {
				c.Redirect(http.StatusFound, "/reader")
				c.Abort()
				return
			}
		} else {
			// 未识别的角色，重定向到登录页面
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// 继续处理请求
		c.Next()
	}
}
