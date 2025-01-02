package main

import (
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"onosutil/model"
	"onosutil/utils"
	"onosutil/utils/errors"
	"path/filepath"
)

func init() {
	if err := setupConfig(); err != nil {
		log.Fatal(err)
	}
}

func setupConfig() error {
	configPath := pflag.StringP("config", "c", "./config.yaml", "configuration file path")
	// 数据库默认配置
	viper.SetDefault("db.name", "database")
	viper.SetDefault("db.host", "localhost")
	viper.SetDefault("db.port", 3306)
	viper.SetDefault("db.user", "username")
	viper.SetDefault("db.password", "password")
	// todo: 其他配置
	configDir, configFile := filepath.Split(*configPath)
	return flushConfig(configDir, configFile)
}

// flushConfig 初始化配置项
func flushConfig(configDir, configFile string) error {
	viper.AddConfigPath(configDir)
	viper.SetConfigFile(configFile)
	fullpath := filepath.Join(configDir, configFile)
	if utils.FileExists(fullpath) { // 如果用户提供了配置文件，则合并
		err := viper.MergeInConfig()
		if err != nil {
			log.Error("flushConfig error in MergeInConfig err: ", err)
			return err
		}
	} else {
		err := viper.SafeWriteConfigAs(fullpath)
		if err != nil {
			log.Error("flushConfig error in SafeWriteConfigAs err: ", err)
			return err
		}
	}
	// 配置热更新
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
	})
	viper.WatchConfig()
	return nil
}

func SetupORM(config *viper.Viper) (orm.Ormer, error) {
	if config == nil {
		return nil, errors.New(errors.CodeDataBaseConfigEmpty, "database config is nil")
	}
	orm.RegisterModel(new(model.Device))
	orm.RegisterModel(new(model.Link))
	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4", config.GetString("user"),
		config.GetString("password"), config.GetString("host"),
		config.GetString("port"), config.GetString("database"))
	if err := orm.RegisterDataBase("default", "mysql", dataSource); err != nil {
		return nil, errors.New(errors.CodeRegisterDatabaseFailed, "ORM RegisterDataBase failed")
	}
	if err := orm.RunSyncdb("default", false, true); err != nil {
		return nil, errors.New(errors.CodeRunSyncdbFailed, "ORM RunSyncdb failed")
	}
	return orm.NewOrm(), nil
}
