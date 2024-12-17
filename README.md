# redis-go
A lite version of the Redis server written in Go with support for commands from the first version of Redis.

### If you have Go installed and set up:

Run: 
```bash
go install
```

Usage:

Start the server and listen for client connections:
```bash 
redis-go
```

Connect to the server from any Redis client of your choice:

You can use the following guide from Redis to  [install redis-cli without installing the server.](https://redis.io/blog/get-redis-cli-without-installing-redis-server/)

Too long to read? Just run:
```bash
npm install -g redis-cli
```

To start the CLI client and connect to the server, make sure the server is still running, open a new terminal and run:
```bash
rdcli
```
Pass commands through the CLI client and get responses.


## Available commands
PING, ECHO, INFO, SET GET, EXISTS, DEL, INCR, DECR, LPUSH, RPUSH, LRANGE

### Available expiry options
EX, EXAT, PX, PXAT

## If you don't have Go installed:

Usage:
```bash
redis-go.exe 
```

Follow every other step listed above
