package main

import (
	"context"
	"log"
	"my_web/backend/internal/article"
	"my_web/backend/internal/config"
	"my_web/backend/internal/httpserver"
	"my_web/backend/internal/infra"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 读取配置
	config, err := config.ReadConfig("config/", "config", "json")
	if err != nil {
		log.Fatalf("读取配置失败: %v", err)
	}

	// 初始化应用依赖
	db, err := infra.InitDatabase(&config.Database)
	if err != nil {
		return
	}

	rdb, err := infra.InitRedis(&config.Redis)
	if err != nil {
		return
	}

	ctx := context.Background()
	articleServ := article.NewArticleService(ctx, db, rdb)
	articleHandler := article.NewHandler(articleServ)

	// 在 goroutine 中启动服务
	srv := httpserver.NewHttpserver(
		&config.Httpserver,
		articleHandler,
	)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 优雅退出处理
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("正在关闭服务...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("服务已退出")
}
