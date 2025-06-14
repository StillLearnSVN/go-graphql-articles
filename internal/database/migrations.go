package database

import (
    "fmt"
)

func (db *DB) RunMigrations() error {
    // Create authors table
    createAuthorsTable := `
    CREATE TABLE IF NOT EXISTS authors (
        id SERIAL PRIMARY KEY,
        name VARCHAR(255) NOT NULL UNIQUE,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );`
    
    // Create articles table with proper indexing
    createArticlesTable := `
    CREATE TABLE IF NOT EXISTS articles (
        id SERIAL PRIMARY KEY,
        title VARCHAR(500) NOT NULL,
        body TEXT NOT NULL,
        author_id INTEGER NOT NULL REFERENCES authors(id),
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );`
    
    // Create indexes for efficient querying
    createIndexes := []string{
        "CREATE INDEX IF NOT EXISTS idx_articles_created_at ON articles(created_at DESC);",
        "CREATE INDEX IF NOT EXISTS idx_articles_author_id ON articles(author_id);",
        "CREATE INDEX IF NOT EXISTS idx_articles_title_gin ON articles USING gin(to_tsvector('english', title));",
        "CREATE INDEX IF NOT EXISTS idx_articles_body_gin ON articles USING gin(to_tsvector('english', body));",
        "CREATE INDEX IF NOT EXISTS idx_authors_name ON authors(name);",
    }
    
    // Execute migrations
    if _, err := db.Exec(createAuthorsTable); err != nil {
        return fmt.Errorf("failed to create authors table: %w", err)
    }
    
    if _, err := db.Exec(createArticlesTable); err != nil {
        return fmt.Errorf("failed to create articles table: %w", err)
    }
    
    for _, indexSQL := range createIndexes {
        if _, err := db.Exec(indexSQL); err != nil {
            return fmt.Errorf("failed to create index: %w", err)
        }
    }
    
    fmt.Println("Database migrations completed successfully")
    return nil
}