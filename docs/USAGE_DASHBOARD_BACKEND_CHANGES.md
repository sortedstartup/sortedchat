# Usage Dashboard Backend Implementation Plan

## Overview
This document outlines the backend changes required to implement a usage dashboard that tracks and displays AI token usage, costs, and chat statistics for users. The dashboard will provide insights similar to the design showing total tokens, costs, active chats, average cost per chat, token usage over time, usage by AI model, and recent chat sessions.

## Current State Analysis

### Existing Database Schema
The current schema already has the foundation for usage tracking:

**chat_messages table:**
- `input_token_count`, `output_token_count`, `cached_token_count`: Token usage per message
- `cost`: Calculated cost per message
- `model`: AI model used
- `created_at`: Timestamp for time-based analysis
- `user_id`: User association

**chat_list table:**
- `cost`, `input_token_count`, `output_token_count`, `cached_token_count`: Aggregated chat-level statistics

**model_metadata table:**
- `input_token_cost`, `output_token_cost`, `cached_token_cost`: Pricing information

### Existing Functionality
- Token counting and cost calculation already implemented
- Per-chat and per-message statistics stored
- User isolation with `user_id`

## Required Backend Changes

### 1. Proto File Changes (`proto/chatservice.proto`)

#### Add Usage Dashboard Service
```protobuf
service UsageDashboard {
    rpc GetUsageOverview(GetUsageOverviewRequest) returns (GetUsageOverviewResponse);
    rpc GetUsageTimeSeries(GetUsageTimeSeriesRequest) returns (GetUsageTimeSeriesResponse);
    rpc GetUsageByModel(GetUsageByModelRequest) returns (GetUsageByModelResponse);
    rpc GetRecentChatSessions(GetRecentChatSessionsRequest) returns (GetRecentChatSessionsResponse);
}
```

#### Add Message Definitions
```protobuf
// Usage Overview (top cards)
message GetUsageOverviewRequest {
    TimePeriod period = 1;
}

message GetUsageOverviewResponse {
    UsageOverview overview = 1;
}

message UsageOverview {
    int64 total_tokens = 1;
    double total_cost = 2;
    int32 active_chats = 3;
    double avg_cost_per_chat = 4;
    double tokens_change_percent = 5;
    double cost_change_percent = 6;
    int32 chats_today = 7;
}

// Time Series Data
message GetUsageTimeSeriesRequest {
    TimePeriod period = 1;
    TimeGranularity granularity = 2;
}

message GetUsageTimeSeriesResponse {
    repeated TimeSeriesPoint input_tokens = 1;
    repeated TimeSeriesPoint output_tokens = 2;
}

message TimeSeriesPoint {
    string date = 1; // ISO date string
    int64 value = 2;
}

// Usage by Model
message GetUsageByModelRequest {
    TimePeriod period = 1;
}

message GetUsageByModelResponse {
    repeated ModelUsage model_usage = 1;
}

message ModelUsage {
    string model_name = 1;
    int64 total_tokens = 2;
    double total_cost = 3;
    string model_color = 4; // For UI consistency
}

// Recent Chat Sessions
message GetRecentChatSessionsRequest {
    int32 limit = 1; // Default 10, max 50
}

message GetRecentChatSessionsResponse {
    repeated ChatSession sessions = 1;
}

message ChatSession {
    string chat_id = 1;
    string model = 2;
    int32 input_tokens = 3;
    int32 output_tokens = 4;
    double cost = 5;
    string date = 6; // ISO datetime string
}

// Enums
enum TimePeriod {
    LAST_WEEK = 0;
    LAST_MONTH = 1;
    LAST_3_MONTHS = 2;
    LAST_YEAR = 3;
}

enum TimeGranularity {
    DAILY = 0;
    WEEKLY = 1;
    MONTHLY = 2;
}
```

### 2. Database Changes

