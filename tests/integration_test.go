package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/StillLearnSVN/go-graphql-articles/internal/database"
	"github.com/StillLearnSVN/go-graphql-articles/internal/graph"
	"github.com/graphql-go/handler"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type IntegrationTestSuite struct {
	suite.Suite
	db      *database.DB
	handler http.Handler
}

func (suite *IntegrationTestSuite) SetupSuite() {
	err := godotenv.Load("../.env_test")
	suite.Require().NoError(err, "Error loading .env_test file")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("SSL_MODE"),
	)

	initialDB, err := sql.Open("postgres", connStr)
	suite.Require().NoError(err)
	defer initialDB.Close()

	_, _ = initialDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", os.Getenv("DB_NAME")))

	_, err = initialDB.Exec(fmt.Sprintf("CREATE DATABASE %s", os.Getenv("DB_NAME")))
	suite.Require().NoError(err)


	suite.db, err = database.NewConnection()
	suite.Require().NoError(err, "Failed to connect to the test database. Ensure database.NewConnection() reads env vars.")

	err = suite.db.RunMigrations()
	suite.Require().NoError(err)

	// Setup GraphQL handler
	resolver := graph.NewResolver(suite.db)
	schema, err := graph.CreateSchema(resolver)
	suite.Require().NoError(err)

	suite.handler = handler.New(&handler.Config{
		Schema: &schema,
		Pretty: true,
	})
}

func (suite *IntegrationTestSuite) TearDownSuite() {
	suite.db.Close()
}

func (suite *IntegrationTestSuite) SetupTest() {
	// Clean up data before each test
	_, err := suite.db.Exec("DELETE FROM articles")
	suite.Require().NoError(err)
	_, err = suite.db.Exec("DELETE FROM authors")
	suite.Require().NoError(err)
}

func (suite *IntegrationTestSuite) TestCreateArticle() {
	mutation := `
        mutation {
            createArticle(input: {
                title: "Test Article"
                body: "This is a test article body"
                authorName: "John Doe"
            }) {
                id
                title
                body
                author {
                    id
                    name
                }
                createdAt
            }
        }
    `

	response := suite.executeGraphQL(mutation)

	assert.NotNil(suite.T(), response["data"])
	data := response["data"].(map[string]interface{})
	article := data["createArticle"].(map[string]interface{})

	assert.NotEmpty(suite.T(), article["id"])
	assert.Equal(suite.T(), "Test Article", article["title"])
	assert.Equal(suite.T(), "This is a test article body", article["body"])

	author := article["author"].(map[string]interface{})
	assert.Equal(suite.T(), "John Doe", author["name"])
	assert.NotEmpty(suite.T(), article["createdAt"])
}

func (suite *IntegrationTestSuite) TestCreateArticle_ValidationError() {
	mutation := `
        mutation {
            createArticle(input: {
                title: ""
                body: "This is a test article body"
                authorName: "John Doe"
            }) {
                id
                title
            }
        }
    `

	response := suite.executeGraphQL(mutation)

	assert.NotNil(suite.T(), response["errors"])
	errors := response["errors"].([]interface{})
	assert.Greater(suite.T(), len(errors), 0)
}

func (suite *IntegrationTestSuite) TestGetArticles() {
	// First, create some test articles
	suite.createTestArticle("First Article", "Content 1", "Alice")
	suite.createTestArticle("Second Article", "Content 2", "Bob")
	suite.createTestArticle("Third Article", "Content 3", "Alice")

	query := `
        query {
            articles(first: 10) {
                edges {
                    node {
                        id
                        title
                        body
                        author {
                            name
                        }
                    }
                    cursor
                }
                pageInfo {
                    hasNextPage
                    hasPreviousPage
                    startCursor
                    endCursor
                }
                totalCount
            }
        }
    `

	response := suite.executeGraphQL(query)

	assert.NotNil(suite.T(), response["data"])
	data := response["data"].(map[string]interface{})
	articles := data["articles"].(map[string]interface{})

	edges := articles["edges"].([]interface{})
	assert.Equal(suite.T(), 3, len(edges))
	assert.Equal(suite.T(), float64(3), articles["totalCount"])

	// Check that articles are sorted by created_at DESC
	firstEdge := edges[0].(map[string]interface{})
	firstNode := firstEdge["node"].(map[string]interface{})
	assert.Equal(suite.T(), "Third Article", firstNode["title"])
}

