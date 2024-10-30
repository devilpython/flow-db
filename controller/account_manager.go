package controller

import (
	"fmt"

	"github.com/devilpython/devil-db/db/model"
	"github.com/devilpython/devil-db/db/model_action"
	"github.com/devilpython/devil-db/db/sql_interface"
	"github.com/devilpython/devil-db/db/sql_utils"
	"github.com/devilpython/devil-tools/cache"
	devil "github.com/devilpython/devil-tools/utils"
	"github.com/devilpython/flow-db/constants"
	"github.com/devilpython/flow-db/global_keys"
	"github.com/gin-gonic/gin"
)

// 注册动作
func Register(context *gin.Context) {
	dataMap, message, validSuccess := getDataMapParam()
	if validSuccess {
		accountModel, hasModel := model.GetModel("account")
		if hasModel && accountModel.PrimaryKey.IsPrimaryKey {
			delete(dataMap, accountModel.PrimaryKey.Name)
			message, validSuccess = model_action.SaveData("account", dataMap)
		} else {
			message, _ = devil.GetMessage("account-table-error")
		}
	}

	if validSuccess {
		devil.SetGlobalData(global_keys.DbMustCommit, true)
		message, _ := devil.GetMessage("register-message")
		devil.ShowMessage(context, true, message)
	} else {
		devil.ShowMessage(context, false, message)
	}
}

// 注销账号动作
func CancellationOfAccount(context *gin.Context) {
	dataMap, message, validSuccess := getDataMapParam()
	if validSuccess {
		ticketObj, hasTicket := dataMap["ticket"]
		if !hasTicket {
			message, _ := devil.GetMessage("no-ticket")
			devil.ShowMessage(context, false, message)
			return
		}
		ticket := fmt.Sprintf("%v", ticketObj)
		key := fmt.Sprintf(constants.TicketKey, ticket)
		loginInfo := make(map[string]interface{})
		err := cache.GetObject(key, &loginInfo)
		if err != nil {
			message, _ := devil.GetMessage("ticket-error")
			devil.ShowMessage(context, false, message)
			return
		}
		message, validSuccess = model_action.RemoveData("account", loginInfo)
		if validSuccess {
			cache.Remove(key)
		}
	}

	if validSuccess {
		devil.SetGlobalData(global_keys.DbMustCommit, true)
		message, _ := devil.GetMessage("cancellation-message")
		devil.ShowMessage(context, true, message)
	} else {
		devil.ShowMessage(context, false, message)
	}
}

// 登录账号动作
func Login(context *gin.Context) {
	dataMap, message, validSuccess := getDataMapParam()
	accountManager, hasManager := model.GetAccountManager()
	if validSuccess && hasManager {
		action := sql_utils.Action{}
		message, validSuccess = executeValidater(accountManager.AccountModel, action, sql_interface.ModelPermissionsOperationTypeQuery, dataMap)
		if validSuccess {
			executeOperator(accountManager.AccountModel, action, sql_interface.ModelPermissionsOperationTypeQuery, dataMap)
			var loginInfoArray []map[string]interface{}
			loginInfoArray, message, validSuccess = model_action.QueryData("account", dataMap)
			if validSuccess && len(loginInfoArray) == 1 {
				loginInfo := loginInfoArray[0]
				ticket := devil.CreateId()
				loginInfo["ticket"] = ticket
				key := fmt.Sprintf(constants.TicketKey, ticket)
				adminParam := make(map[string]interface{})
				adminParam["account_id"] = loginInfo["id"]
				_, _, isAdmin := model_action.GetData("account", adminParam)
				loginInfo["is_admin"] = isAdmin
				cache.SetObjectEx(key, loginInfo, 60*60*4)
				delete(loginInfo, "id")
				message, _ := devil.GetMessage("login-message")
				devil.ShowDataMessage(context, true, message, loginInfo)
			} else {
				message, _ = devil.GetMessage("nil-account")
			}
		}
	}
	if !validSuccess {
		devil.ShowMessage(context, false, message)
	}
}

// 登录账号动作
func Logout(context *gin.Context) {
	dataMap, message, validSuccess := getDataMapParam()
	if validSuccess {
		ticketObj, hasTicket := dataMap["ticket"]
		if !hasTicket {
			message, _ := devil.GetMessage("no-ticket")
			devil.ShowMessage(context, false, message)
			return
		}
		ticket := fmt.Sprintf("%v", ticketObj)
		key := fmt.Sprintf(constants.TicketKey, ticket)
		loginInfo := make(map[string]interface{})
		err := cache.GetObject(key, &loginInfo)
		if err != nil {
			message, _ := devil.GetMessage("ticket-error")
			devil.ShowMessage(context, false, message)
			return
		}
		cache.Remove(key)
		message, _ := devil.GetMessage("logout-message")
		devil.ShowMessage(context, true, message)
	} else {
		devil.ShowMessage(context, false, message)
	}
}

