package article

import (
	"time"
)

type ArticleStatus uint8

const (
	ArticlePublic = iota
	ArticlePrivate
)

type Article struct {
	ID         int           `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
	Title      string        `json:"title"`               // 标题
	Desc       string        `json:"desc" gorm:"text"`    // 描述
	Content    string        `json:"content" gorm:"text"` // 正文
	AuthorName string        `json:"authorName"`          // 作者
	Views      uint          `json:"views"`               // 浏览数
	Tags       string        `json:"tags"`                // 标签（逗号分隔形式）
	Cover      string        `json:"cover"`               // 封面
	Status     ArticleStatus `json:"status"`              // 状态
	IsDelete   bool          `json:"is_delete"`
}

// ---------------------------------------

type ArticleWithoutContent struct {
	ID         int       `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Title      string    `json:"title"`
	AuthorName string    `json:"authorName"` // 作者
	Views      uint      `json:"views"`      // 浏览数
	Tags       string    `json:"tags"`       // 标签（逗号分隔）
	Cover      string    `json:"cover"`      // 封面
}
