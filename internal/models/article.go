package models

import (
    "time"
    "encoding/base64"
    "fmt"
    "strconv"
    "strings"
)

type Article struct {
    ID        int       `json:"id"`
    Title     string    `json:"title"`
    Body      string    `json:"body"`
    AuthorID  int       `json:"author_id"`
    Author    *Author   `json:"author,omitempty"`
    CreatedAt time.Time `json:"created_at"`
}

type ArticleInput struct {
    Title      string `json:"title"`
    Body       string `json:"body"`
    AuthorName string `json:"authorName"`
}

type PageInfo struct {
    HasNextPage     bool    `json:"hasNextPage"`
    HasPreviousPage bool    `json:"hasPreviousPage"`
    StartCursor     *string `json:"startCursor"`
    EndCursor       *string `json:"endCursor"`
}

type ArticleEdge struct {
    Node   *Article `json:"node"`
    Cursor string   `json:"cursor"`
}

type ArticleConnection struct {
    Edges      []*ArticleEdge `json:"edges"`
    PageInfo   *PageInfo      `json:"pageInfo"`
    TotalCount int            `json:"totalCount"`
}

// EncodeCursor creates a base64 encoded cursor from article ID and timestamp
func EncodeCursor(id int, createdAt time.Time) string {
    cursor := fmt.Sprintf("%d:%d", id, createdAt.Unix())
    return base64.StdEncoding.EncodeToString([]byte(cursor))
}

// DecodeCursor decodes a base64 cursor to get article ID and timestamp
func DecodeCursor(cursor string) (int, time.Time, error) {
    decoded, err := base64.StdEncoding.DecodeString(cursor)
    if err != nil {
        return 0, time.Time{}, err
    }
    
    parts := strings.Split(string(decoded), ":")
    if len(parts) != 2 {
        return 0, time.Time{}, fmt.Errorf("invalid cursor format")
    }
    
    id, err := strconv.Atoi(parts[0])
    if err != nil {
        return 0, time.Time{}, err
    }
    
    timestamp, err := strconv.ParseInt(parts[1], 10, 64)
    if err != nil {
        return 0, time.Time{}, err
    }
    
    return id, time.Unix(timestamp, 0), nil
}