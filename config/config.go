package config

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
)

var baseConfig *Config

type Config struct {
	Nginx        int    `json:"nginx"`
	Redispass    string `mapstructure:"redispass"`
	ListenDomain string `mapstructure:"listendomain"`
	Dev          int    `mapstructure:"dev"`
	Database     struct {
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		DBName   string `mapstructure:"dbname"`
	} `mapstructure:"database"`
	Server struct {
		Defaultip   string `mapstructure:"default_ip"`
		Subdomain   string `mapstructure:"subdomain"`
		Port        string `mapstructure:"http_port"`
		Admindomain string `mapstructure:"admin_domain"`
		Adminport   string `mapstructure:"admin_port"`
		Seckey      string `mapstructure:"seckey"`
		SSL         struct {
			Enabled  bool `mapstructure:"enabled"`
			CertFile bool `mapstructure:"cert_file"`
			KeyFile  bool `mapstructure:"key_file"`
		} `mapstructure:"ssl""`
	} `mapstructure:"server"`
	Sqldebug int `mapstructure:"sqldebug"`
}

func GetBase() *Config {
	return baseConfig
}

func Init() error {
	if os.Getenv("env") == "test" {
		viper.SetConfigFile("config-test.yaml")
	} else {
		viper.SetConfigFile("config.yaml")
	}
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	logrus.Info("load config")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("无法读取配置文件:", err)
		return err
	}
	logrus.Info("read config")
	if err := viper.Unmarshal(&baseConfig); err != nil {
		fmt.Println("无法解析配置文件:", err)
		return err
	}

	return nil
}
