package article

import (
	"context"
	"my_web/backend/internal/utils"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Service struct {
	DB  *gorm.DB
	RDB *redis.Client

	task utils.TaskRunner
}

func NewArticleService(ctx context.Context, db *gorm.DB, rdb *redis.Client) *Service {
	service := &Service{
		DB:  db,
		RDB: rdb,
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
	ids, err := repoGetAllArticleIDs(s.DB)
	if err != nil {
		return
	}

	for _, id := range ids {
		num, err := cacheGetViewUV(ctx, s.RDB, id)
		if err != nil || num == 0 {
			continue
		}

		err = cacheDelViewUV(ctx, s.RDB, id)
		if err != nil {
			continue
		}

		_ = repoIncrementViews(s.DB, id, num)
	}
}

// 分页查找
func (s *Service) GetArticlesByPage(ctx context.Context, page, pageSize int) ([]ArticleWithoutContent, int, error) {
	articles, total, err := cacheGetArticlesByPage(ctx, s.RDB, page, pageSize)
	if err == nil {
		return articles, total, nil
	}

	if err == ErrCacheMiss {
		articles, total, err = repoGetArticlesByPage(s.DB, page, pageSize)
		if err != nil {
			return nil, 0, err
		}

		cacheSetArticlesByPage(ctx, s.RDB, page, pageSize, articles, total)
		return articles, total, nil
	}

	return repoGetArticlesByPage(s.DB, page, pageSize)
}

// 获取热门文章，目前只基于view数，后续增加其他项综合判断
func (s *Service) GetArticlesByPopular(ctx context.Context, limit int) ([]ArticleWithoutContent, error) {
	articles, err := cacheGetArticlesByPopular(ctx, s.RDB, limit)
	if err == nil {
		return articles, nil
	}

	if err == ErrCacheMiss {
		articles, err = repoGetArticlesByPopular(s.DB, limit)
		if err != nil {
			return nil, err
		}

		go cacheSetArticlesByPopular(ctx, s.RDB, limit, articles)
		return articles, nil
	}

	return repoGetArticlesByPopular(s.DB, limit)
}

// 通过ID获取文章，获取后增加views
// userID: 用户标识，可以是用户ID或IP地址，用于防重复计数
func (s *Service) GetArticleByID(ctx context.Context, id int, userID string) (*Article, error) {
	// cache hit
	article, err := cacheGetArticleByID(ctx, s.RDB, id)
	if err == nil {
		cacheAddViewUV(ctx, s.RDB, id, userID)
		return article, nil
	} else {
		article, err = repoGetArticleByID(s.DB, id)
		if err != nil {
			return nil, err
		}

		// sync write-back to the DB and increasing views
		cacheAddViewUV(ctx, s.RDB, id, userID)
		cacheSetArticleByID(ctx, s.RDB, id, article)
		return article, nil
	}
}

// Todo 按tags找文章
func (s *Service) GetArticlesByTag(limit int) ([]Article, error) {
	return nil, nil
}
