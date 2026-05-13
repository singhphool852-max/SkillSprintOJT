# System Design Questions with Detailed Answers

## 1. Design Instagram

### Functional Requirements
- Upload photos/videos
- Follow users, view feed
- Like, comment, share
- Search users and hashtags
- Stories (24-hour expiry)

### Non-Functional Requirements
- 500M daily active users
- 100M photos uploaded per day
- Low latency for feed (<200ms)
- High availability (99.99%)
- Eventual consistency acceptable

### Capacity Estimation
- **Storage**: 100M photos/day × 2MB avg = 200TB/day = 73PB/year
- **Bandwidth**: 200TB/day ÷ 86400s = 2.3GB/s upload
- **QPS**: 500M users × 50 requests/day ÷ 86400s = 290K QPS

### High-Level Architecture

```
┌─────────────┐
│   Client    │
│ (iOS/Web)   │
└──────┬──────┘
       │
┌──────▼──────────────────────────────────────┐
│         CDN (CloudFront)                     │
│    (Images, Videos, Static Assets)           │
└──────────────────────────────────────────────┘
       │
┌──────▼──────────────────────────────────────┐
│      Load Balancer (ELB)                     │
└──────┬──────────────────────────────────────┘
       │
┌──────▼──────────────────────────────────────┐
│      API Gateway                             │
│  (Auth, Rate Limiting, Routing)              │
└──────┬──────────────────────────────────────┘
       │
       ├─────────────────┬─────────────────┬──────────────┐
       │                 │                 │              │
┌──────▼──────┐  ┌──────▼──────┐  ┌──────▼──────┐  ┌───▼────┐
│   User      │  │   Feed      │  │   Media     │  │ Search │
│  Service    │  │  Service    │  │  Service    │  │Service │
└──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └───┬────┘
       │                 │                 │              │
┌──────▼──────┐  ┌──────▼──────┐  ┌──────▼──────┐  ┌───▼────┐
│  User DB    │  │  Feed DB    │  │   Object    │  │Elastic │
│ (MySQL)     │  │ (Cassandra) │  │  Storage    │  │ Search │
└─────────────┘  └─────────────┘  │    (S3)     │  └────────┘
                                   └─────────────┘
       │
┌──────▼──────────────────────────────────────┐
│      Redis Cache                             │
│  (User profiles, Feed cache, Session)        │
└──────────────────────────────────────────────┘
       │
┌──────▼──────────────────────────────────────┐
│      Message Queue (Kafka)                   │
│  (Async tasks: notifications, analytics)     │
└──────────────────────────────────────────────┘
```

### Database Schema

**Users Table (MySQL)**
```sql
users (
  id BIGINT PRIMARY KEY,
  username VARCHAR(50) UNIQUE,
  email VARCHAR(100) UNIQUE,
  password_hash VARCHAR(255),
  profile_pic_url TEXT,
  bio TEXT,
  follower_count INT DEFAULT 0,
  following_count INT DEFAULT 0,
  created_at TIMESTAMP
)
```

**Posts Table (MySQL)**
```sql
posts (
  id BIGINT PRIMARY KEY,
  user_id BIGINT,
  caption TEXT,
  media_urls JSON,
  location VARCHAR(100),
  like_count INT DEFAULT 0,
  comment_count INT DEFAULT 0,
  created_at TIMESTAMP,
  INDEX(user_id, created_at)
)
```

**Followers Table (MySQL)**
```sql
followers (
  follower_id BIGINT,
  followee_id BIGINT,
  created_at TIMESTAMP,
  PRIMARY KEY(follower_id, followee_id),
  INDEX(followee_id)
)
```

**Feed Table (Cassandra)** - Denormalized for fast reads
```
feed_timeline (
  user_id BIGINT,
  post_id BIGINT,
  post_timestamp TIMESTAMP,
  PRIMARY KEY(user_id, post_timestamp)
) WITH CLUSTERING ORDER BY (post_timestamp DESC)
```

### Key Components

