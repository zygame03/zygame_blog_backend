package config

import (
	"log"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	Httpserver HttpserverConfig `mapstructure:"httpserver"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Redis      RedisConfig      `mapstructure:"redis"`
}

type HttpserverConfig struct {
	Port string      `mapstructure:"port"`
	Cors *CorsConfig `mapstructure:"cors"`
}

// CorsConfig CORS 配置
type CorsConfig struct {
	AllowedOrigins   []string `mapstructure:"allowedOrigins"`
	AllowedMethods   []string `mapstructure:"allowedMethods"`
	AllowedHeaders   []string `mapstructure:"allowedHeaders"`
	ExposeHeaders    []string `mapstructure:"exposeHeaders"`
	AllowCredentials bool     `mapstructure:"allowCredentials"`
	MaxAge           int      `mapstructure:"maxAge"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	Port     uint   `mapstructure:"port"`
	SSLMode  string `mapstructure:"sslmode"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	Protocol int    `mapstructure:"protocol"`
}

// ReadConfig 读取配置文件
func ReadConfig(fpath, fname, ftype string) (*Config, error) {
	viper.Reset()

	viper.AddConfigPath(fpath)
	viper.SetConfigName(fname)
	viper.SetConfigType(ftype)
	if err := viper.ReadInConfig(); err != nil {
		log.Println("读取配置文件失败", fpath, fname, ftype, err)
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Println("解析配置文件失败", err)
		return nil, err
	}

	log.Println("配置读取成功")
	return &cfg, nil
}
