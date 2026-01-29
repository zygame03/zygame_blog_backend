package article

import (
	"gorm.io/gorm"
)

func repoGetAllArticleIDs(db *gorm.DB) ([]int, error) {
	ids := []int{}

	result := db.
		Model(&Article{}).
		Select("id").
		Where("is_delete = false AND status = ?", ArticlePublic).
		Pluck("id", &ids)

	if result.Error != nil {
		return nil, nil
	}
	return ids, nil
}

// repoGetArticlesByPage
func repoGetArticlesByPage(db *gorm.DB, page, pageSize int) ([]ArticleWithoutContent, int, error) {
	var articles []ArticleWithoutContent
	var total int64

	result := db.
		Model(Article{}).
		Where("is_delete = false AND status = ?", ArticlePublic).
		Count(&total)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	result = db.Model(Article{}).
		Where("is_delete = false AND status = ?", ArticlePublic).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&articles)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return articles, int(total), nil
}

// repoGetArticleByID
func repoGetArticleByID(db *gorm.DB, id int) (*Article, error) {
	var article Article

	result := db.First(&article, id)
	if result.Error != nil {
		return &article, result.Error
	}

	return &article, nil
}

func repoMGetArticleByID(db *gorm.DB, ids []int) ([]*Article, error) {
	var articles []*Article

	err := db.
		Where("id IN ?", ids).
		Find(&articles).
		Error
	if err != nil {
		return nil, err
	}

	return articles, nil
}

// repoGetArticlesByPopular
func repoGetArticlesByPopular(db *gorm.DB, limit int) ([]ArticleWithoutContent, error) {
	var articles []ArticleWithoutContent

	result := db.
		Model(&Article{}).
		Where("is_delete = false AND status = ?", ArticlePublic).
		Order("views DESC").
		Select("id, created_at, updated_at, title, author_name, views, tags, cover").
		Limit(limit).
		Find(&articles)
	if result.Error != nil {
		return nil, result.Error
	}

	return articles, nil
}

// repoIncrementViews 增加文章的 views
func repoIncrementViews(db *gorm.DB, id int, increment int64) error {
	result := db.
		Model(&Article{}).
		Where("id = ?", id).
		UpdateColumn("views", gorm.Expr("views + ?", increment))

	return result.Error
}

// repoBatchUpdateViews 批量更新文章的 views
func repoBatchUpdateViews(db *gorm.DB, viewsMap map[int]int64) error {
	if len(viewsMap) == 0 {
		return nil
	}

	db.Transaction(func(tx *gorm.DB) error {
		for id, increment := range viewsMap {
			if err := tx.Model(&Article{}).
				Where("id = ?", id).
				UpdateColumn("views", gorm.Expr("views + ?", increment)).Error; err != nil {
				return err
			}
		}
		return nil
	})

	return nil
}