func (suite *IntegrationTestSuite) TestGetArticles_WithAuthorFilter() {
	// Create test articles
	suite.createTestArticle("Alice Article 1", "Content 1", "Alice")
	suite.createTestArticle("Bob Article", "Content 2", "Bob")
	suite.createTestArticle("Alice Article 2", "Content 3", "Alice")

	query := `
        query {
            articles(first: 10, author: "Alice") {
                edges {
                    node {
                        title
                        author {
                            name
                        }
                    }
                }
                totalCount
            }
        }
    `

	response := suite.executeGraphQL(query)

	data := response["data"].(map[string]interface{})
	articles := data["articles"].(map[string]interface{})

	edges := articles["edges"].([]interface{})
	assert.Equal(suite.T(), 2, len(edges))
	assert.Equal(suite.T(), float64(2), articles["totalCount"])

	// Verify all articles are by Alice
	for _, edge := range edges {
		node := edge.(map[string]interface{})["node"].(map[string]interface{})
		author := node["author"].(map[string]interface{})
		assert.Equal(suite.T(), "Alice", author["name"])
	}
}

func (suite *IntegrationTestSuite) TestGetArticles_WithTextSearch() {
	// Create test articles
	suite.createTestArticle("GraphQL Tutorial", "Learn GraphQL basics", "Alice")
	suite.createTestArticle("REST API Guide", "Building REST APIs", "Bob")
	suite.createTestArticle("Database Design", "GraphQL with databases", "Alice")

	query := `
        query {
            articles(first: 10, query: "GraphQL") {
                edges {
                    node {
                        title
                        body
                    }
                }
                totalCount
            }
        }
    `

	response := suite.executeGraphQL(query)

	data := response["data"].(map[string]interface{})
	articles := data["articles"].(map[string]interface{})

	edges := articles["edges"].([]interface{})
	assert.Equal(suite.T(), 2, len(edges))
	assert.Equal(suite.T(), float64(2), articles["totalCount"])
}

func (suite *IntegrationTestSuite) TestGetArticles_Pagination() {
	// Create test articles
	for i := 1; i <= 5; i++ {
		suite.createTestArticle(fmt.Sprintf("Article %d", i), fmt.Sprintf("Content %d", i), "Author")
	}

	// Get first 2 articles
	query := `
        query {
            articles(first: 2) {
                edges {
                    node {
                        title
                    }
                    cursor
                }
                pageInfo {
                    hasNextPage
                    endCursor
                }
            }
        }
    `

	response := suite.executeGraphQL(query)
	data := response["data"].(map[string]interface{})
	articles := data["articles"].(map[string]interface{})

	edges := articles["edges"].([]interface{})
	assert.Equal(suite.T(), 2, len(edges))

	pageInfo := articles["pageInfo"].(map[string]interface{})
	assert.True(suite.T(), pageInfo["hasNextPage"].(bool))

	endCursor := pageInfo["endCursor"].(string)
	assert.NotEmpty(suite.T(), endCursor)

	// Get next page
	nextQuery := fmt.Sprintf(`
        query {
            articles(first: 2, after: "%s") {
                edges {
                    node {
                        title
                    }
                }
                pageInfo {
                    hasNextPage
                }
            }
        }
    `, endCursor)

	nextResponse := suite.executeGraphQL(nextQuery)
	nextData := nextResponse["data"].(map[string]interface{})
	nextArticles := nextData["articles"].(map[string]interface{})

	nextEdges := nextArticles["edges"].([]interface{})
	assert.Equal(suite.T(), 2, len(nextEdges))
}

func (suite *IntegrationTestSuite) createTestArticle(title, body, authorName string) {
	mutation := fmt.Sprintf(`
        mutation {
            createArticle(input: {
                title: "%s"
                body: "%s"
                authorName: "%s"
            }) {
                id
            }
        }
    `, title, body, authorName)

	suite.executeGraphQL(mutation)
}

func (suite *IntegrationTestSuite) executeGraphQL(query string) map[string]interface{} {
	requestBody := map[string]string{
		"query": query,
	}

	jsonBody, err := json.Marshal(requestBody)
	suite.Require().NoError(err)

	req := httptest.NewRequest("POST", "/graphql", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	suite.handler.ServeHTTP(recorder, req)

	var response map[string]interface{}
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	suite.Require().NoError(err)

	return response
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