#### 1. Media Upload Service
```
Flow:
1. Client requests signed URL from API
2. API generates S3 pre-signed URL (valid 15 min)
3. Client uploads directly to S3
4. S3 triggers Lambda for image processing
5. Lambda creates thumbnails (150x150, 640x640)
6. Store metadata in Posts table
7. Fanout post to followers' feeds (async)
```

#### 2. Feed Generation Service

**Push Model (Fanout-on-write)**:
- When user posts, write to all followers' feeds
- Pros: Fast read (pre-computed)
- Cons: Slow write for celebrities (millions of followers)

**Pull Model (Fanout-on-read)**:
- Fetch posts from followed users on demand
- Pros: Fast write
- Cons: Slow read (join multiple tables)

**Hybrid Approach** (Instagram's actual approach):
- Push for normal users (<10K followers)
- Pull for celebrities (>10K followers)
- Cache celebrity posts in Redis

```python
def get_feed(user_id, page=1, limit=20):
    # Get from cache first
    cache_key = f"feed:{user_id}:{page}"
    cached = redis.get(cache_key)
    if cached:
        return cached
    
    # Get following list
    following = get_following(user_id)
    
    # Separate celebrities and normal users
    celebrities = [u for u in following if u.follower_count > 10000]
    normal_users = [u for u in following if u.follower_count <= 10000]
    
    # Pull from pre-computed feeds (push model)
    feed_posts = cassandra.query(
        "SELECT * FROM feed_timeline WHERE user_id = ? LIMIT ?",
        user_id, limit
    )
    
    # Pull from celebrities (pull model)
    celeb_posts = mysql.query(
        "SELECT * FROM posts WHERE user_id IN ? ORDER BY created_at DESC LIMIT ?",
        celebrities, limit
    )
    
    # Merge and sort
    merged = merge_sort(feed_posts, celeb_posts, key='created_at')[:limit]
    
    # Cache for 5 minutes
    redis.setex(cache_key, 300, merged)
    
    return merged
```

#### 3. Search Service
- **Elasticsearch** for full-text search
- Index: users (username, name), hashtags, locations
- Real-time indexing via Kafka consumer

```json
{
  "mappings": {
    "properties": {
      "username": { "type": "text", "analyzer": "standard" },
      "name": { "type": "text" },
      "bio": { "type": "text" },
      "follower_count": { "type": "integer" }
    }
  }
}
```

#### 4. Notification Service
- Kafka topics: `likes`, `comments`, `follows`
- Workers consume and send push notifications
- Firebase Cloud Messaging (FCM) for mobile
- WebSocket for real-time web notifications

### Scalability Considerations

1. **Database Sharding**:
   - Shard users by `user_id % num_shards`
   - Shard posts by `post_id % num_shards`
   - Consistent hashing for cache

2. **CDN Strategy**:
   - Serve images from edge locations
   - Cache popular posts (viral content)
   - Lazy loading for feed

3. **Caching Layers**:
   - L1: Application cache (in-memory)
   - L2: Redis cluster (distributed)
   - L3: CDN (edge cache)

4. **Async Processing**:
   - Feed fanout (Kafka)
   - Image processing (Lambda)
   - Analytics (Spark streaming)

### Monitoring & Observability
- Metrics: Latency (p50, p95, p99), Error rate, QPS
- Logs: Centralized logging (ELK stack)
- Tracing: Distributed tracing (Jaeger)
- Alerts: PagerDuty for critical issues

---

## 2. Design WhatsApp

### Functional Requirements
- One-on-one messaging
- Group chats (up to 256 members)
- Media sharing (images, videos, documents)
- End-to-end encryption
- Read receipts, typing indicators
- Last seen, online status

### Non-Functional Requirements
- 2 billion users
- 100 billion messages/day
- Real-time delivery (<100ms)
- 99.999% availability
- Strong consistency for message ordering

### Capacity Estimation
- **Messages**: 100B/day ÷ 86400s = 1.16M messages/sec
- **Storage**: 100B × 100 bytes = 10TB/day = 3.65PB/year
- **Connections**: 2B users × 10% online = 200M concurrent WebSocket connections

### High-Level Architecture

```
┌─────────────┐
│   Client    │
│ (Mobile App)│
└──────┬──────┘
       │ WebSocket
┌──────▼──────────────────────────────────────┐
│      Connection Manager                      │
│  (Maintains WebSocket connections)           │
│  (Sticky sessions via user_id hash)          │
└──────┬──────────────────────────────────────┘
       │
┌──────▼──────────────────────────────────────┐
│      Message Gateway                         │
│  (Routes messages, handles presence)         │
└──────┬──────────────────────────────────────┘
       │
       ├─────────────────┬─────────────────┬──────────────┐
       │                 │                 │              │
┌──────▼──────┐  ┌──────▼──────┐  ┌──────▼──────┐  ┌───▼────┐
│   Chat      │  │   User      │  │   Media     │  │ Group  │
│  Service    │  │  Service    │  │  Service    │  │Service │
└──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └───┬────┘
       │                 │                 │              │
┌──────▼──────┐  ┌──────▼──────┐  ┌──────▼──────┐  ┌───▼────┐
│  Messages   │  │  Users DB   │  │   Object    │  │ Groups │
│     DB      │  │ (Cassandra) │  │  Storage    │  │   DB   │
│(Cassandra)  │  └─────────────┘  │    (S3)     │  │(MySQL) │
└─────────────┘                    └─────────────┘  └────────┘
       │
┌──────▼──────────────────────────────────────┐
│      Redis (Presence, Session, Cache)        │
└──────────────────────────────────────────────┘
       │
┌──────▼──────────────────────────────────────┐
│      Kafka (Message Queue)                   │
│  (Offline message delivery, analytics)       │
└──────────────────────────────────────────────┘
```



### Database Schema

**Messages Table (Cassandra)**
```
messages (
  message_id UUID,
  conversation_id UUID,
  sender_id BIGINT,
  receiver_id BIGINT,
  content TEXT,
  media_url TEXT,
  timestamp TIMESTAMP,
  is_delivered BOOLEAN,
  is_read BOOLEAN,
  PRIMARY KEY((conversation_id), timestamp, message_id)
) WITH CLUSTERING ORDER BY (timestamp DESC)
```

**Users Table (Cassandra)**
```
users (
  user_id BIGINT PRIMARY KEY,
  phone_number VARCHAR(20) UNIQUE,
  username VARCHAR(50),
  profile_pic_url TEXT,
  public_key TEXT,  -- For E2E encryption
  last_seen TIMESTAMP,
  status VARCHAR(50)
)
```

**Groups Table (MySQL)**
```sql
groups (
  group_id BIGINT PRIMARY KEY,
  name VARCHAR(100),
  created_by BIGINT,
  created_at TIMESTAMP
)

group_members (
  group_id BIGINT,
  user_id BIGINT,
  role VARCHAR(20),  -- admin, member
  joined_at TIMESTAMP,
  PRIMARY KEY(group_id, user_id)
)
```

### Key Components

#### 1. WebSocket Connection Manager
```python
class ConnectionManager:
    def __init__(self):
        self.active_connections = {}  # user_id -> WebSocket
        self.user_to_server = {}      # user_id -> server_id (Redis)
    
    async def connect(self, user_id, websocket):
        self.active_connections[user_id] = websocket
        
        # Register in Redis for routing
        redis.hset("user_connections", user_id, SERVER_ID)
        
        # Update presence
        redis.setex(f"presence:{user_id}", 300, "online")
        
        # Deliver pending messages
        await self.deliver_pending_messages(user_id)
    
    async def disconnect(self, user_id):
        del self.active_connections[user_id]
        redis.hdel("user_connections", user_id)
        redis.setex(f"presence:{user_id}", 300, "offline")
    
    async def send_message(self, user_id, message):
        if user_id in self.active_connections:
            await self.active_connections[user_id].send_json(message)
        else:
            # User offline - queue message
            kafka.produce("offline_messages", {
                "user_id": user_id,
                "message": message
            })
```

#### 2. Message Delivery Flow

**Sender → Receiver (Both Online)**:
```
1. Sender sends message via WebSocket
2. Gateway validates and assigns message_id
3. Store in Cassandra (async)
4. Lookup receiver's server from Redis
5. Route to receiver's WebSocket connection
6. Receiver sends ACK (delivered)
7. Update message status in DB
8. Forward ACK to sender
9. Receiver reads message → send read receipt
```

**Sender → Receiver (Receiver Offline)**:
```
1. Sender sends message
2. Store in Cassandra
3. Push to Kafka topic "offline_messages"
4. When receiver comes online:
   - Fetch undelivered messages from Cassandra
   - Deliver via WebSocket
   - Mark as delivered
```

#### 3. End-to-End Encryption (Signal Protocol)
```
Key Exchange:
1. Each user generates identity key pair (public/private)
2. Public keys stored on server
3. Sender fetches receiver's public key
4. Generate session key using Diffie-Hellman
5. Encrypt message with session key
6. Server only sees encrypted blob

Message Encryption:
plaintext → AES-256(plaintext, session_key) → ciphertext
Server stores: {
  "message_id": "...",
  "encrypted_content": "...",  // Server can't decrypt
  "sender_id": "...",
  "receiver_id": "..."
}
```

#### 4. Group Messaging
```python
async def send_group_message(group_id, sender_id, message):
    # Get group members
    members = mysql.query(
        "SELECT user_id FROM group_members WHERE group_id = ?",
        group_id
    )
    
    # Store message once
    message_id = uuid.uuid4()
    cassandra.execute(
        "INSERT INTO group_messages (message_id, group_id, sender_id, content, timestamp) VALUES (?, ?, ?, ?, ?)",
        message_id, group_id, sender_id, message, datetime.now()
    )
    
    # Fanout to all members (async)
    for member in members:
        if member.user_id == sender_id:
            continue
        
        # Check if online
        server_id = redis.hget("user_connections", member.user_id)
        if server_id:
            # Send via WebSocket
            await route_to_server(server_id, member.user_id, {
                "type": "group_message",
                "group_id": group_id,
                "message_id": message_id,
                "sender_id": sender_id,
                "content": message
            })
        else:
            # Queue for offline delivery
            kafka.produce("offline_messages", {
                "user_id": member.user_id,
                "message_id": message_id
            })
```

#### 5. Presence & Typing Indicators
```python
# Presence (Last Seen)
def update_presence(user_id):
    redis.setex(f"presence:{user_id}", 300, "online")
    redis.set(f"last_seen:{user_id}", datetime.now().isoformat())

def get_presence(user_id):
    status = redis.get(f"presence:{user_id}")
    if status == "online":
        return "online"
    else:
        last_seen = redis.get(f"last_seen:{user_id}")
        return f"last seen {format_time(last_seen)}"

# Typing Indicator (Ephemeral)
def send_typing_indicator(sender_id, receiver_id):
    # Don't store, just route
    server_id = redis.hget("user_connections", receiver_id)
    if server_id:
        route_to_server(server_id, receiver_id, {
            "type": "typing",
            "sender_id": sender_id
        })
```

### Scalability & Reliability

#### 1. Horizontal Scaling
- **Connection Servers**: Shard by user_id hash
- **Message Servers**: Stateless, scale independently
- **Database**: Cassandra auto-sharding by partition key

#### 2. Message Ordering
- Use Cassandra clustering key (timestamp, message_id)
- Client-side sequence numbers for conflict resolution
- Vector clocks for distributed ordering

#### 3. Fault Tolerance
- **Message Durability**: Write to Cassandra before ACK
- **Connection Failover**: Reconnect with exponential backoff
- **Data Replication**: Cassandra RF=3 (3 replicas)

#### 4. Load Balancing
- Consistent hashing for user → server mapping
- Sticky sessions (user always connects to same server)
- Health checks and auto-scaling

### Monitoring
- **Metrics**: Message delivery rate, latency, connection count
- **Alerts**: High error rate, connection drops, DB latency spikes
- **Dashboards**: Real-time message throughput, active users

---

## 3. Design YouTube

### Functional Requirements
- Upload videos
- Watch videos (streaming)
- Search videos
- Like, comment, subscribe
- Recommendations

### Non-Functional Requirements
- 2 billion users
- 500 hours of video uploaded per minute
- 1 billion hours watched per day
- Low latency streaming (<2s buffering)
- 99.9% availability

### Capacity Estimation
- **Upload**: 500 hours/min × 60 min = 30,000 hours/min
  - Avg video size: 1GB/hour → 30TB/min = 43PB/day
- **Streaming**: 1B hours/day ÷ 24 = 41.6M concurrent streams
- **Bandwidth**: 41.6M × 5Mbps = 208 Tbps

### High-Level Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
┌──────▼──────────────────────────────────────┐
│         CDN (Akamai/CloudFront)              │
│    (Video streaming, thumbnails)             │
└──────┬──────────────────────────────────────┘
       │
┌──────▼──────────────────────────────────────┐
│      Load Balancer                           │
└──────┬──────────────────────────────────────┘
       │
       ├─────────────────┬─────────────────┬──────────────┐
       │                 │                 │              │
┌──────▼──────┐  ┌──────▼──────┐  ┌──────▼──────┐  ┌───▼────┐
│   Upload    │  │  Streaming  │  │   Search    │  │ Recom- │
│  Service    │  │   Service   │  │  Service    │  │mendation│
└──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └───┬────┘
       │                 │                 │              │
┌──────▼──────┐  ┌──────▼──────┐  ┌──────▼──────┐  ┌───▼────┐
│   Video     │  │   Metadata  │  │  Elastic    │  │  ML    │
│  Storage    │  │     DB      │  │  Search     │  │ Models │
│   (S3)      │  │  (MySQL)    │  └─────────────┘  └────────┘
└──────┬──────┘  └─────────────┘
       │
┌──────▼──────────────────────────────────────┐
│      Video Processing Pipeline               │
│  (Transcoding, Thumbnail, Metadata)          │
└──────────────────────────────────────────────┘
```

### Key Components

#### 1. Video Upload Service
```
Flow:
1. Client requests upload URL
2. API generates S3 pre-signed URL
3. Client uploads raw video to S3 (multipart upload)
4. S3 triggers Lambda → SQS queue
5. Transcoding workers pick from queue
6. Transcode to multiple formats:
   - 360p, 480p, 720p, 1080p, 4K
   - Codecs: H.264, VP9, AV1
7. Generate thumbnails (3 frames)
8. Extract metadata (duration, resolution)
9. Store processed videos in S3
10. Update metadata DB
11. Index in Elasticsearch
12. Notify user (upload complete)
```

**Transcoding Pipeline (AWS Elastic Transcoder / FFmpeg)**:
```python
def transcode_video(video_id, s3_key):
    resolutions = [
        {"name": "360p", "width": 640, "height": 360, "bitrate": "500k"},
        {"name": "720p", "width": 1280, "height": 720, "bitrate": "2500k"},
        {"name": "1080p", "width": 1920, "height": 1080, "bitrate": "5000k"}
    ]
    
    for res in resolutions:
        output_key = f"{video_id}/{res['name']}.mp4"
        
        # FFmpeg command
        cmd = f"""
        ffmpeg -i s3://{BUCKET}/{s3_key} \
               -vf scale={res['width']}:{res['height']} \
               -b:v {res['bitrate']} \
               -c:v libx264 \
               -c:a aac \
               s3://{BUCKET}/{output_key}
        """
        
        subprocess.run(cmd, shell=True)
    
    # Generate thumbnails
    generate_thumbnails(video_id, s3_key)
    
    # Update DB
    mysql.execute(
        "UPDATE videos SET status = 'ready', processed_at = NOW() WHERE id = ?",
        video_id
    )
```

#### 2. Video Streaming Service

**Adaptive Bitrate Streaming (HLS/DASH)**:
```
1. Client requests video manifest (.m3u8)
2. Server returns playlist with multiple quality levels
3. Client measures bandwidth
4. Client requests appropriate quality chunks
5. Dynamically switch quality based on network

Manifest Example:
#EXTM3U
#EXT-X-STREAM-INF:BANDWIDTH=500000,RESOLUTION=640x360
360p/playlist.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=2500000,RESOLUTION=1280x720
720p/playlist.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=5000000,RESOLUTION=1920x1080
1080p/playlist.m3u8
```

**CDN Strategy**:
- Popular videos cached at edge locations
- Long-tail videos served from origin
- Cache invalidation on video update

#### 3. Recommendation System
```python
# Collaborative Filtering + Content-Based
def get_recommendations(user_id, limit=20):
    # User's watch history
    history = get_watch_history(user_id)
    
    # Similar users (collaborative filtering)
    similar_users = find_similar_users(user_id, history)
    
    # Videos watched by similar users
    collab_recs = get_videos_from_users(similar_users)
    
    # Content-based (similar to watched videos)
    content_recs = []
    for video in history:
        similar_videos = find_similar_videos(video.id)
        content_recs.extend(similar_videos)
    
    # Merge and rank
    combined = merge_and_rank(collab_recs, content_recs)
    
    # Apply business rules
    filtered = apply_filters(combined, user_id)
    
    return filtered[:limit]

# ML Model (Matrix Factorization)
def train_recommendation_model():
    # User-Video interaction matrix
    interactions = get_user_video_matrix()
    
    # Factorize into user and video embeddings
    model = MatrixFactorization(
        n_users=len(users),
        n_videos=len(videos),
        n_factors=100
    )
    
    model.fit(interactions, epochs=10)
    
    # Store embeddings in vector DB (Pinecone)
    store_embeddings(model.user_embeddings, model.video_embeddings)
```

### Database Schema

**Videos Table**
```sql
videos (
  id BIGINT PRIMARY KEY,
  title VARCHAR(200),
  description TEXT,
  uploader_id BIGINT,
  duration INT,
  view_count BIGINT DEFAULT 0,
  like_count INT DEFAULT 0,
  dislike_count INT DEFAULT 0,
  status VARCHAR(20),  -- processing, ready, failed
  upload_date TIMESTAMP,
  INDEX(uploader_id, upload_date),
  INDEX(view_count DESC)
)
```

**Watch History (Cassandra)**
```
watch_history (
  user_id BIGINT,
  video_id BIGINT,
  watched_at TIMESTAMP,
  watch_duration INT,
  PRIMARY KEY((user_id), watched_at, video_id)
)
```

### Scalability Considerations

1. **Storage Optimization**:
   - Compress videos (H.265 saves 50% vs H.264)
   - Delete unpopular videos after 1 year
   - Tiered storage (hot/warm/cold)

2. **Caching**:
   - Video metadata in Redis
   - Popular videos in CDN
   - Search results cached

3. **Database Sharding**:
   - Shard videos by video_id
   - Shard users by user_id
   - Separate read replicas for analytics

4. **Async Processing**:
   - View count updates (batch every 5 min)
   - Recommendation model training (daily)
   - Analytics (Spark jobs)

---

## 4. Design Uber

### Functional Requirements
- Riders request rides
- Drivers accept rides
- Real-time location tracking
- ETA calculation
- Fare calculation
- Payment processing

### Non-Functional Requirements
- 100M daily active users
- 15M rides per day
- Real-time updates (<1s)
- 99.99% availability
- Accurate location tracking

### High-Level Architecture

```
┌─────────────┐          ┌─────────────┐
│   Rider     │          │   Driver    │
│     App     │          │     App     │
└──────┬──────┘          └──────┬──────┘
       │                        │
       └────────┬───────────────┘
                │
┌───────────────▼────────────────────────┐
│      WebSocket Gateway                 │
│  (Real-time location updates)          │
└───────────────┬────────────────────────┘
                │
┌───────────────▼────────────────────────┐
│      Location Service                  │
│  (Redis Geo, QuadTree)                 │
└───────────────┬────────────────────────┘
                │
       ├────────┴────────┬──────────────┐
       │                 │              │
┌──────▼──────┐  ┌──────▼──────┐  ┌───▼────┐
│   Matching  │  │   Trip      │  │Payment │
│   Service   │  │  Service    │  │Service │
└──────┬──────┘  └──────┬──────┘  └───┬────┘
       │                 │              │
┌──────▼──────┐  ┌──────▼──────┐  ┌───▼────┐
│  Drivers    │  │   Trips     │  │Payments│
│     DB      │  │     DB      │  │   DB   │
└─────────────┘  └─────────────┘  └────────┘
```

### Key Algorithms

#### 1. Driver Matching (Geospatial Search)
```python
def find_nearby_drivers(rider_lat, rider_lon, radius_km=5):
    # Redis Geo commands
    nearby = redis.georadius(
        "drivers:online",
        rider_lon, rider_lat,
        radius_km, "km",
        withdist=True,
        sort="ASC",
        count=10
    )
    
    # Filter by availability
    available = []
    for driver_id, distance in nearby:
        status = redis.hget(f"driver:{driver_id}", "status")
        if status == "available":
            available.append({
                "driver_id": driver_id,
                "distance": distance
            })
    
    return available

# QuadTree for efficient spatial indexing
class QuadTree:
    def __init__(self, boundary, capacity=4):
        self.boundary = boundary  # Rectangle
        self.capacity = capacity
        self.drivers = []
        self.divided = False
    
    def insert(self, driver):
        if not self.boundary.contains(driver.location):
            return False
        
        if len(self.drivers) < self.capacity:
            self.drivers.append(driver)
            return True
        
        if not self.divided:
            self.subdivide()
        
        return (self.northeast.insert(driver) or
                self.northwest.insert(driver) or
                self.southeast.insert(driver) or
                self.southwest.insert(driver))
    
    def query_range(self, range_rect):
        found = []
        if not self.boundary.intersects(range_rect):
            return found
        
        for driver in self.drivers:
            if range_rect.contains(driver.location):
                found.append(driver)
        
        if self.divided:
            found.extend(self.northeast.query_range(range_rect))
            found.extend(self.northwest.query_range(range_rect))
            found.extend(self.southeast.query_range(range_rect))
            found.extend(self.southwest.query_range(range_rect))
        
        return found
```

#### 2. ETA Calculation
```python
def calculate_eta(start_lat, start_lon, end_lat, end_lon):
    # Use Google Maps Directions API
    response = maps_api.directions(
        origin=(start_lat, start_lon),
        destination=(end_lat, end_lon),
        mode="driving",
        departure_time="now",  # Considers traffic
        traffic_model="best_guess"
    )
    
    duration = response['routes'][0]['legs'][0]['duration_in_traffic']['value']
    distance = response['routes'][0]['legs'][0]['distance']['value']
    
    return {
        "eta_seconds": duration,
        "distance_meters": distance
    }

# Fallback: Haversine distance + avg speed
def estimate_eta_fallback(start, end):
    distance_km = haversine_distance(start, end)
    avg_speed_kmh = 30  # City average
    eta_hours = distance_km / avg_speed_kmh
    return eta_hours * 3600  # Convert to seconds
```

#### 3. Surge Pricing
```python
def calculate_surge_multiplier(lat, lon, timestamp):
    # Get demand (ride requests) in area
    demand = redis.zcount(
        f"requests:{get_geohash(lat, lon)}",
        timestamp - 300,  # Last 5 minutes
        timestamp
    )
    
    # Get supply (available drivers)
    supply = len(find_nearby_drivers(lat, lon, radius_km=2))
    
    # Calculate multiplier
    if supply == 0:
        return 3.0  # Max surge
    
    ratio = demand / supply
    
    if ratio < 1:
        return 1.0  # No surge
    elif ratio < 2:
        return 1.5
    elif ratio < 3:
        return 2.0
    else:
        return 3.0  # Max surge
```

### Real-Time Location Tracking
```python
# Driver sends location every 5 seconds
def update_driver_location(driver_id, lat, lon):
    # Update Redis Geo
    redis.geoadd("drivers:online", lon, lat, driver_id)
    
    # Update trip if active
    trip_id = redis.hget(f"driver:{driver_id}", "active_trip")
    if trip_id:
        # Store location history
        redis.zadd(
            f"trip:{trip_id}:route",
            {f"{lat},{lon}": time.time()}
        )
        
        # Notify rider via WebSocket
        rider_id = get_rider_for_trip(trip_id)
        websocket_send(rider_id, {
            "type": "driver_location",
            "lat": lat,
            "lon": lon
        })
```

This covers the major system design questions. Would you like me to add more questions or go deeper into any specific area?