// 修改账号密码
func ModifyPassword(context *gin.Context) {
	dataMap, message, validSuccess := getDataMapParam()
	accountManager, hasManager := model.GetAccountManager()
	if validSuccess && hasManager {
		ticketObj, hasTicket := dataMap["ticket"]
		if !hasTicket {
			message, _ := devil.GetMessage("no-ticket")
			devil.ShowMessage(context, false, message)
			return
		}
		ticket := fmt.Sprintf("%v", ticketObj)
		key := fmt.Sprintf(constants.TicketKey, ticket)
		loginInfo := make(map[string]interface{})
		err := cache.GetObject(key, &loginInfo)
		if err != nil {
			message, _ := devil.GetMessage("ticket-error-must-login")
			devil.ShowMessage(context, false, message)
			return
		}
		passwordObj, hasPassword := dataMap["password"]
		if !hasPassword || passwordObj == nil {
			message, _ := devil.GetMessage("nil-password")
			devil.ShowMessage(context, false, message)
			return
		}
		oldPasswordObj, hasOldPassword := dataMap["old_password"]
		if !hasOldPassword {
			message, _ := devil.GetMessage("nil-old-password")
			devil.ShowMessage(context, false, message)
			return
		}
		oldPassword := fmt.Sprintf("%v", oldPasswordObj)
		_, hasId := loginInfo["id"]
		if !hasId {
			message, _ := devil.GetMessage("ticket-data-error")
			devil.ShowMessage(context, false, message)
			return
		}
		paramMap := make(map[string]interface{})
		paramMap[accountManager.AccountModel.Id] = loginInfo[accountManager.AccountModel.Id]
		paramMap[accountManager.AccountModel.Password] = oldPassword
		var loginInfoArray []map[string]interface{}
		loginInfoArray, message, validSuccess = model_action.QueryData("account", paramMap)
		if validSuccess && len(loginInfoArray) == 1 {
			dataMap["id"] = loginInfo["id"]
			message, validSuccess = model_action.SaveData("account", dataMap)
			if validSuccess {
				devil.SetGlobalData(global_keys.DbMustCommit, true)
				message, _ := devil.GetMessage("password-modify-success")
				devil.ShowMessage(context, true, message)
			} else {
				message, _ := devil.GetMessage("db-error")
				devil.ShowMessage(context, false, message)
			}
		} else {
			message, _ := devil.GetMessage("password-error")
			devil.ShowMessage(context, false, message)
			return
		}
	} else {
		devil.ShowMessage(context, false, message)
	}
}

// 修改账号信息
func ModifyAccountInfo(context *gin.Context) {
	dataMap, message, validSuccess := getDataMapParam()
	accountManager, hasManager := model.GetAccountManager()
	if validSuccess && hasManager {
		ticketObj, hasTicket := dataMap["ticket"]
		if !hasTicket {
			message, _ := devil.GetMessage("no-ticket")
			devil.ShowMessage(context, false, message)
			return
		}
		ticket := fmt.Sprintf("%v", ticketObj)
		key := fmt.Sprintf(constants.TicketKey, ticket)
		loginInfo := make(map[string]interface{})
		err := cache.GetObject(key, &loginInfo)
		if err != nil {
			message, _ := devil.GetMessage("ticket-error-must-login")
			devil.ShowMessage(context, false, message)
			return
		}
		passwordObj, hasPassword := dataMap["password"]
		if !hasPassword {
			message, _ := devil.GetMessage("nil-password")
			devil.ShowMessage(context, false, message)
			return
		}
		password := fmt.Sprintf("%v", passwordObj)
		_, hasId := loginInfo["id"]
		if !hasId {
			message, _ := devil.GetMessage("ticket-data-error")
			devil.ShowMessage(context, false, message)
			return
		}
		paramMap := make(map[string]interface{})
		paramMap[accountManager.AccountModel.Id] = loginInfo[accountManager.AccountModel.Id]
		paramMap[accountManager.AccountModel.Password] = password
		var loginInfoArray []map[string]interface{}
		loginInfoArray, message, validSuccess = model_action.QueryData("account", paramMap)
		if validSuccess && len(loginInfoArray) == 1 {
			dataMap["id"] = loginInfo["id"]
			message, validSuccess = model_action.SaveData("account", dataMap)
			if validSuccess {
				devil.SetGlobalData(global_keys.DbMustCommit, true)
				message, _ := devil.GetMessage("account-modify-success")
				devil.ShowMessage(context, true, message)
			} else {
				//message, _ := devil.GetMessage("db-error")
				devil.ShowMessage(context, false, message)
			}
		} else {
			message, _ := devil.GetMessage("password-error")
			devil.ShowMessage(context, false, message)
			return
		}
	} else {
		devil.ShowMessage(context, false, message)
	}
}

