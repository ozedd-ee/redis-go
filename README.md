# redis-go
A lite version of the Redis server written in Go with support for commands from the first version of Redis.

## 🚀 Getting Started

### If you have Go installed

Use the provided `Makefile` for common tasks:

#### 📦 Build the server

```bash
make build
```

#### 🟢 Start the server (foreground)

```bash
make run
```

#### 🧪 Run unit tests

```bash
make test
```

#### ⚙️ Run benchmarks

```bash
make bench
```

---

## 🐳 If you don’t have Go installed (use Docker)

#### 📦 Build Docker image

```bash
make docker-build
```

#### 🚀 Run the server container

```bash
make docker-run
```

#### 🚫 Stop the container

```bash
make docker-stop
```

---

##  Connecting via Redis CLI

Use any Redis client (e.g., `redis-cli`, `rdcli`, or GUI tools like RedisInsight).

### Install `redis-cli` (optional):

```bash
npm install -g redis-cli
```

### Connect to the server:

```bash
rdcli
```

Once connected, you can run any supported command.

---

## ✅ Supported Commands

```
PING, ECHO, INFO, SET, GET, EXISTS, DEL, INCR, DECR, LPUSH, RPUSH, LRANGE
```

## 🕒 Supported Expiry Options

```
EX, EXAT, PX, PXAT
```

---
