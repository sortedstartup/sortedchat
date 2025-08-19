# sortedchat

Sorted Chat is a UI to chat with multiple LLM models without being locked into one.
First we will have all popular models working

Then we will start implementing new features which are generally not present in chat ui,
like RAG, creating and hosting web apps.

# Run Command

```
CGO_CFLAGS="-I$(pwd)/sqlite3" go run -tags "sqlite_fts5" ./mono/
```

# Wails Run Command(GO)
```
CGO_CFLAGS="-I$(pwd)/../sqlite3" go run -tags "sqlite_fts5",dev,webkit2_41 main.go wails.go
```


# Run Postgres
```
docker run -d \
  --name sortedchat_postgres_dev \
  -e POSTGRES_DB=sortedchat_dev \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=dev_password \
  -p 5432:5432 \
  --restart unless-stopped \
  pgvector/pgvector:pg15
```

# Export Environment Variables
export DB_TYPE=postgres
export POSTGRES_HOST=localhost  
export POSTGRES_PASSWORD=dev_password
export POSTGRES_DATABASE=sortedchat_dev
export POSTGRES_PORT=5432
export POSTGRES_USERNAME=postgres