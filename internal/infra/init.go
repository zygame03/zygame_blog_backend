package infra

import (
	"fmt"
	"log"
	"my_web/backend/internal/article"
	"my_web/backend/internal/config"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// InitDatabase 初始化数据库连接
func InitDatabase(conf *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		conf.Host, conf.User, conf.Password, conf.DBName, conf.Port, conf.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 自动迁移
	if err := db.AutoMigrate(
		&article.Article{},
	); err != nil {
		return nil, fmt.Errorf("数据库自动迁移失败: %w", err)
	}

	log.Println("数据库初始化成功")
	return db, nil
}

// InitRedis 初始化 Redis 连接
func InitRedis(conf *config.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       conf.DB,
		Protocol: conf.Protocol,
	})

	log.Println("Redis 初始化成功")
	return rdb, nil
}
