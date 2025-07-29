# sortedchat

Sorted Chat is a UI to chat with multiple LLM models without being locked into one.
First we will have all popular models working

Then we will start implementing new features which are generally not present in chat ui,
like RAG, creating and hosting web apps.

# Run Command

```
CGO_CFLAGS="-I$(pwd)/sqlite3" go run -tags "sqlite_fts5" ./mono/
```