// 获得账号Token
func GetToken(context *gin.Context) {
	dataMap, message, validSuccess := getDataMapParam()
	if validSuccess {
		ticketObj, hasTicket := dataMap["ticket"]
		if !hasTicket {
			message, _ := devil.GetMessage("no-ticket")
			devil.ShowMessage(context, false, message)
			return
		}
		ticket := fmt.Sprintf("%v", ticketObj)
		key := fmt.Sprintf(constants.TicketKey, ticket)
		loginInfo := make(map[string]interface{})
		err := cache.GetObject(key, &loginInfo)
		if err != nil {
			message, _ := devil.GetMessage("ticket-error")
			devil.ShowMessage(context, false, message)
			return
		}
		id, hasId := loginInfo["id"]
		if hasId {
			paramMap := make(map[string]interface{})
			paramMap["account_id"] = id
			var tokenInfo map[string]interface{}
			tokenInfo, message, validSuccess = model_action.GetData("token", paramMap)
			var token interface{}
			if validSuccess {
				hasToken := false
				token, hasToken = tokenInfo["token"]
				if !hasToken {
					message, validSuccess = model_action.UpdateData("token", paramMap)
					token, _ = paramMap["token"]
				}
			} else {
				message, validSuccess = model_action.InsertData("token", paramMap)
				token, _ = paramMap["token"]
			}
			if validSuccess {
				devil.SetGlobalData(global_keys.DbMustCommit, true)
				context.JSON(200, gin.H{
					"successful": true,
					"message":    "ok",
					"token":      token,
				})
			} else {
				message, _ := devil.GetMessage("db-error")
				devil.ShowMessage(context, false, message)
			}
		} else {
			message, _ := devil.GetMessage("ticket-data-error")
			devil.ShowMessage(context, false, message)
		}
	} else {
		devil.ShowMessage(context, false, message)
	}
}

// 超管获得账号Token
func GetTokenForAdmin(context *gin.Context) {
	dataMap, message, validSuccess := getDataMapParam()
	if validSuccess {
		accountIdObj, hasAccountId := dataMap["account_id"]
		if !hasAccountId {
			message, _ := devil.GetMessage("no-account-id")
			devil.ShowMessage(context, false, message)
			return
		}
		accountParamMap := make(map[string]interface{})
		accountParamMap["id"] = accountIdObj
		_, message, validSuccess = model_action.GetData("account", accountParamMap)
		if validSuccess {
			var token interface{}
			delete(dataMap, "token")
			tokenInfo := make(map[string]interface{})
			tokenInfo, message, validSuccess = model_action.GetData("token", dataMap)
			if validSuccess {
				token, _ = tokenInfo["token"]
			} else {
				message, validSuccess = model_action.InsertData("token", dataMap)
				token, _ = dataMap["token"]
			}
			if validSuccess {
				context.JSON(200, gin.H{
					"successful": true,
					"message":    "ok",
					"token":      token,
				})
			} else {
				message, _ := devil.GetMessage("db-error")
				devil.ShowMessage(context, false, message)
			}
		} else {
			message, _ := devil.GetMessage("no-account")
			devil.ShowMessage(context, false, message)
		}
	} else {
		devil.ShowMessage(context, false, message)
	}
}

