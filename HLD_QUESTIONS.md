# High-Level Design (HLD) Questions with Answers

## 1. Design a URL Shortener (like bit.ly)

### Requirements
- Functional: Shorten long URLs, redirect short URLs to original
- Non-functional: High availability, low latency, scalable to billions of URLs

### HLD Components

**Architecture:**
```
Client → Load Balancer → API Servers → Cache (Redis) → Database (MySQL/PostgreSQL)
                                     ↓
                              Analytics Service
```

**Key Components:**
1. **API Gateway**: Rate limiting, authentication
2. **URL Generation Service**: Creates unique short codes (base62 encoding)
3. **Redirect Service**: Handles GET requests, cache-first lookup
4. **Database**: Stores URL mappings (short_code → original_url)
5. **Cache Layer**: Redis for hot URLs (80/20 rule)
6. **Analytics Service**: Track clicks, geography, referrers

**Database Schema:**
```sql
urls (
  id BIGINT PRIMARY KEY,
  short_code VARCHAR(10) UNIQUE,
  original_url TEXT,
  user_id BIGINT,
  created_at TIMESTAMP,
  expires_at TIMESTAMP,
  click_count INT
)
```

**Short Code Generation:**
- Use base62 encoding (a-z, A-Z, 0-9) = 62^7 = 3.5 trillion combinations
- Counter-based: Auto-increment ID → base62 encode
- Or MD5 hash first 7 chars (handle collisions)

**Scalability:**
- Horizontal scaling of API servers
- Database sharding by short_code hash
- CDN for static assets
- Read replicas for analytics

---

## 2. Design a Rate Limiter

### Requirements
- Limit requests per user/IP (e.g., 100 req/min)
- Distributed system support
- Low latency overhead

### HLD Components

**Algorithms:**

1. **Token Bucket**: Refill tokens at fixed rate, consume per request
2. **Leaky Bucket**: Process requests at constant rate, queue overflow drops
3. **Fixed Window**: Count requests in time windows (simple but has burst issue)
4. **Sliding Window Log**: Track timestamps of each request
5. **Sliding Window Counter**: Hybrid approach (recommended)

**Architecture:**
```
Client → API Gateway (Rate Limiter Middleware) → Redis Cluster → Backend Services
```

**Implementation (Sliding Window Counter):**
```
Key: user_id:timestamp_window
Value: request_count
TTL: window_duration

Algorithm:
1. current_window = floor(now / window_size)
2. previous_window = current_window - 1
3. current_count = redis.get(user:current_window)
4. previous_count = redis.get(user:previous_window)
5. weight = (now % window_size) / window_size
6. estimated_count = previous_count * (1 - weight) + current_count
7. if estimated_count < limit: allow else: reject
```

**Redis Commands:**
```redis
INCR user:123:1234567890
EXPIRE user:123:1234567890 120
```

**Distributed Considerations:**
- Use Redis cluster for high availability
- Lua scripts for atomic operations
- Eventual consistency acceptable (slight over-limit OK)

---

## 3. Design a Notification System

### Requirements
- Support multiple channels: Email, SMS, Push, In-app
- Handle millions of notifications/day
- Delivery guarantees, retry logic

### HLD Components

**Architecture:**
```
Event Source → Message Queue (Kafka) → Notification Service → Channel Handlers
                                                            ↓
                                                    [Email, SMS, Push, In-app]
                                                            ↓
                                                    Third-party APIs
                                                    (SendGrid, Twilio, FCM)
```

**Components:**
1. **Event Producer**: Triggers notifications (user actions, scheduled jobs)
2. **Message Queue**: Kafka/RabbitMQ for buffering and reliability
3. **Notification Service**: Orchestrates delivery, applies user preferences
4. **Channel Handlers**: Specialized workers for each channel
5. **Template Service**: Manages notification templates
6. **User Preference Service**: Stores opt-in/opt-out settings
7. **Delivery Tracker**: Logs delivery status, retries

**Database Schema:**
```sql
notifications (
  id BIGINT PRIMARY KEY,
  user_id BIGINT,
  type VARCHAR(50),
  channel VARCHAR(20),
  status VARCHAR(20),
  payload JSON,
  created_at TIMESTAMP,
  sent_at TIMESTAMP
)

user_preferences (
  user_id BIGINT,
  channel VARCHAR(20),
  notification_type VARCHAR(50),
  enabled BOOLEAN
)
```

**Flow:**
1. Event occurs → Publish to Kafka topic
2. Notification Service consumes → Check user preferences
3. Enrich with template → Route to channel handler
4. Channel handler → Call third-party API
5. Log result → Retry on failure (exponential backoff)

**Scalability:**
- Partition Kafka by user_id
- Separate worker pools per channel
- Rate limit third-party API calls
- Dead letter queue for failed notifications

---

## 4. Design a Distributed Cache

### Requirements
- Low latency (< 1ms)
- High throughput (millions of ops/sec)
- Consistency and availability trade-offs

### HLD Components

**Architecture:**
```
Client → Cache Client Library → Consistent Hashing → Cache Nodes (Redis/Memcached)
                                                   ↓
                                            Replication & Sharding
```

**Key Design Decisions:**

1. **Sharding Strategy:**
   - Consistent hashing with virtual nodes
   - Minimizes data movement on node add/remove
   - Hash function: MD5(key) % total_slots

2. **Replication:**
   - Master-slave replication per shard
   - Async replication for performance
   - Read from replicas, write to master

3. **Eviction Policies:**
   - LRU (Least Recently Used)
   - LFU (Least Frequently Used)
   - TTL-based expiration

4. **Cache Patterns:**
   - **Cache-aside**: App checks cache, loads from DB on miss
   - **Write-through**: Write to cache and DB synchronously
   - **Write-back**: Write to cache, async flush to DB

**Handling Cache Failures:**
- Circuit breaker pattern
- Fallback to database
- Cache warming on node recovery

**Monitoring:**
- Hit/miss ratio
- Latency percentiles (p50, p95, p99)
- Memory usage per node
- Eviction rate