#### Migration File: `add_usage_analytics_indexes.up.sql`
```sql
-- Indexes for efficient usage analytics queries
CREATE INDEX IF NOT EXISTS idx_chat_messages_user_created_at ON chat_messages(user_id, created_at);
CREATE INDEX IF NOT EXISTS idx_chat_messages_user_model ON chat_messages(user_id, model);
CREATE INDEX IF NOT EXISTS idx_chat_messages_user_created_model ON chat_messages(user_id, created_at, model);
CREATE INDEX IF NOT EXISTS idx_chat_list_user_created_at ON chat_list(user_id, created_at);

-- Composite indexes for dashboard queries
CREATE INDEX IF NOT EXISTS idx_chat_messages_dashboard ON chat_messages(user_id, created_at, model, input_token_count, output_token_count, cost);
```

#### Optional: Usage Summary Table (for performance optimization)
```sql
-- Optional table for pre-computed daily summaries (implement later if needed)
CREATE TABLE IF NOT EXISTS usage_daily_summary (
    id BIGSERIAL PRIMARY KEY,
    user_id TEXT NOT NULL,
    date DATE NOT NULL,
    model TEXT NOT NULL,
    total_input_tokens BIGINT DEFAULT 0,
    total_output_tokens BIGINT DEFAULT 0,
    total_cached_tokens BIGINT DEFAULT 0,
    total_cost NUMERIC(12,6) DEFAULT 0,
    message_count INTEGER DEFAULT 0,
    chat_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, date, model)
);
```

### 3. DAO Layer Changes

#### Add to `dao/dao.go` interface:
```go
// Usage analytics methods
GetUsageOverview(userID string, period TimePeriod) (*UsageOverview, error)
GetUsageTimeSeries(userID string, period TimePeriod, granularity TimeGranularity) (*UsageTimeSeries, error)
GetUsageByModel(userID string, period TimePeriod) ([]*ModelUsage, error)
GetRecentChatSessions(userID string, limit int) ([]*ChatSession, error)
```

#### Add to `dao/models.go`:
```go
type UsageOverview struct {
    TotalTokens        int64   `db:"total_tokens"`
    TotalCost          float64 `db:"total_cost"`
    ActiveChats        int32   `db:"active_chats"`
    AvgCostPerChat     float64 `db:"avg_cost_per_chat"`
    TokensChangePercent float64
    CostChangePercent   float64
    ChatsToday         int32   `db:"chats_today"`
}

type UsageTimeSeries struct {
    InputTokens  []TimeSeriesPoint
    OutputTokens []TimeSeriesPoint
}

type TimeSeriesPoint struct {
    Date  string `db:"date"`
    Value int64  `db:"value"`
}

type ModelUsage struct {
    ModelName   string  `db:"model_name"`
    TotalTokens int64   `db:"total_tokens"`
    TotalCost   float64 `db:"total_cost"`
}

type ChatSession struct {
    ChatID       string  `db:"chat_id"`
    Model        string  `db:"model"`
    InputTokens  int32   `db:"input_tokens"`
    OutputTokens int32   `db:"output_tokens"`
    Cost         float64 `db:"cost"`
    Date         string  `db:"date"`
}

type TimePeriod int
const (
    LastWeek TimePeriod = iota
    LastMonth
    Last3Months
    LastYear
)

type TimeGranularity int
const (
    Daily TimeGranularity = iota
    Weekly
    Monthly
)
```

### 4. Service Layer Implementation

#### Create `service/usage_dashboard_service.go`:
```go
package service

type UsageDashboardService struct {
    dao dao.DAO
}

func NewUsageDashboardService(dao dao.DAO) *UsageDashboardService {
    return &UsageDashboardService{dao: dao}
}

func (s *UsageDashboardService) GetUsageOverview(ctx context.Context, userID string, period TimePeriod) (*pb.UsageOverview, error)
func (s *UsageDashboardService) GetUsageTimeSeries(ctx context.Context, userID string, period TimePeriod, granularity TimeGranularity) (*pb.GetUsageTimeSeriesResponse, error)
func (s *UsageDashboardService) GetUsageByModel(ctx context.Context, userID string, period TimePeriod) (*pb.GetUsageByModelResponse, error)
func (s *UsageDashboardService) GetRecentChatSessions(ctx context.Context, userID string, limit int) (*pb.GetRecentChatSessionsResponse, error)
```

