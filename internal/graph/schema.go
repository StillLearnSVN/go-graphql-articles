package graph

import (
	"github.com/graphql-go/graphql"
)

func CreateSchema(resolver *Resolver) (graphql.Schema, error) {
	// Author type
	authorType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Author",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.NewNonNull(graphql.ID),
			},
			"name": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
	})

	// Article type
	articleType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Article",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.NewNonNull(graphql.ID),
			},
			"title": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
			"body": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
			"author": &graphql.Field{
				Type: graphql.NewNonNull(authorType),
			},
			"createdAt": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
	})

	// PageInfo type
	pageInfoType := graphql.NewObject(graphql.ObjectConfig{
		Name: "PageInfo",
		Fields: graphql.Fields{
			"hasNextPage": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Boolean),
			},
			"hasPreviousPage": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Boolean),
			},
			"startCursor": &graphql.Field{
				Type: graphql.String,
			},
			"endCursor": &graphql.Field{
				Type: graphql.String,
			},
		},
	})

	// ArticleEdge type
	articleEdgeType := graphql.NewObject(graphql.ObjectConfig{
		Name: "ArticleEdge",
		Fields: graphql.Fields{
			"node": &graphql.Field{
				Type: graphql.NewNonNull(articleType),
			},
			"cursor": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
	})

	// ArticleConnection type
	articleConnectionType := graphql.NewObject(graphql.ObjectConfig{
		Name: "ArticleConnection",
		Fields: graphql.Fields{
			"edges": &graphql.Field{
				Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(articleEdgeType))),
			},
			"pageInfo": &graphql.Field{
				Type: graphql.NewNonNull(pageInfoType),
			},
			"totalCount": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Int),
			},
		},
	})

	// ArticleInput type
	articleInputType := graphql.NewInputObject(graphql.InputObjectConfig{
		Name: "ArticleInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"title": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
			"body": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
			"authorName": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
	})

	// Query type
	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"articles": &graphql.Field{
				Type: graphql.NewNonNull(articleConnectionType),
				Args: graphql.FieldConfigArgument{
					"first": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
					"after": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"query": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"author": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: resolver.GetArticles,
			},
		},
	})

	// Mutation type
	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"createArticle": &graphql.Field{
				Type: graphql.NewNonNull(articleType),
				Args: graphql.FieldConfigArgument{
					"input": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(articleInputType),
					},
				},
				Resolve: resolver.CreateArticle,
			},
		},
	})

	return graphql.NewSchema(graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
	})
}
