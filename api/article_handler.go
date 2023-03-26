package api

import (
	"net/http"
	"time"

	db "github.com/aliml92/realworld-gin-sqlc/db/sqlc"
	"github.com/gin-gonic/gin"
)

type articleResponse struct {
	Article struct {
		Slug           string     `json:"slug"`
		Title          string     `json:"title"`
		Description    string     `json:"description"`
		Body           string     `json:"body"`
		TagList        []string   `json:"tagList"`
		CreatedAt      time.Time  `json:"createdAt"`
		UpdatedAt      time.Time  `json:"updatedAt"`
		Favorited      bool       `json:"favorited"`
		FavoritesCount int64      `json:"favoritesCount"`
		Author         struct {
			Username  string 		`json:"username"`
			Bio       *string 		`json:"bio"`
			Image     *string 		`json:"image"`
			Following bool   		`json:"following"`	
		} `json:"author"`
	} `json:"article"`
}

func newArticleResponse(article *db.GetArticleBySlugRow) *articleResponse {
	resp := new(articleResponse)
	resp.Article.Slug = article.Slug
	resp.Article.Title = article.Title
	resp.Article.Description = article.Description
	resp.Article.Body = article.Body
	resp.Article.TagList = article.TagList
	resp.Article.CreatedAt = article.CreatedAt
	resp.Article.UpdatedAt = article.UpdatedAt
	resp.Article.Favorited = article.FavoritesCount > 0
	resp.Article.FavoritesCount = article.FavoritesCount
	resp.Article.Author.Username = article.Username.String
	if article.Bio.Valid {
		resp.Article.Author.Bio = &article.Bio.String
	}
	if article.Image.Valid {
		resp.Article.Author.Image = &article.Image.String
	}
	return resp
}


// GetArticle godoc
// @Summary Get an article
// @Description Get an article by slug
// @Tags articles
// @Accept  json
// @Produce  json
// @Param slug path string true "Article slug"
// @Success 200 {object} articleResponse
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /articles/{slug} [get]
func (s *Server) GetArticle(c *gin.Context) {
	slug := c.Param("slug")
	article, err := Nullable(s.store.GetArticleBySlug(c, slug))
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	if article == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}
	c.JSON(http.StatusOK, newArticleResponse(article))
}


type createArticleReq struct {
	Article struct {
		Title string `json:"title" binding:"required"`
		Description string `json:"description" binding:"required"`
		Body string `json:"body" binding:"required"`
		TagList []string `json:"tagList" binding:"omitempty"`
	} `json:"article" binding:"required"`
}

func (req *createArticleReq) bind(c *gin.Context, p *db.CreateArticleTxParams) error {
	if err := c.ShouldBindJSON(req); err != nil {
		return err
	}
	// p.Slug = slugify(req.Article.Title) // Implement
	// generate random slug and check if it does not exist in db
	// do it until it is unique
	p.CreateArticleParams.Slug = "how-to-train-your-dragon-1sd35f48"
	p.CreateArticleParams.Title = req.Article.Title
	p.CreateArticleParams.Description = req.Article.Description
	p.CreateArticleParams.Body = req.Article.Body
	p.Tags = req.Article.TagList
	return nil
}

// CreateArticle godoc
// @Summary Create article
// @Description Create Article
// @Tags articles
// @Accept json
// @Produce json
// @Success 201 {object} articleResponse
// @Failure 422 {object} Error
// @Failure 500 {object} Error
// @Security Bearer 
// @Router /articles [post]  
func (s *Server) CreateArticle(c *gin.Context) {
	id := "some_id"
	var (
		req createArticleReq
		p db.CreateArticleTxParams
	)
	if err := req.bind(c, &p); err != nil {
		c.JSON(http.StatusUnprocessableEntity, NewError(err))
	}
	p.AuthorID = id
	articleTx, err := s.store.CreateArticleTx(c, p)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewError(err))
		return
	}
	c.JSON(http.StatusCreated, newArticleTxResponse(articleTx))
}

func newArticleTxResponse(articleTx *db.CreateArticleTxResult) *articleResponse {
	resp := new(articleResponse)
	resp.Article.Slug = articleTx.Article.Slug
	resp.Article.Title = articleTx.Article.Title
	resp.Article.Description = articleTx.Article.Description
	resp.Article.Body = articleTx.Article.Body
	resp.Article.TagList = articleTx.Tags
	resp.Article.CreatedAt = articleTx.Article.CreatedAt
	resp.Article.UpdatedAt = articleTx.Article.UpdatedAt
	resp.Article.Favorited = false
	resp.Article.FavoritesCount = 0
	resp.Article.Author.Username = articleTx.User.Username 
	if articleTx.User.Bio.Valid {
		resp.Article.Author.Bio = &articleTx.User.Bio.String
	}
	if articleTx.User.Image.Valid {
		resp.Article.Author.Image = &articleTx.User.Image.String
	}
	return resp
}

