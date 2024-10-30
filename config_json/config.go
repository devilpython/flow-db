package config_json

import (
	"log"
	"sync"

	"github.com/devilpython/flow-db/constants"

	"github.com/devilpython/devil-tools/utils"
)

// 配置结构体
type Configuration struct {
	DbServerUri string
	ServerPort  int
}

// 配置对象
var config Configuration
var once sync.Once
var configOk = false

// 获得配置结构体实例
func GetConfigInstance() (Configuration, bool) {
	once.Do(func() {
		dataMap := utils.GetConfigMap(constants.ConfigFilePath)
		if dataMap != nil {
			_config, err := utils.ConvertMapToStruct(*dataMap, Configuration{})
			if err != nil {
				log.Println("config_json error:", err)
			} else {
				c, ok := _config.(*Configuration)
				if ok {
					config = *c
					configOk = true
				}
			}
		}
	})
	return config, configOk
}
