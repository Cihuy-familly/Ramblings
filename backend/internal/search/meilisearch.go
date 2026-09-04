package search

import (
	"log"

	"github.com/meilisearch/meilisearch-go"
	"gorm.io/gorm"

	"blog-platform-backend/internal/models"
)

// Service wraps the Meilisearch client and provides indexed search operations.
type Service struct {
	client   meilisearch.ServiceManager
	posts    meilisearch.IndexManager
	creators meilisearch.IndexManager
}

// NewService initializes a Meilisearch client, ensures indexes exist, and
// configures searchable attributes.
func NewService(host, apiKey string) *Service {
	client := meilisearch.New(host, meilisearch.WithAPIKey(apiKey))

	// Create posts index with primary key specified to avoid inference conflicts
	// (multiple fields ending with "id" in the document)
	client.CreateIndex(&meilisearch.IndexConfig{
		Uid:        "posts",
		PrimaryKey: "id",
	})

	postsIndex := client.Index("posts")
	postsIndex.UpdateSearchableAttributes(&[]string{
		"title",
		"content",
		"excerpt",
		"creator_name",
		"category_names",
	})
	postsIndex.UpdateFilterableAttributes(&[]interface{}{"published"})

	// Create creators index
	client.CreateIndex(&meilisearch.IndexConfig{
		Uid:        "creators",
		PrimaryKey: "id",
	})

	creatorsIndex := client.Index("creators")
	creatorsIndex.UpdateSearchableAttributes(&[]string{
		"display_name",
		"bio",
	})

	return &Service{
		client:   client,
		posts:    postsIndex,
		creators: creatorsIndex,
	}
}

// --- Post indexing ---

// PostDocument represents a post in the Meilisearch index.
type PostDocument struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Slug         string   `json:"slug"`
	Content      string   `json:"content"`
	Excerpt      string   `json:"excerpt"`
	ThumbnailURL string   `json:"thumbnail_url"`
	Published    bool     `json:"published"`
	CreatorID    string   `json:"creator_id"`
	CreatorName  string   `json:"creator_name"`
	CategoryNames []string `json:"category_names"`
	CreatedAt    int64    `json:"created_at"`
	UpdatedAt    int64    `json:"updated_at"`
}

// IndexPost adds or updates a post in the Meilisearch index.
func (s *Service) IndexPost(post models.Post, creatorName string, categoryNames []string) {
	doc := PostDocument{
		ID:           post.ID,
		Title:        post.Title,
		Slug:         post.Slug,
		Content:      post.Content,
		Excerpt:      post.Excerpt,
		ThumbnailURL: post.ThumbnailURL,
		Published:    post.Published,
		CreatorID:    post.CreatorID,
		CreatorName:  creatorName,
		CategoryNames: categoryNames,
		CreatedAt:    post.CreatedAt.UnixMilli(),
		UpdatedAt:    post.UpdatedAt.UnixMilli(),
	}
	if _, err := s.posts.AddDocuments(doc, nil); err != nil {
		log.Printf("Meilisearch: failed to index post %s: %v", post.ID, err)
	}
}

// RemovePost deletes a post from the Meilisearch index.
func (s *Service) RemovePost(postID string) {
	if _, err := s.posts.DeleteDocument(postID, nil); err != nil {
		log.Printf("Meilisearch: failed to remove post %s: %v", postID, err)
	}
}

// --- Creator indexing ---

// CreatorDocument represents a creator in the Meilisearch index.
type CreatorDocument struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Slug        string `json:"slug"`
	AvatarURL   string `json:"avatar_url"`
	Bio         string `json:"bio"`
	CreatedAt   int64  `json:"created_at"`
}

// IndexCreator adds or updates a creator in the Meilisearch index.
func (s *Service) IndexCreator(creator models.Creator) {
	doc := CreatorDocument{
		ID:          creator.ID,
		DisplayName: creator.DisplayName,
		Slug:        creator.Slug,
		AvatarURL:   creator.AvatarURL,
		Bio:         creator.Bio,
		CreatedAt:   creator.CreatedAt.UnixMilli(),
	}
	if _, err := s.creators.AddDocuments(doc, nil); err != nil {
		log.Printf("Meilisearch: failed to index creator %s: %v", creator.ID, err)
	}
}

// RemoveCreator deletes a creator from the Meilisearch index.
func (s *Service) RemoveCreator(creatorID string) {
	if _, err := s.creators.DeleteDocument(creatorID, nil); err != nil {
		log.Printf("Meilisearch: failed to remove creator %s: %v", creatorID, err)
	}
}

// --- Search ---

