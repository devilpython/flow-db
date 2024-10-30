package middle_ware

import (
	"github.com/devilpython/devil-db/db/model_account"
	"github.com/devilpython/devil-tools/middle_ware"
	devil "github.com/devilpython/devil-tools/utils"
	"github.com/devilpython/flow-db/config_json"
	"github.com/devilpython/flow-db/global_keys"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/xormplus/xorm"
)

// 打开数据库
func OpenDataBase() gin.HandlerFunc {
	return func(context *gin.Context) {
		conf, _ := config_json.GetConfigInstance()
		engine, err := xorm.NewEngine("mysql", conf.DbServerUri)
		if err != nil {
			devil.ShowErrorMessage(context, err.Error())
			return
		}
		defer func() {
			_ = engine.Close()
		}()
		// 访问令牌
		_, _ = engine.Transaction(func(session *xorm.Session) (interface{}, error) {
			devil.SetGlobalData(global_keys.KeyDbSession, session)
			context.Next()
			mustCommitData, hasData := devil.GetGlobalData(global_keys.DbMustCommit)
			if hasData {
				mustCommit, ok := mustCommitData.(bool)
				if ok && mustCommit {
					_ = session.Commit()
				} else {
					_ = session.Rollback()
				}
			} else {
				_ = session.Rollback()
			}
			return nil, nil
		})
	}
}

// 用户权限验证
func AuthorizeForUser() gin.HandlerFunc {
	return func(context *gin.Context) {
		token := getToken(context)
		if len(token) > 0 {
			// 访问令牌
			accountId, isAdmin, msg := model_account.GetAccountIdForToken(token)
			if len(accountId) > 0 {
				devil.SetGlobalData(global_keys.KeyIsAdmin, isAdmin)
				devil.SetGlobalData(global_keys.KeyAccountId, accountId)
				context.Next()
			} else {
				// 验证不通过，不再调用后续的函数处理
				devil.ShowErrorMessage(context, msg)
			}
		} else {
			message, _ := devil.GetMessage("program-error")
			devil.ShowErrorMessage(context, message)
		}
	}
}

// 管理员权限验证
func AuthorizeForAdmin() gin.HandlerFunc {
	return func(context *gin.Context) {
		adminObj, hasData := devil.GetGlobalData(global_keys.KeyIsAdmin)
		if hasData {
			isAdmin, _ := adminObj.(bool)
			if isAdmin {
				context.Next()
			} else {
				message, _ := devil.GetMessage("admin-error")
				devil.ShowErrorMessage(context, message)
			}
		} else {
			// 验证不通过，不再调用后续的函数处理
			message, _ := devil.GetMessage("admin-error")
			devil.ShowErrorMessage(context, message)
		}
	}
}

// 获得Token
func getToken(context *gin.Context) string {
	data, ok := devil.GetGlobalData(middle_ware.KeyPostData)
	if ok {
		dataMap, hasData := data.(map[string]interface{})
		if hasData {
			token, existToken := dataMap["token"]
			if existToken {
				tokenStr, tokenOk := token.(string)
				if tokenOk {
					return tokenStr
				}
			}
		}
	}
	return ""
}
