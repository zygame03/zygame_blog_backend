package article

import (
	"context"
	"my_web/backend/internal/utils"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Service struct {
	DB  *Database
	RDB *Cache

	task utils.TaskRunner
}

func NewArticleService(ctx context.Context, db *gorm.DB, rdb *redis.Client) *Service {
	service := &Service{
		DB:  NewDatabase(db),
		RDB: NewArticleCache(rdb),
	}

	service.task = *utils.NewTaskRunner(
		service,
		utils.WithInterval(1*time.Hour),
		utils.WithTimeout(1*time.Minute),
	)

	service.task.Start(ctx)

	return service
}

func (s *Service) Run(ctx context.Context) {
	ids, err := s.DB.GetAllArticleIDs()
	if err != nil {
		return
	}

	for _, id := range ids {
		num, err := s.RDB.GetViewUV(ctx, id)
		if err != nil || num == 0 {
			continue
		}

		err = s.RDB.DelViewUV(ctx, id)
		if err != nil {
			continue
		}

		_ = s.DB.IncrementViews(id, num)
	}
}

// 分页查找
func (s *Service) GetArticlesByPage(ctx context.Context, page, pageSize int) ([]ArticleWithoutContent, int, error) {
	articles, total, err := s.RDB.GetArticlesByPage(ctx, page, pageSize)
	if err == nil {
		return articles, total, nil
	}

	if err == ErrCacheMiss {
		articles, total, err = s.DB.GetArticlesByPage(page, pageSize)
		if err != nil {
			return nil, 0, err
		}

		s.RDB.SetArticlesByPage(ctx, page, pageSize, articles, total)
		return articles, total, nil
	}

	return s.DB.GetArticlesByPage(page, pageSize)
}

// 获取热门文章，目前只基于view数，后续增加其他项综合判断
func (s *Service) GetArticlesByPopular(ctx context.Context, limit int) ([]ArticleWithoutContent, error) {
	articles, err := s.RDB.GetArticlesByPopular(ctx, limit)
	if err == nil {
		return articles, nil
	}

	if err == ErrCacheMiss {
		articles, err = s.DB.GetArticlesByPopular(limit)
		if err != nil {
			return nil, err
		}

		go s.RDB.SetArticlesByPopular(ctx, limit, articles)
		return articles, nil
	}

	return s.DB.GetArticlesByPopular(limit)
}

// 通过ID获取文章，获取后增加views
// userID: 用户标识，可以是用户ID或IP地址，用于防重复计数
func (s *Service) GetArticleByID(ctx context.Context, id int, userID string) (*Article, error) {
	// cache hit
	article, err := s.RDB.GetArticleByID(ctx, id)
	if err == nil {
		s.RDB.AddViewUV(ctx, id, userID)
		return article, nil
	} else {
		article, err = s.DB.GetArticleByID(id)
		if err != nil {
			return nil, err
		}

		// sync write-back to the DB and increasing views
		s.RDB.AddViewUV(ctx, id, userID)
		s.RDB.SetArticleByID(ctx, id, article)
		return article, nil
	}
}

// Todo 按tags找文章
func (s *Service) GetArticlesByTag(limit int) ([]Article, error) {
	return nil, nil
}
