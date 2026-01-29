package article

import (
	"my_web/backend/internal/httpserver"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	httpserver.BaseHandler
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{
		service: s,
	}
}

func (h *Handler) RegisterRoutes(e *gin.Engine) {
	r := e.Group("/api/article")
	{
		r.GET("", h.getArticles)
		r.GET("/hotArticles", h.getHotArticles)
		r.GET("/:id", h.getArticleDetail)
	}
}

// 获取文章列表
func (h *Handler) getArticles(ctx *gin.Context) {
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil {
		h.Fail(ctx, httpserver.ErrRequest, err)
		return
	}

	pageSize, err := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	if err != nil {
		h.Fail(ctx, httpserver.ErrRequest, err)
		return
	}

	articles, total, err := h.service.GetArticlesByPage(ctx.Request.Context(), page, pageSize)
	if err != nil {
		h.Fail(ctx, httpserver.ErrDBOp, err)
		return
	}

	h.Success(ctx, httpserver.PageResult[ArticleWithoutContent]{
		Page:  page,
		Size:  pageSize,
		Total: total,
		Data:  articles,
	})
}

func (h *Handler) getHotArticles(ctx *gin.Context) {
	data, err := h.service.GetArticlesByPopular(ctx, 10)
	if err != nil {
		h.Fail(ctx, httpserver.ErrDBOp, err)
		return
	}

	h.Success(ctx, data)
}

// 获取文章详情（带正文）
func (h *Handler) getArticleDetail(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.Fail(ctx, httpserver.ErrRequest, err)
		return
	}

	// 获取用户标识（优先使用用户ID，否则使用IP地址）
	userID := ctx.GetString("userID") // 如果中间件设置了用户ID
	if userID == "" {
		userID = ctx.ClientIP() // 使用IP地址作为标识
	}

	data, err := h.service.GetArticleByID(ctx.Request.Context(), id, userID)
	if err != nil {
		h.Fail(ctx, httpserver.ErrDBOp, err)
		return
	}

	h.Success(ctx, data)
}
