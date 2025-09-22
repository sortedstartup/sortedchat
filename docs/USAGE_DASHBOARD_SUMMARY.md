# Usage Dashboard - Backend Implementation Summary

## Quick Overview
The current backend already has **90% of the required infrastructure** for the usage dashboard. Token counting, cost calculation, and data storage are already implemented. We only need to add new API endpoints to aggregate and present this data.

## Key Components Needed

### 1. New Proto Service (in `chatservice.proto`)
```protobuf
service UsageDashboard {
    rpc GetUsageOverview(GetUsageOverviewRequest) returns (GetUsageOverviewResponse);
    rpc GetUsageTimeSeries(GetUsageTimeSeriesRequest) returns (GetUsageTimeSeriesResponse);
    rpc GetUsageByModel(GetUsageByModelRequest) returns (GetUsageByModelResponse);
    rpc GetRecentChatSessions(GetRecentChatSessionsRequest) returns (GetRecentChatSessionsResponse);
}
```

### 2. Four Main API Endpoints Required

| Endpoint | Purpose | Data Source |
|----------|---------|-------------|
| `GetUsageOverview` | Top dashboard cards (total tokens, cost, active chats) | `chat_messages` table aggregation |
| `GetUsageTimeSeries` | Line chart (token usage over time) | `chat_messages` grouped by date |
| `GetUsageByModel` | Model usage breakdown | `chat_messages` grouped by model |
| `GetRecentChatSessions` | Recent chat sessions table | `chat_messages` grouped by chat_id |

### 3. Database Optimizations
```sql
-- Add these indexes for better query performance
CREATE INDEX idx_chat_messages_user_created_at ON chat_messages(user_id, created_at);
CREATE INDEX idx_chat_messages_user_model ON chat_messages(user_id, model);
CREATE INDEX idx_chat_messages_dashboard ON chat_messages(user_id, created_at, model, input_token_count, output_token_count, cost);
```

### 4. File Structure for Implementation

```
backend/chatservice/
├── proto/           # Auto-generated from .proto file
├── api/
│   └── usage_dashboard_api.go    # New API handlers
├── service/
│   └── usage_dashboard_service.go # New business logic
├── dao/
│   ├── dao.go       # Add new interface methods
│   ├── models.go    # Add new data structures
│   ├── dao_sqlite.go   # Implement SQLite queries
│   └── dao_postgres.go # Implement PostgreSQL queries
└── dao/db/
    └── migrations/
        └── 14_usage_analytics_indexes.up.sql # New migration
```

## Implementation Steps

### Phase 1: Foundation (Day 1)
1. Update `proto/chatservice.proto` with new service definitions
2. Run `go generate` to generate Go/TypeScript code
3. Create database migration for performance indexes

### Phase 2: Backend Logic (Day 2-3)
1. Add interface methods to `dao/dao.go`
2. Implement SQL queries in `dao_sqlite.go` and `dao_postgres.go`
3. Create service layer with business logic
4. Add API handlers with authentication

### Phase 3: Integration (Day 4)
1. Register new service in `mono/main.go`
2. Test with sample data
3. Verify performance and security

## Key Advantages

✅ **Minimal Database Changes**: Only adding indexes, no schema changes
✅ **No Breaking Changes**: New service doesn't affect existing APIs  
✅ **Reuses Existing Data**: All required data is already being collected
✅ **User Isolation**: Built-in security with existing `user_id` filtering
✅ **Scalable**: Designed for efficient querying with proper indexes

## Sample Data Flow

```
User Request → API Handler → Service Layer → DAO Layer → Database Query → Response
     ↓              ↓             ↓            ↓             ↓              ↓
Auth Check → Validate Input → Business Logic → SQL Query → Raw Data → Proto Response
```

## Expected Query Performance

With proper indexes, all dashboard queries should execute in **< 100ms** even with:
- 100,000+ chat messages per user
- 1-year time period analysis
- Multiple model comparisons

The design prioritizes database-level aggregation over application-level processing for optimal performance. 