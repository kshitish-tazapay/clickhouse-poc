
# ClickHouse Replication with ClickHouse Keeper (Docker Compose)

This project demonstrates how to set up a replicated ClickHouse cluster using ClickHouse Keeper instead of ZooKeeper, running entirely on a single machine with Docker Compose. This setup mimics the ClickHouse replication architecture as described at: [ClickHouse Docs](https://clickhouse.com/docs/architecture/replication)

## 🧱 Topology

- 3 ClickHouse Keeper nodes
- 1 shard with 2 ClickHouse replicas

## 📁 Project Structure

```
clickhouse-replication-keeper/
├── docker-compose.yml
├── configs/
│   ├── clickhouse/
│   │   ├── macros1.xml
│   │   ├── macros2.xml
│   │   ├── use-keeper.xml
│   │   ├── remote-servers.xml
│   │   └── network-and-logging.xml
│   └── keeper/
│       ├── keeper1.xml
│       ├── keeper2.xml
│       └── keeper3.xml
```

## 🚀 Setup on EC2

### 1. Launch EC2

- Use Ubuntu 22.04 or Amazon Linux 2.
- Open ports: 22, 8123, 9000, 9001, 8124, 9181-9183.

### 2. SSH into Instance

```bash
ssh -i /path/to/key.pem ec2-user@<EC2-IP>
```

### 3. Install Docker & Compose

**Ubuntu:**
```bash
sudo apt update
sudo apt install -y docker.io docker-compose
sudo usermod -aG docker $USER
newgrp docker
```

**Amazon Linux 2:**
```bash
sudo yum update -y
sudo yum install -y docker
sudo service docker start
sudo usermod -a -G docker ec2-user
newgrp docker

sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
docker-compose --version
```

### 4. Clone the Repository

```bash
git clone https://github.com/one2nc/clickhouse-playground.git
cd clickhouse-playground/cluster-setup-using-docker
```

### 5. Start Cluster

```bash
docker-compose up -d
```

## ✅ Verify Replication

```bash
docker exec -it clickhouse1 clickhouse-client
```

```sql
SELECT * FROM system.clusters;
SELECT * FROM system.replicas;
```

## 📦 Populate data

1. Connect to `clickhouse1`
    ```bash
    docker exec -it clickhouse1 clickhouse-client
    ```

2. Create the database `uk`
    ```sql
    CREATE DATABASE uk ON CLUSTER replicated_cluster;
    ```

3. Create table
    ```sql
    CREATE TABLE uk.uk_price_paid ON CLUSTER replicated_cluster
    (
        price UInt32,
        date Date,
        postcode1 LowCardinality(String),
        postcode2 LowCardinality(String),
        type Enum8('terraced' = 1, 'semi-detached' = 2, 'detached' = 3, 'flat' = 4, 'other' = 0),
        is_new UInt8,
        duration Enum8('freehold' = 1, 'leasehold' = 2, 'unknown' = 0),
        addr1 String,
        addr2 String,
        street LowCardinality(String),
        locality LowCardinality(String),
        town LowCardinality(String),
        district LowCardinality(String),
        county LowCardinality(String)
    )
    ENGINE = ReplicatedMergeTree
    ORDER BY (postcode1, postcode2, addr1, addr2);
    ```

4. Insert data from the url
    ```sql
    INSERT INTO uk.uk_price_paid
    SELECT
        toUInt32(price_string) AS price,
        parseDateTimeBestEffortUS(time) AS date,
        splitByChar(' ', postcode)[1] AS postcode1,
        splitByChar(' ', postcode)[2] AS postcode2,
        transform(a, ['T', 'S', 'D', 'F', 'O'], ['terraced', 'semi-detached', 'detached', 'flat', 'other']) AS type,
        b = 'Y' AS is_new,
        transform(c, ['F', 'L', 'U'], ['freehold', 'leasehold', 'unknown']) AS duration,
        addr1,
        addr2,
        street,
        locality,
        town,
        district,
        county
    FROM url(
        'http://prod1.publicdata.landregistry.gov.uk.s3-website-eu-west-1.amazonaws.com/pp-complete.csv',
        'CSV',
        'uuid_string String,
        price_string String,
        time String,
        postcode String,
        a String,
        b String,
        c String,
        addr1 String,
        addr2 String,
        street String,
        locality String,
        town String,
        district String,
        county String,
        d String,
        e String'
    ) SETTINGS max_http_get_redirects=10;
    ```

5. Check if data is present
    ```sql
    SELECT count() FROM uk.uk_price_paid;

    SELECT formatReadableSize(total_bytes) FROM system.tables WHERE name = 'uk_price_paid'
    ```

6. Connect to `clickhouse2`
    ```bash
    docker exec -it clickhouse2 clickhouse-client
    ```

7. Check if data is present.
    ```sql
    SELECT count() FROM uk.uk_price_paid;

    SELECT formatReadableSize(total_bytes) FROM system.tables WHERE name = 'uk_price_paid'
    ```

## 🏃‍♂️ Run the Go Client

A simple Go client is included to demonstrate connecting to the ClickHouse cluster and running queries from your local machine.

### 1. Install Go (if not already installed)

Follow instructions at [golang.org/doc/install](https://golang.org/doc/install).

### 2. Update the ClickHouse Host

Edit `go-client/main.go` and replace `<ec2-ip>` with your EC2 instance's public IP address in the ClickHouse connection string.

### 3. Install Dependencies

```bash
cd go-client
go mod tidy
```

### 4. Run the Example

```bash
go run main.go
```

The client connects to `clickhouse1` at your EC2 IP on port `9000` and prints query results.


## 🧹 Tear Down

```bash
docker-compose down -v
```
