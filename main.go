package main

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/devilpython/flow-db/config_json"
	"github.com/devilpython/flow-db/controller"
	"github.com/devilpython/flow-db/middle_ware"

	devilMiddle "github.com/devilpython/devil-tools/middle_ware"
	gin_cors "github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	conf, ok := config_json.GetConfigInstance()
	if !ok {
		fmt.Println("配置文件加载失败")
	} else {
		port := flag.Int("port", conf.ServerPort, "端口号")
		flag.Parse()

		router := gin.Default()
		//router.Use(gin_cors.Default(), logger.GinLogger())
		router.Use(gin_cors.Default(), devilMiddle.PostDataBinding())
		router.GET("/", controller.ShowVersion)
		router.GET("/favicon.ico", controller.ShowFavicon)
		router.Use(middle_ware.OpenDataBase())

		//数据库相关操作
		dbGroup := router.Group("/db")
		{
			dbGroup.Use(middle_ware.AuthorizeForUser())
			dbGroup.GET("/:nick/save.do", controller.Save)
			dbGroup.POST("/:nick/save.do", controller.Save)
			dbGroup.GET("/:nick/remove.do", controller.Remove)
			dbGroup.POST("/:nick/remove.do", controller.Remove)
			dbGroup.GET("/:nick/get.do", controller.Get)
			dbGroup.POST("/:nick/get.do", controller.Get)
			dbGroup.GET("/:nick/list.do", controller.Query)
			dbGroup.POST("/:nick/list.do", controller.Query)
			dbGroup.GET("/:nick/search.do", controller.Query)
			dbGroup.POST("/:nick/search.do", controller.Query)
		}

		//token相关操作
		tokenGroup := router.Group("/token")
		{
			tokenGroup.Use(middle_ware.AuthorizeForUser())
			tokenGroup.GET("/get-account-info.do", controller.GetAccountInfoForToken)
			tokenGroup.POST("/get-account-info.do", controller.GetAccountInfoForToken)
		}

		//账号相关操作
		accountGroup := router.Group("/account")
		{
			accountGroup.GET("/register.do", controller.Register)
			accountGroup.POST("/register.do", controller.Register)
			accountGroup.GET("/cancellation-of-account.do", controller.CancellationOfAccount)
			accountGroup.POST("/cancellation-of-account.do", controller.CancellationOfAccount)
			accountGroup.GET("/login.do", controller.Login)
			accountGroup.POST("/login.do", controller.Login)
			accountGroup.GET("/logout.do", controller.Logout)
			accountGroup.POST("/logout.do", controller.Logout)
			accountGroup.GET("/modify-password.do", controller.ModifyPassword)
			accountGroup.POST("/modify-password.do", controller.ModifyPassword)
			accountGroup.GET("/modify-account-info.do", controller.ModifyAccountInfo)
			accountGroup.POST("/modify-account-info.do", controller.ModifyAccountInfo)
			accountGroup.GET("/get-token.do", controller.GetToken)
			accountGroup.POST("/get-token.do", controller.GetToken)
			accountGroup.GET("/update-token.do", controller.UpdateToken)
			accountGroup.POST("/update-token.do", controller.UpdateToken)
			accountGroup.GET("/get-account-info.do", controller.GetAccountInfo)
			accountGroup.POST("/get-account-info.do", controller.GetAccountInfo)
		}

		//管理员相关操作
		adminGroup := router.Group("/admin")
		{
			dbGroup.Use(middle_ware.AuthorizeForAdmin())
			adminGroup.GET("/get-token.do", controller.GetTokenForAdmin)
			adminGroup.POST("/get-token.do", controller.GetTokenForAdmin)
			adminGroup.GET("/update-token.do", controller.UpdateTokenForAdmin)
			adminGroup.POST("/update-token.do", controller.UpdateTokenForAdmin)
		}

		// 指定地址和端口号
		err := router.Run("0.0.0.0:" + strconv.Itoa(*port))
		if err != nil {
			fmt.Println("Error:", err)
		}
	}
}
