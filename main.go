package main

import (
	"LibrarySystemGolang/controllers"
	"fmt"
	"github.com/gin-gonic/gin"
	"html/template"
)

func safeJS(data interface{}) template.JS {
	return template.JS(fmt.Sprintf("%v", data))
}
func main() {
	r := gin.Default()
	// 注册自定义函数
	r.SetFuncMap(template.FuncMap{
		"safeJS": safeJS,
	})
	r.Use(controllers.ErrorMiddleware())
	r.LoadHTMLGlob("views/*")
	r.Static("/views", "./views")
	r.Static("/static", "./static")
	// 登录和登出
	r.GET("/", controllers.ToLogin)
	r.GET("/login", controllers.ToLogin)
	r.POST("/login", controllers.LoginCheck)
	r.GET("/logout", controllers.Logout)

	// 需要登录才能访问的路由
	authorized := r.Group("", controllers.AuthMiddleware())
	{
		// 管理员相关路由
		authorized.GET("/admin", controllers.AdminMain)
		authorized.GET("/admin/current_user", controllers.GetCurrentUserInfo)
		authorized.GET("/admin/repasswd", controllers.AdminShowRePassWrodPage)
		authorized.PUT("/admin/repasswd", controllers.AdminRePassWrod)

		// 管理员图书相关路由
		authorized.GET("/admin/book", controllers.AdminShowBookPage)
		authorized.GET("/admin/book/:id", controllers.AdminBookDetail)
		authorized.GET("/admin/book/hot", controllers.BookHot)
		authorized.GET("/admin/book/update/:id", controllers.AdminShowBookUpdatePage)
		authorized.PUT("/admin/book/:id", controllers.AdminBookUpdate)
		authorized.DELETE("/admin/book/:id", controllers.AdminBookDelete)
		authorized.GET("/admin/book/add", controllers.AdminShowBookAddPage)
		authorized.POST("/admin/book", controllers.AdminBookCreate)
		authorized.GET("/admin/book/import", controllers.AdminShowBookImportPage)
		authorized.POST("/admin/book/import", controllers.AdminBookImport)
		authorized.GET("/admin/book/export", controllers.AdminBookExport)

		// 管理员借阅相关路由
		authorized.GET("/admin/reserve", controllers.AdminReserveList)
		authorized.PUT("/admin/reserve/:id", controllers.AdminReserveAccept)
		authorized.DELETE("/admin/reserve/:id", controllers.AdminReserveDelete)
		authorized.GET("/admin/lend", controllers.AdminLendList)
		authorized.DELETE("/admin/lend/:id", controllers.AdminLendDelete)

		// 管理员读者相关路由
		authorized.GET("/admin/reader", controllers.AdminReader)
		authorized.POST("/admin/reader", controllers.AdminReaderCreate)
		authorized.DELETE("/admin/reader/:id", controllers.AdminReaderDelete)
		authorized.GET("/admin/reader/:id", controllers.AdminReaderQuery)
		authorized.PUT("/admin/reader/:id", controllers.AdminReaderUpdate)
		authorized.GET("/admin/reader/add", controllers.AdminShowReaderAddPage)

		// 读者相关路由
		authorized.GET("/reader", controllers.ReaderMain)
		authorized.GET("/reader/current_user", controllers.GetCurrentUserInfo)
		authorized.GET("/reader/info", controllers.ReaderInfo)
		authorized.GET("/reader/info/update/:id", controllers.ReaderShowInfoUpdatePage)
		authorized.PUT("/reader/info/:id", controllers.ReaderInfoUpdate)
		authorized.GET("/reader/repasswd", controllers.ReaderShowRePassWordPage)
		authorized.PUT("/reader/repasswd", controllers.ReaderRePassWord)

		// 读者图书相关路由
		authorized.GET("/reader/book", controllers.ReaderBook)
		authorized.GET("/reader/book/:id", controllers.ReaderBookDetail)
		authorized.GET("/reader/book/hot", controllers.BookHot)

		// 读者借阅相关路由
		authorized.GET("/reader/lend", controllers.ReaderLend)
		authorized.PUT("/reader/lend/:id", controllers.ReaderLendBook)
		authorized.PUT("/reader/return/:id", controllers.ReaderReturnBook)
		authorized.PUT("/reader/reserve/:id", controllers.ReaderReservationBook)
	}

	// 404页面
	r.NoRoute(controllers.NotFound)

	r.Run(":8080")
}
