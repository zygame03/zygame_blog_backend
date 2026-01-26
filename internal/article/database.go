package article

import (
	"gorm.io/gorm"
)

type Database struct {
	DB *gorm.DB
}

func NewDatabase(db *gorm.DB) *Database {
	return &Database{
		DB: db,
	}
}

func (db *Database) GetAllArticleIDs() ([]int, error) {
	ids := []int{}

	result := db.DB.
		Model(&Article{}).
		Select("id").
		Where("is_delete = false AND status = ?", ArticlePublic).
		Pluck("id", &ids)

	if result.Error != nil {
		return nil, nil
	}
	return ids, nil
}

// GetArticlesByPage
func (db *Database) GetArticlesByPage(page, pageSize int) ([]ArticleWithoutContent, int, error) {
	var articles []ArticleWithoutContent
	var total int64

	result := db.DB.Model(Article{}).
		Where("is_delete = false AND status = ?", ArticlePublic).
		Count(&total)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	result = db.DB.Model(Article{}).
		Where("is_delete = false AND status = ?", ArticlePublic).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&articles)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return articles, int(total), nil
}

// GetArticleByID
func (db *Database) GetArticleByID(id int) (*Article, error) {
	var article Article

	result := db.DB.First(&article, id)
	if result.Error != nil {
		return &article, result.Error
	}

	return &article, nil
}

func (db *Database) MGetArticleByID(ids []int) ([]*Article, error) {
	var articles []*Article

	err := db.DB.Where("id IN ?", ids).
		Find(&articles).
		Error
	if err != nil {
		return nil, err
	}

	return articles, nil
}

// GetArticlesByPopular
func (db *Database) GetArticlesByPopular(limit int) ([]ArticleWithoutContent, error) {
	var articles []ArticleWithoutContent

	result := db.DB.Model(&Article{}).
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

// IncrementViews 增加文章的 views
func (db *Database) IncrementViews(id int, increment int64) error {
	result := db.DB.Model(&Article{}).
		Where("id = ?", id).
		UpdateColumn("views", gorm.Expr("views + ?", increment))

	return result.Error
}

// BatchUpdateViews 批量更新文章的 views
func (db *Database) BatchUpdateViews(viewsMap map[int]int64) error {
	if len(viewsMap) == 0 {
		return nil
	}

	db.DB.Transaction(func(tx *gorm.DB) error {
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
