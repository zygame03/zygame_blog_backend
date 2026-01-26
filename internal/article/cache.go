package article

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrCacheMiss = errors.New("cache miss")
)

type Cache struct {
	baseTimeout time.Duration
	userTimeout time.Duration
	RDB         *redis.Client
}

func NewArticleCache(rdb *redis.Client) *Cache {
	return &Cache{
		baseTimeout: 60 * time.Minute,
		userTimeout: 720 * time.Minute,
		RDB:         rdb,
	}
}

func (c *Cache) GetArticlesByPage(ctx context.Context, page, pageSize int) ([]ArticleWithoutContent, int, error) {
	// 获取文章列表
	key := ArticleByPageKey(page, pageSize)
	data, err := c.RDB.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, 0, ErrCacheMiss
	}
	if err != nil {
		return nil, 0, fmt.Errorf("缓存获取异常 %w", err)
	}

	var articles []ArticleWithoutContent
	if err = json.Unmarshal([]byte(data), &articles); err != nil {
		return nil, 0, fmt.Errorf("反序列化失败 %w", err)
	}

	// 获取总数
	totalData, err := c.RDB.Get(ctx, ArticleTotalKey()).Result()
	if err == redis.Nil {
		return articles, 0, ErrCacheMiss // 列表有但总数未命中
	}
	if err != nil {
		return nil, 0, fmt.Errorf("获取总数失败 %w", err)
	}

	total, err := strconv.Atoi(totalData)
	if err != nil {
		return nil, 0, fmt.Errorf("总数解析失败 %w", err)
	}

	return articles, total, nil
}

func (c *Cache) SetArticlesByPage(ctx context.Context, page, pageSize int, articles []ArticleWithoutContent, total int) error {
	// 序列化文章列表
	data, err := json.Marshal(articles)
	if err != nil {
		return fmt.Errorf("序列化失败 %w", err)
	}

	// 使用 Pipeline 批量设置
	pipe := c.RDB.Pipeline()
	pipe.Set(ctx, ArticleByPageKey(page, pageSize), data, c.baseTimeout)
	pipe.Set(ctx, ArticleTotalKey(), strconv.Itoa(total), c.baseTimeout)

	_, err = pipe.Exec(ctx)
	return err
}

func (c *Cache) GetArticleByID(ctx context.Context, id int) (*Article, error) {
	key := ArticleByIDKey(id)
	data, err := c.RDB.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, fmt.Errorf("缓存获取异常 %w", err)
	}

	var article Article
	if err := json.Unmarshal([]byte(data), &article); err != nil {
		return nil, fmt.Errorf("反序列化失败 %w", err)
	}

	return &article, nil
}

func (c *Cache) SetArticleByID(ctx context.Context, id int, article *Article) error {
	data, err := json.Marshal(article)
	if err != nil {
		return fmt.Errorf("序列化失败 %w", err)
	}

	err = c.RDB.Set(ctx, ArticleByIDKey(id), data, c.baseTimeout).Err()
	if err != nil {
		return err
	}

	return nil
}

func (c *Cache) GetArticlesByPopular(ctx context.Context, limit int) ([]ArticleWithoutContent, error) {
	key := ArticleByPopularKey(limit)

	data, err := c.RDB.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, fmt.Errorf("缓存获取异常 %w", err)
	}

	var articles []ArticleWithoutContent
	if err := json.Unmarshal([]byte(data), &articles); err != nil {
		return nil, fmt.Errorf("反序列化失败 %w", err)
	}

	return articles, nil
}

func (c *Cache) SetArticlesByPopular(ctx context.Context, limit int, articles []ArticleWithoutContent) error {
	data, err := json.Marshal(articles)
	if err != nil {
		return fmt.Errorf("序列化失败 %w", err)
	}

	err = c.RDB.Set(ctx, ArticleByPopularKey(limit), data, c.baseTimeout).Err()
	if err != nil {
		return err
	}

	return nil
}

func (c *Cache) AddViewUV(ctx context.Context, id int, userID string) error {
	return c.RDB.PFAdd(ctx, ArticleViewKey(id), userID).Err()
}

func (c *Cache) GetViewUV(ctx context.Context, id int) (int64, error) {
	return c.RDB.PFCount(ctx, ArticleViewKey(id)).Result()
}

func (c *Cache) DelViewUV(ctx context.Context, id int) error {
	return c.RDB.Del(ctx, ArticleViewKey(id)).Err()
}
