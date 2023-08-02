package typesense

import (
	"context"
	"strings"
	"time"

	ts "github.com/typesense/typesense-go/typesense"
	"github.com/typesense/typesense-go/typesense/api"
	"github.com/typesense/typesense-go/typesense/api/pointer"

	"github.com/aliml92/realworld-gin-sqlc/config"
	"github.com/aliml92/realworld-gin-sqlc/search"
)

type TypesenseHandler struct {
	Client     *ts.Client
	Collection string
}

func NewTypesenseHandler(client *ts.Client, collectionName string) *TypesenseHandler {
	return &TypesenseHandler{
		Client:     client,
		Collection: collectionName,
	}
}

func NewClient(config *config.Config) *ts.Client {
	client := ts.NewClient(
		ts.WithServer(config.TypesenseAddr),
		ts.WithAPIKey(config.TypesenseAPIKEY),
		ts.WithConnectionTimeout(5*time.Second),
		ts.WithCircuitBreakerMaxRequests(50),
		ts.WithCircuitBreakerInterval(2*time.Minute),
		ts.WithCircuitBreakerTimeout(1*time.Minute),
	)
	return client
}

func (th TypesenseHandler) CreateCollection() error {
	schema := &api.CollectionSchema{
		Name: th.Collection,
		Fields: []api.Field{
			{
				Name:  "author_id",
				Type:  "string",
				Facet: pointer.True(),
			},
			{
				Name:  "slug",
				Type:  "string",
				Facet: pointer.True(),
			},
			{
				Name:  "title",
				Type:  "string",
				Facet: pointer.False(),
			},
			{
				Name:  "description",
				Type:  "string",
				Facet: pointer.False(),
			},
			{
				Name:  "body",
				Type:  "string",
				Facet: pointer.False(),
			},
			{
				Name:  "created_at",
				Type:  "int64",
				Facet: pointer.True(),
			},
			{
				Name:  "updated_at",
				Type:  "int64",
				Facet: pointer.True(),
			},
		},
		DefaultSortingField: pointer.String("created_at"),
	}

	_, err := th.Client.Collections().Create(schema)
	if err != nil {
		if strings.Contains(err.Error(), "exists") {
			return nil
		}
		return err
	}
	return nil
}

func (th *TypesenseHandler) CreateArticle(_ context.Context, article search.Article) error {
	a := struct {
		ID          string `json:"id,omitempty"`
		AuthorID    string `json:"author_id"`
		Slug        string `json:"slug"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Body        string `json:"body"`
		CreatedAt   int64  `json:"created_at"`
		UpdatedAt   int64  `json:"updated_at"`
	}{
		ID:          article.ID,
		AuthorID:    article.AuthorID,
		Slug:        article.Slug,
		Title:       article.Title,
		Description: article.Description,
		Body:        article.Body,
		CreatedAt:   article.CreatedAt.Unix(),
		UpdatedAt:   article.CreatedAt.Unix(),
	}
	_, err := th.Client.Collection(th.Collection).Documents().Create(a)
	return err
}

func (th *TypesenseHandler) UpdateArticle(_ context.Context, article search.ArticleUpdate) error {
	a := struct {
		Slug        string `json:"slug,omitempty"`
		Title       string `json:"title,omitmepty"`
		Description string `json:"description,omitempty"`
		Body        string `json:"body,omitempty"`
		UpdatedAt   int64  `json:"updated_at,omitempty"`
	}{
		Slug:        *article.Updates.Slug,
		Title:       *article.Updates.Title,
		Description: *article.Updates.Description,
		Body:        *article.Updates.Body,
		UpdatedAt:   article.Updates.UpdatedAt.Unix(),
	}
	_, err := th.Client.Collection(th.Collection).Document(article.ID).Update(a)
	if strings.Contains(err.Error(), "201") {
		return nil
	}
	return err
}

func (th *TypesenseHandler) DeleteArticle(_ context.Context, id string) error {
	_, err := th.Client.Collection(th.Collection).Document(id).Delete()
	if err != nil {
		return err
	}
	return nil
}

func (th *TypesenseHandler) Search(_ context.Context, params search.SearchParams) (*search.ArticlesWithCount, error) {
	searchParams := api.SearchCollectionParams{
		Q:        params.Q,
		QueryBy:  "title, description, body",
		NumTypos: pointer.String("1,2,2"),
		Page:     &params.Page,
		PerPage:  &params.PerPage,
	}
	result, err := th.Client.Collection(th.Collection).Documents().Search(&searchParams)
	if err != nil {
		return nil, err
	}
	total := result.Found
	if *total == 0 {
		return nil, nil
	}
	var articles []search.Article
	for _, hit := range *result.Hits {
		document := *hit.Document
		article := search.Article{
			ID:          document["id"].(string),
			AuthorID:    document["author_id"].(string),
			Slug:        document["slug"].(string),
			Title:       document["title"].(string),
			Description: document["description"].(string),
			Body:        document["body"].(string),
		}
		createdAt := document["created_at"].(float64)
		updatedAt := document["updated_at"].(float64)
		article.CreatedAt = time.Unix(int64(createdAt), 0)
		article.UpdatedAt = time.Unix(int64(updatedAt), 0)
		articles = append(articles, article)
	}
	return &search.ArticlesWithCount{
		Total:    *total,
		Articles: articles,
	}, nil
}
