# Matching Engine Service

A high-performance order matching engine service that processes buy and sell orders using Kafka for message handling.

## Prerequisites

- Docker and Docker Compose
- Go 1.23 or later (for local development)
- `kafkacat` or `kcat` (for testing)

## Quick Start

### 1. Start the Services

```bash
docker-compose up -d
```

This will start:
- Kafka broker
- Zookeeper
- Matching Service

### 2. Create Required Topics

After services are up, create the required Kafka topics:

```bash
# Create orders topic
docker exec matching-engine-kafka /usr/bin/kafka-topics \
    --create \
    --bootstrap-server localhost:9092 \
    --topic orders \
    --partitions 1 \
    --replication-factor 1  

# Create matches topic
docker exec matching-engine-kafka /usr/bin/kafka-topics \
    --create \
    --bootstrap-server localhost:9092 \
    --topic matches \
    --partitions 1 \
    --replication-factor 1 
```

You should see both `orders` and `matches` topics in the list.

To check topic details:
```bash
# Check orders topic configuration
docker exec matching-engine-kafka /usr/bin/kafka-topics \
    --describe \
    --bootstrap-server localhost:9092 \
    --topic orders  

# Check matches topic configuration
docker exec matching-engine-kafka /usr/bin/kafka-topics \
    --describe \
    --bootstrap-server localhost:9092 \
    --topic matches  
```

### 3. Check Services Status

```bash
docker-compose ps
```

### 4. Publishing Test Orders

You can publish orders using kafkacat/kcat:

```bash
# Publish a buy order
echo '{"ID":"order1","Side":"BUY","Price":100.50,"Quantity":10,"Timestamp":"'$(date -u +"%Y-%m-%dT%H:%M:%S.%NZ")'"}' | kcat -b localhost:9092 -t orders -P  


# Publish a sell order
echo '{"ID":"order2","Side":"SELL","Price":100.50,"Quantity":10,"Timestamp":"'$(date -u +"%Y-%m-%dT%H:%M:%S.%NZ")'"}' | kcat -b localhost:9092 -t orders -P  
```

### 5. Monitor Match Events

```bash
kcat -b localhost:9092 -t matches -C
```

## Configuration

Configuration is handled through environment variables in the docker-compose.yml file:

```yaml
environment:
  - KAFKA_BROKERS=kafka:9092
  - KAFKA_GROUP_ID=matching-engine
  - KAFKA_ORDERS_TOPIC=orders
  - KAFKA_MATCHES_TOPIC=matches
```

## Message Formats

### Order Message

```json
{
  "id": "string",       // Unique order ID
  "side": "string",     // "BUY" or "SELL"
  "price": float,       // Order price
  "quantity": int     // Order quantity
}
```

### Match Event

```json
{
  "orderID": "string",  // Original order ID
  "matchID": "string",  // Unique match ID
  "price": float,       // Match price
  "quantity": int     // Matched quantity
}
```

## Project Structure

```
.
├── cmd/
│   └── matching/
│       └── main.go           # Service entry point
├── internal/
│   ├── application/
│   │   └── service/          # Application services
│   ├── domain/
│   │   ├── entities/         # Domain entities
│   │   └── ports/           # Interface definitions
│   └── infrastructure/
│       ├── messaging/        # Kafka implementation
│       ├── orderbook/        # Order book implementation
│       └── repository/       # Data storage
├── pkg/
│   ├── common/
│   │   └── config/          # Configuration
│   └── messaging/
│       └── kafka/           # Kafka utilities
├── docker-compose.yml
└── README.md
```

## Local Development

1. Clone the repository:
```bash
git clone https://github.com/mehdivijeh/matching-engine.git
```

2. Install dependencies:
```bash
go mod download
```

3. Build the service:
```bash
go build -o matching-engine ./cmd/matching
```

## Monitoring and Logs

View service logs:
```bash
# View matching engine logs
docker-compose logs -f matching-engine

# View Kafka logs
docker-compose logs -f kafka

# View all logs
docker-compose logs -f
```