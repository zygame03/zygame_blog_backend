package httpserver

import (
	"my_web/backend/internal/config"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Router interface {
	RegisterRoutes(*gin.Engine)
}

func NewHttpserver(conf *config.HttpserverConfig, routers ...Router) *http.Server {
	e := gin.New()

	e.Use(cors.New(cors.Config{
		AllowOrigins:     conf.Cors.AllowedOrigins,
		AllowMethods:     conf.Cors.AllowedMethods,
		AllowHeaders:     conf.Cors.AllowedHeaders,
		ExposeHeaders:    conf.Cors.ExposeHeaders,
		AllowCredentials: conf.Cors.AllowCredentials,
		MaxAge:           time.Duration(conf.Cors.MaxAge) * time.Hour,
	}))

	for _, r := range routers {
		r.RegisterRoutes(e)
	}

	return &http.Server{
		Addr:    conf.Port,
		Handler: e,
	}
}
