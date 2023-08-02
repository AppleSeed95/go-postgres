package search

import (
	"context"
	"time"
)

type Searcher interface {
	CreateArticle(ctx context.Context, article Article) error
	UpdateArticle(ctx context.Context, article ArticleUpdate) error
	DeleteArticle(ctx context.Context, id string) error
	Search(ctx context.Context, params SearchParams) (*ArticlesWithCount, error)
}

type ArticlesWithCount struct {
	Articles []Article `json:"articles"`
	Total    int       `json:"totarticlesCount"`
}

type Article struct {
	ID          string    `json:"id,omitempty"`
	AuthorID    string    `json:"author_id"`
	Slug        string    `json:"slug"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Body        string    `json:"body"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}

type ArticleUpdate struct {
	ID      string `json:"id"`
	Updates struct {
		Slug        *string   `json:"slug,omitempty"`
		Title       *string   `json:"title,omitempty"`
		Description *string   `json:"description,omitempty"`
		Body        *string   `json:"body,omitempty"`
		UpdatedAt   time.Time `json:"updated_at"`
	}
}

type SearchParams struct {
	Q       string
	Page    int
	PerPage int
}
