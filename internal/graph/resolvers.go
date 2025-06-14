package graph

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/StillLearnSVN/go-graphql-articles/internal/database"
	"github.com/StillLearnSVN/go-graphql-articles/internal/models"
	"github.com/graphql-go/graphql"
)

type Resolver struct {
	db *database.DB
}

func NewResolver(db *database.DB) *Resolver {
	return &Resolver{db: db}
}

func (r *Resolver) CreateArticle(p graphql.ResolveParams) (interface{}, error) {
	input := p.Args["input"].(map[string]interface{})

	title := input["title"].(string)
	body := input["body"].(string)
	authorName := input["authorName"].(string)

	// Validate input
	if strings.TrimSpace(title) == "" {
		return nil, fmt.Errorf("title cannot be empty")
	}
	if strings.TrimSpace(body) == "" {
		return nil, fmt.Errorf("body cannot be empty")
	}
	if strings.TrimSpace(authorName) == "" {
		return nil, fmt.Errorf("author name cannot be empty")
	}

	// Begin transaction for data consistency
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert or get author
	var authorID int
	err = tx.QueryRow(`
        INSERT INTO authors (name) VALUES ($1) 
        ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name 
        RETURNING id`, authorName).Scan(&authorID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert/get author: %w", err)
	}

	// Insert article
	var article models.Article
	err = tx.QueryRow(`
        INSERT INTO articles (title, body, author_id) 
        VALUES ($1, $2, $3) 
        RETURNING id, title, body, author_id, created_at`,
		title, body, authorID).Scan(
		&article.ID, &article.Title, &article.Body,
		&article.AuthorID, &article.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert article: %w", err)
	}

	// Get author details
	var author models.Author
	err = tx.QueryRow("SELECT id, name FROM authors WHERE id = $1", authorID).Scan(&author.ID, &author.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get author: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	article.Author = &author

	return map[string]interface{}{
		"id":        strconv.Itoa(article.ID),
		"title":     article.Title,
		"body":      article.Body,
		"author":    author,
		"createdAt": article.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (r *Resolver) GetArticles(p graphql.ResolveParams) (interface{}, error) {
	// Parse pagination parameters
	first := 10 // default
	if f, ok := p.Args["first"].(int); ok && f > 0 {
		first = f
		if first > 100 { // Limit to prevent abuse
			first = 100
		}
	}

	var after *string
	if a, ok := p.Args["after"].(string); ok && a != "" {
		after = &a
	}

	queryText := ""
	if q, ok := p.Args["query"].(string); ok {
		queryText = strings.TrimSpace(q)
	}

	authorFilter := ""
	if a, ok := p.Args["author"].(string); ok {
		authorFilter = strings.TrimSpace(a)
	}

	// Build the SQL query with proper indexing
	var baseQuery strings.Builder
	var countQuery strings.Builder
	var args []interface{}
	argIndex := 1

	baseQuery.WriteString(`
        SELECT a.id, a.title, a.body, a.author_id, a.created_at, 
               au.id, au.name
        FROM articles a
        JOIN authors au ON a.author_id = au.id
    `)

	countQuery.WriteString(`
        SELECT COUNT(*)
        FROM articles a
        JOIN authors au ON a.author_id = au.id
    `)

	// Build WHERE clause
	var whereConditions []string

	// Text search using PostgreSQL full-text search
	if queryText != "" {
		whereConditions = append(whereConditions, fmt.Sprintf(`
            (to_tsvector('english', a.title) @@ plainto_tsquery('english', $%d) 
             OR to_tsvector('english', a.body) @@ plainto_tsquery('english', $%d))`, argIndex, argIndex))
		args = append(args, queryText)
		argIndex++
	}

	// Author filter
	if authorFilter != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("au.name ILIKE $%d", argIndex))
		args = append(args, "%"+authorFilter+"%")
		argIndex++
	}

	// Cursor-based pagination
	if after != nil {
		cursorID, cursorTime, err := models.DecodeCursor(*after)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %w", err)
		}
		whereConditions = append(whereConditions, fmt.Sprintf(`
            (a.created_at < $%d OR (a.created_at = $%d AND a.id < $%d))`, argIndex, argIndex, argIndex+1))
		args = append(args, cursorTime, cursorID)
		argIndex += 2
	}

	// Apply WHERE clause if needed
	if len(whereConditions) > 0 {
		whereClause := " WHERE " + strings.Join(whereConditions, " AND ")
		baseQuery.WriteString(whereClause)
		countQuery.WriteString(whereClause)
	}

	// Add ORDER BY and LIMIT
	baseQuery.WriteString(" ORDER BY a.created_at DESC, a.id DESC")
	baseQuery.WriteString(fmt.Sprintf(" LIMIT $%d", argIndex))
	args = append(args, first+1) // Get one extra to check if there's a next page

	// Execute count query
	var totalCount int
	if err := r.db.QueryRow(countQuery.String(), args[:len(args)-1]...).Scan(&totalCount); err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Execute main query
	rows, err := r.db.Query(baseQuery.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query articles: %w", err)
	}
	defer rows.Close()

	var articles []*models.Article
	for rows.Next() {
		var article models.Article
		var author models.Author

		err := rows.Scan(
			&article.ID, &article.Title, &article.Body,
			&article.AuthorID, &article.CreatedAt,
			&author.ID, &author.Name,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan article: %w", err)
		}

		article.Author = &author
		articles = append(articles, &article)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// Determine pagination info
	hasNextPage := len(articles) > first
	if hasNextPage {
		articles = articles[:first] // Remove the extra item
	}

	hasPreviousPage := after != nil

	// Create edges
	edges := make([]map[string]interface{}, len(articles))
	for i, article := range articles {
		cursor := models.EncodeCursor(article.ID, article.CreatedAt)
		edges[i] = map[string]interface{}{
			"node": map[string]interface{}{
				"id":    strconv.Itoa(article.ID),
				"title": article.Title,
				"body":  article.Body,
				"author": map[string]interface{}{
					"id":   strconv.Itoa(article.Author.ID),
					"name": article.Author.Name,
				},
				"createdAt": article.CreatedAt.Format(time.RFC3339),
			},
			"cursor": cursor,
		}
	}

	// Create page info
	var startCursor, endCursor *string
	if len(edges) > 0 {
		start := edges[0]["cursor"].(string)
		end := edges[len(edges)-1]["cursor"].(string)
		startCursor = &start
		endCursor = &end
	}

	pageInfo := map[string]interface{}{
		"hasNextPage":     hasNextPage,
		"hasPreviousPage": hasPreviousPage,
		"startCursor":     startCursor,
		"endCursor":       endCursor,
	}

	return map[string]interface{}{
		"edges":      edges,
		"pageInfo":   pageInfo,
		"totalCount": totalCount,
	}, nil
}