### 5. API Layer Implementation

#### Create `api/usage_dashboard_api.go`:
```go
package api

type UsageDashboardAPI struct {
    service *service.UsageDashboardService
    pb.UnimplementedUsageDashboardServer
}

func NewUsageDashboardAPI(dao dao.DAO) *UsageDashboardAPI {
    return &UsageDashboardAPI{
        service: service.NewUsageDashboardService(dao),
    }
}

// Implement all proto service methods with auth middleware
```

### 6. Key SQL Queries (High-Level)

#### Usage Overview Query:
```sql
-- Get totals for current period
SELECT 
    SUM(input_token_count + output_token_count + cached_token_count) as total_tokens,
    SUM(cost) as total_cost,
    COUNT(DISTINCT chat_id) as active_chats,
    AVG(cost) as avg_cost_per_chat
FROM chat_messages 
WHERE user_id = ? AND created_at >= ?

-- Get comparison data for previous period
-- Calculate percentage changes
```

#### Time Series Query:
```sql
-- Daily granularity example
SELECT 
    DATE(created_at) as date,
    SUM(input_token_count) as input_tokens,
    SUM(output_token_count) as output_tokens
FROM chat_messages 
WHERE user_id = ? AND created_at >= ?
GROUP BY DATE(created_at)
ORDER BY date
```

#### Usage by Model Query:
```sql
SELECT 
    model as model_name,
    SUM(input_token_count + output_token_count + cached_token_count) as total_tokens,
    SUM(cost) as total_cost
FROM chat_messages 
WHERE user_id = ? AND created_at >= ?
GROUP BY model
ORDER BY total_tokens DESC
```

#### Recent Chat Sessions Query:
```sql
SELECT 
    chat_id,
    model,
    SUM(input_token_count) as input_tokens,
    SUM(output_token_count) as output_tokens,
    SUM(cost) as cost,
    MAX(created_at) as date
FROM chat_messages 
WHERE user_id = ?
GROUP BY chat_id, model
ORDER BY date DESC
LIMIT ?
```

## Implementation Strategy

### Phase 1: Core Infrastructure
1. Update proto file with new service and messages
2. Run `go generate` to generate Go and TypeScript code
3. Create database migration for indexes
4. Implement DAO interface methods

### Phase 2: Business Logic
1. Implement service layer with business logic
2. Add time period calculations and comparisons
3. Implement caching strategies if needed

### Phase 3: API Layer
1. Create API handlers with authentication
2. Add input validation and error handling
3. Register service in mono/main.go

### Phase 4: Testing & Optimization
1. Add unit tests for service methods
2. Test with sample data
3. Monitor query performance and optimize if needed

## Performance Considerations

1. **Indexes**: Added specific indexes for dashboard queries
2. **Time Range Filtering**: All queries filter by user_id and created_at first
3. **Aggregation**: Use database aggregation functions instead of application-level calculations
4. **Caching**: Consider caching frequently accessed data (daily summaries)
5. **Pagination**: Limit result sets, especially for recent chat sessions

## Security Considerations

1. **User Isolation**: All queries must filter by user_id from authenticated context
2. **Input Validation**: Validate time periods and limits
3. **Rate Limiting**: Consider rate limiting for dashboard endpoints
4. **Data Privacy**: Ensure no cross-user data leakage

## Migration Strategy

1. Database migrations can be applied without downtime
2. New service can be deployed alongside existing services
3. Frontend can progressively adopt new endpoints
4. No breaking changes to existing APIs

## Future Enhancements

1. **Real-time Updates**: WebSocket support for live dashboard updates
2. **Alerts**: Usage thresholds and cost alerts
3. **Exports**: CSV/PDF export functionality
4. **Detailed Analytics**: Per-project usage tracking
5. **Usage Predictions**: Forecasting based on historical data 