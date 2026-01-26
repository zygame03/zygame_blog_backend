package article

import (
	"fmt"
)

func ArticleTotalKey() string {
	return "Article:Total"
}

func ArticleByIDKey(id int) string {
	return fmt.Sprintf("Article:ByID:%d", id)
}

func ArticleByPageKey(page, pageSize int) string {
	return fmt.Sprintf("Article:ByPage:%d:%d", page, pageSize)
}

func ArticleByPopularKey(limit int) string {
	return fmt.Sprintf("Article:ByPopular:%d", limit)
}

func ArticleActiveViewIDsKey() string {
	return "Article:View:ActiveIDs"
}

func ArticleViewKey(id int) string {
	if id == -1 {
		return "Article:view:UV:*"
	}
	return fmt.Sprintf("Article:View:UV:%d", id)
}
