# Go GraphQL Articles 
 Skill test - Internship Backend Engineer

 ![image](https://miro.medium.com/v2/resize:fit:1100/format:webp/1*9aXf_XFjX4PTX7PYi0U8ng.png)

A sample Go project demonstrating a GraphQL API for managing articles and authors, using PostgreSQL as the database. The project is containerized with Docker and includes integration and unit tests.

## Features

- GraphQL API for articles and authors
- PostgreSQL database integration
- Database migrations
- Docker and Docker Compose support
- Unit and integration tests

## Project Structure

```
├── cmd/
│   └── server/           # Main application entrypoint
│       └── main.go
├── internal/
│   ├── database/         # Database connection and migrations
│   │   ├── connection.go
│   │   └── migrations.go
│   ├── graph/            # GraphQL resolvers and schema
│   │   ├── resolvers.go
│   │   └── schema.go
│   └── models/           # Data models
│       ├── article.go
│       └── author.go
├── postgresql.conf/      # Custom PostgreSQL configuration (optional)
├── tests/                # Unit and integration tests
│   ├── integration_test.go
│   └── unit_test.go
├── Dockerfile            # Dockerfile for the Go application
├── docker-compose.yml    # Docker Compose for multi-container setup
├── go.mod                # Go module definition
├── go.sum                # Go module checksums
```

## Getting Started

### Prerequisites
- [Go](https://golang.org/dl/) 1.24+
- [Docker](https://www.docker.com/get-started)
- [Docker Compose](https://docs.docker.com/compose/)

### Running with Docker Compose

1. Build and start the services:
   ```bash
   docker-compose up --build
   ```
2. The GraphQL server will be available at `http://localhost:8080/query`.
3. Access the PostgreSQL database at `localhost:5432` (default user/password: `postgres`/`postgres`).

### Running Locally (without Docker)

1. Start a local PostgreSQL instance (see `docker-compose.yml` for configuration).
2. Set the required environment variables (see below).
3. Run the application:
   ```bash
   go run ./cmd/server/main.go
   ```

## Environment Variables

- `DB_HOST` (default: `localhost`)
- `DB_PORT` (default: `5432`)
- `DB_USER` (default: `postgres`)
- `DB_PASSWORD` (default: `postgres`)
- `DB_NAME` (default: `articles_db`)

## Database Migrations

Migrations are handled in `internal/database/migrations.go`. On startup, the application will automatically apply pending migrations.

## Testing

Run all tests:
```bash
go test ./...
```

Or run specific tests:
```bash
go test ./tests/unit_test.go
```

## GraphQL Playground

You can use [GraphQL Playground](https://github.com/graphql/graphql-playground) or [Altair](https://altair.sirmuel.design/) to interact with the API at `http://localhost:8080/query`.

## Example GraphQL Query
- Get 10 Articles
```
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
        createdAt
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
```
- Search Articles by Keyword
```
query SearchArticlesByKeyword {
  articles(first: 10, query: "GraphQL") {
    totalCount
    edges {
      node {
        id
        title
        body
      }
    }
  }
}
```

- Filter ArticlesByAuthor
```
query FilterArticlesByAuthor {
  articles(first: 5, author: "KumparanTECH") {
    totalCount
    edges {
      node {
        id
        title
        author {
          name
        }
      }
    }
  }
}
```

## Example GraphQL Mutation
- Create an Article
```
mutation {
  createArticle(input: {
    title: "Studi Ungkap Ikan Alami Rasa Sakit Luar Biasa Sebelum Mati usai Ditangkap"
    body: "Dalam sebuah studi yang terbit Scientific Reports mengungkap rasa sakit tersembunyi yang dialami ikan sebelum mati setelah dia ditangkap untuk kemudian di jual di pasar."
    authorName: "kumparanSAINS"
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
```

- Create Multiple Article
```
mutation BuatTigaArtikelTechBaru {
  artikelAppleAI: createArticle(input: {
    title: "Apple Intelligence Resmi Diumumkan, Bawa Fitur AI Canggih ke iPhone"
    body: "Apple akhirnya memasuki era kecerdasan buatan dengan mengumumkan 'Apple Intelligence', serangkaian fitur AI yang akan terintegrasi secara mendalam di iOS 18, iPadOS 18, and macOS Sequoia. Fitur ini berfokus pada privasi pengguna."
    authorName: "KumparanTECH"
  }) {
    id
    title
    author {
      name
    }
  }
  
  artikelGoToTikTok: createArticle(input: {
    title: "GoTo Catat Kenaikan Kunjungan di Tokopedia Pasca Integrasi dengan TikTok"
    body: "PT GoTo Gojek Tokopedia Tbk (GoTo) melaporkan adanya peningkatan signifikan dalam jumlah kunjungan dan transaksi di platform Tokopedia setelah proses integrasi dengan TikTok Shop rampung. Sinergi ini disebut menguntungkan UMKM lokal."
    authorName: "KumparanTECH"
  }) {
    id
    title
    author {
      name
    }
  }
  
  artikelXAI: createArticle(input: {
    title: "Elon Musk Umumkan Superkomputer xAI untuk Saingi OpenAI"
    body: "Elon Musk mengumumkan rencana pembangunan superkomputer yang disebut 'Gigafactory of Compute' untuk mendukung pengembangan model kecerdasan buatan (AI) generasi berikutnya dari startup xAI miliknya."
    authorName: "KumparanTECH"
  }) {
    id
    title
    author {
      name
    }
  }
}
```

## Author
Samuel Volder [@StillLearnSVN](https://github.com/StillLearnSVN)