// PostResult is a search hit for a post.
type PostResult struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Slug         string   `json:"slug"`
	Excerpt      string   `json:"excerpt"`
	ThumbnailURL string   `json:"thumbnail_url"`
	CreatorID    string   `json:"creator_id"`
	CreatorName  string   `json:"creator_name"`
	CategoryNames []string `json:"category_names"`
	CreatedAt    int64    `json:"created_at"`
}

// CreatorResult is a search hit for a creator.
type CreatorResult struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Slug        string `json:"slug"`
	AvatarURL   string `json:"avatar_url"`
	Bio         string `json:"bio"`
}

// SearchPosts searches for published posts matching the query.
func (s *Service) SearchPosts(query string, page, limit int) ([]PostResult, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 12
	}

	hitsPerPage := int64(limit)
	searchPage := int64(page)
	searchRequest := &meilisearch.SearchRequest{
		Page:        searchPage,
		HitsPerPage: &hitsPerPage,
		Filter:      []string{"published = true"},
	}

	result, err := s.posts.Search(query, searchRequest)
	if err != nil {
		return nil, 0, err
	}

	var posts []PostResult
	if err := result.Hits.DecodeInto(&posts); err != nil {
		return nil, 0, err
	}

	return posts, result.TotalHits, nil
}

// SearchCreators searches for creators matching the query.
func (s *Service) SearchCreators(query string, page, limit int) ([]CreatorResult, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 12
	}

	hitsPerPage := int64(limit)
	searchPage := int64(page)
	searchRequest := &meilisearch.SearchRequest{
		Page:        searchPage,
		HitsPerPage: &hitsPerPage,
	}

	result, err := s.creators.Search(query, searchRequest)
	if err != nil {
		return nil, 0, err
	}

	var creators []CreatorResult
	if err := result.Hits.DecodeInto(&creators); err != nil {
		return nil, 0, err
	}

	return creators, result.TotalHits, nil
}

// SearchResult holds combined search results.
type SearchResult struct {
	Posts    []PostResult   `json:"posts"`
	Creators []CreatorResult `json:"creators"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
}

// SearchAll searches both posts and creators, returning posts first.
func (s *Service) SearchAll(query string, page, limit int) (*SearchResult, error) {
	posts, postsTotal, err := s.SearchPosts(query, page, limit)
	if err != nil {
		return nil, err
	}

	creators, _, err := s.SearchCreators(query, page, limit)
	if err != nil {
		return nil, err
	}

	total := postsTotal
	if total < 1 {
		total = int64(len(posts))
	}

	return &SearchResult{
		Posts:    posts,
		Creators: creators,
		Total:    total,
		Page:     page,
	}, nil
}

// SyncAllPosts indexes all posts from the database.
func (s *Service) SyncAllPosts(db *gorm.DB) {
	var posts []models.Post
	db.Preload("Creator").Preload("Categories").Find(&posts)
	for _, post := range posts {
		creatorName := ""
		categoryNames := []string{}
		if post.Creator.DisplayName != "" {
			creatorName = post.Creator.DisplayName
		}
		for _, cat := range post.Categories {
			categoryNames = append(categoryNames, cat.Name)
		}
		doc := PostDocument{
			ID:           post.ID,
			Title:        post.Title,
			Slug:         post.Slug,
			Content:      post.Content,
			Excerpt:      post.Excerpt,
			ThumbnailURL: post.ThumbnailURL,
			Published:    post.Published,
			CreatorID:    post.CreatorID,
			CreatorName:  creatorName,
			CategoryNames: categoryNames,
			CreatedAt:    post.CreatedAt.UnixMilli(),
			UpdatedAt:    post.UpdatedAt.UnixMilli(),
		}
		if _, err := s.posts.AddDocuments(doc, nil); err != nil {
			log.Printf("Meilisearch: failed to sync post %s: %v", post.ID, err)
		}
	}
	log.Printf("Meilisearch: synced %d posts", len(posts))
}

// SyncAllCreators indexes all creators from the database.
func (s *Service) SyncAllCreators(db *gorm.DB) {
	var creators []models.Creator
	db.Find(&creators)
	for _, c := range creators {
		doc := CreatorDocument{
			ID:          c.ID,
			DisplayName: c.DisplayName,
			Slug:        c.Slug,
			AvatarURL:   c.AvatarURL,
			Bio:         c.Bio,
			CreatedAt:   c.CreatedAt.UnixMilli(),
		}
		if _, err := s.creators.AddDocuments(doc, nil); err != nil {
			log.Printf("Meilisearch: failed to sync creator %s: %v", c.ID, err)
		}
	}
	log.Printf("Meilisearch: synced %d creators", len(creators))
}