// 更新账号Token
func UpdateToken(context *gin.Context) {
	dataMap, message, validSuccess := getDataMapParam()
	if validSuccess {
		ticketObj, hasTicket := dataMap["ticket"]
		if !hasTicket {
			message, _ := devil.GetMessage("no-ticket")
			devil.ShowMessage(context, false, message)
			return
		}
		ticket := fmt.Sprintf("%v", ticketObj)
		key := fmt.Sprintf(constants.TicketKey, ticket)
		loginInfo := make(map[string]interface{})
		err := cache.GetObject(key, &loginInfo)
		if err != nil {
			message, _ := devil.GetMessage("ticket-error-must-login")
			devil.ShowMessage(context, false, message)
			return
		}
		id, hasId := loginInfo["id"]
		if hasId {
			paramMap := make(map[string]interface{})
			paramMap["account_id"] = id
			_, _, hasData := model_action.GetData("token", paramMap)
			if hasData {
				message, validSuccess = model_action.UpdateData("token", paramMap)
			} else {
				message, validSuccess = model_action.InsertData("token", paramMap)
			}
			if validSuccess {
				devil.SetGlobalData(global_keys.DbMustCommit, true)
				context.JSON(200, gin.H{
					"successful": true,
					"message":    "ok",
					"token":      paramMap["token"],
				})
			} else {
				message, _ := devil.GetMessage("db-error")
				devil.ShowMessage(context, false, message)
			}
		} else {
			message, _ := devil.GetMessage("ticket-data-error")
			devil.ShowMessage(context, false, message)
		}
	} else {
		devil.ShowMessage(context, false, message)
	}
}

// 超管更新账号Token
func UpdateTokenForAdmin(context *gin.Context) {
	dataMap, message, validSuccess := getDataMapParam()
	if validSuccess {
		accountIdObj, hasAccountId := dataMap["account_id"]
		if !hasAccountId {
			message, _ := devil.GetMessage("no-account-id")
			devil.ShowMessage(context, false, message)
			return
		}
		accountParamMap := make(map[string]interface{})
		accountParamMap["id"] = accountIdObj
		_, message, validSuccess = model_action.GetData("account", accountParamMap)
		if validSuccess {
			delete(dataMap, "token")
			message, validSuccess = model_action.SaveData("token", dataMap)
			if validSuccess {
				devil.SetGlobalData(global_keys.DbMustCommit, true)
				context.JSON(200, gin.H{
					"successful": true,
					"message":    "ok",
					"token":      dataMap["token"],
				})
			} else {
				message, _ := devil.GetMessage("db-error")
				devil.ShowMessage(context, false, message)
			}
		} else {
			message, _ := devil.GetMessage("no-account")
			devil.ShowMessage(context, false, message)
		}

	} else {
		devil.ShowMessage(context, false, message)
	}
}

// 获得账号信息
func GetAccountInfo(context *gin.Context) {
	dataMap, message, validSuccess := getDataMapParam()
	if validSuccess {
		ticketObj, hasTicket := dataMap["ticket"]
		if !hasTicket {
			message, _ := devil.GetMessage("no-ticket")
			devil.ShowMessage(context, false, message)
			return
		}
		ticket := fmt.Sprintf("%v", ticketObj)
		key := fmt.Sprintf(constants.TicketKey, ticket)
		loginInfo := make(map[string]interface{})
		err := cache.GetObject(key, &loginInfo)
		if err != nil {
			message, _ := devil.GetMessage("ticket-error")
			devil.ShowMessage(context, false, message)
			return
		}
		devil.ShowDataMessage(context, true, "ok", loginInfo)
	} else {
		devil.ShowMessage(context, false, message)
	}
}

// 根据Token获得账号信息
func GetAccountInfoForToken(context *gin.Context) {
	paramMap := make(map[string]interface{})
	accountModel, hasModel := model.GetModel("account")
	if hasModel {
		paramMap[accountModel.PrimaryKey.Name], _ = devil.GetGlobalData(global_keys.KeyAccountId)
		dataMap, message, validSuccess := model_action.GetData("account", paramMap)
		if validSuccess {
			devil.ShowDataMessage(context, true, "ok", dataMap)
		} else {
			devil.ShowMessage(context, false, message)
		}
	} else {
		message, _ := devil.GetMessage("account-table-error")
		devil.ShowMessage(context, false, message)
	}
}

// 执行操作器
func executeOperator(modelObj model.AccountModel, action sql_utils.Action, operationType int, dataMap map[string]interface{}) {
	accountModel, hasModel := model.GetModel("account")
	if hasModel {
		//执行操作器
		for operatorIndex := range modelObj.OperationArray {
			modelObj.OperationArray[operatorIndex].Operate(dataMap, action, accountModel.Nick, accountModel.PrimaryKey.Name, operationType)
		}
	}
}

// 执行验证器
func executeValidater(modelObj model.AccountModel, action sql_utils.Action, operationType int, dataMap map[string]interface{}) (string, bool) {
	accountModel, hasModel := model.GetModel("account")
	if hasModel {
		//执行验证器
		for validaterIndex := range modelObj.ValidaterArray {
			message, successful := modelObj.ValidaterArray[validaterIndex].Validate(dataMap, action, accountModel.PrimaryKey.Name, operationType)
			if !successful {
				return message, false
			}
		}
	}
	return "", true
}
