package config

import (
	"github.com/spf13/viper"
	"log"
)

type Config struct { //无所谓大小写
	App struct {
		Name string
		Port string
	}
	Database struct {
		Dsn           string
		MaxIdleConnes int
		MaxOpenConnes int
	}
}

var AppConfig = &Config{}

func InitConfig() {
	//Viper是一种配置文件管理器
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath("./config")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("err reading config , err: %v", err)
	}
	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatalf("err unmarshal config , err: %v", err)
	}
	InitDB()
	InitRedis()
}
