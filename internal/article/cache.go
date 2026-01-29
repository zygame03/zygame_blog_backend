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

const (
	articleCacheExpiration = 60 * time.Minute
)

func cacheGetArticlesByPage(ctx context.Context, rdb *redis.Client, page, pageSize int) ([]ArticleWithoutContent, int, error) {
	// 获取文章列表
	key := ArticleByPageKey(page, pageSize)
	data, err := rdb.Get(ctx, key).Result()
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
	totalData, err := rdb.Get(ctx, ArticleTotalKey()).Result()
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

func cacheSetArticlesByPage(ctx context.Context, rdb *redis.Client, page, pageSize int, articles []ArticleWithoutContent, total int) error {
	// 序列化文章列表
	data, err := json.Marshal(articles)
	if err != nil {
		return fmt.Errorf("序列化失败 %w", err)
	}

	// 使用 Pipeline 批量设置
	pipe := rdb.Pipeline()
	pipe.Set(ctx, ArticleByPageKey(page, pageSize), data, articleCacheExpiration)
	pipe.Set(ctx, ArticleTotalKey(), strconv.Itoa(total), articleCacheExpiration)

	_, err = pipe.Exec(ctx)
	return err
}

func cacheGetArticleByID(ctx context.Context, rdb *redis.Client, id int) (*Article, error) {
	key := ArticleByIDKey(id)
	data, err := rdb.Get(ctx, key).Result()
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

func cacheSetArticleByID(ctx context.Context, rdb *redis.Client, id int, article *Article) error {
	data, err := json.Marshal(article)
	if err != nil {
		return fmt.Errorf("序列化失败 %w", err)
	}

	err = rdb.Set(ctx, ArticleByIDKey(id), data, articleCacheExpiration).Err()
	if err != nil {
		return err
	}

	return nil
}

func cacheGetArticlesByPopular(ctx context.Context, rdb *redis.Client, limit int) ([]ArticleWithoutContent, error) {
	key := ArticleByPopularKey(limit)

	data, err := rdb.Get(ctx, key).Result()
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

func cacheSetArticlesByPopular(ctx context.Context, rdb *redis.Client, limit int, articles []ArticleWithoutContent) error {
	data, err := json.Marshal(articles)
	if err != nil {
		return fmt.Errorf("序列化失败 %w", err)
	}

	err = rdb.Set(ctx, ArticleByPopularKey(limit), data, articleCacheExpiration).Err()
	if err != nil {
		return err
	}

	return nil
}

func cacheAddViewUV(ctx context.Context, rdb *redis.Client, id int, userID string) error {
	return rdb.PFAdd(ctx, ArticleViewKey(id), userID).Err()
}

func cacheGetViewUV(ctx context.Context, rdb *redis.Client, id int) (int64, error) {
	return rdb.PFCount(ctx, ArticleViewKey(id)).Result()
}

func cacheDelViewUV(ctx context.Context, rdb *redis.Client, id int) error {
	return rdb.Del(ctx, ArticleViewKey(id)).Err()
}
