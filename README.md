# Go-Redis

A Redis implementation written in Go. The goal of this project is to delve into the internal workings of Redis, including the RESP protocol, command handling, and replication.

## Current Features

*   **Concurrent TCP Server:** Capable of handling multiple client connections simultaneously.
*   **RESP Parser:** Implementation of a parser for the Redis Serialization Protocol (RESPv2).
*   **Command Handling:** Support for a subset of basic commands:
    *   `PING`: Checks the connection with the server.
    *   `ECHO`: Returns the provided message.
    *   `SET`: Stores a key-value pair.
    *   `GET`: Retrieves the value associated with a key.
    *   `LPUSH, RPUSH`: Stores a key-list.
    *   `XRANGE`: Retrieves list data associated with a key.
    *   `INFO`: Provides information about the server (replication section).
    *   ...etc.
*   **Replication:** Basic master-slave replication functionality.
*   **In-Memory Storage:** A simple in-memory key-value store.

## Getting Started

### Prerequisites

- Go 1.18 or higher.
- `redis-cli` to interact with the server (optional).

### Installation and Running

1.  **Clone the repository (if applicable):**
    ```sh
    git clone <repository-url>
    cd go-redis
    ```

2.  **Build the executable:**
    ```sh
    go build -o ./redis-server main.go
    ```

3.  **Run the server:**
    ```sh
    ./redis-server
    ```
    By default, the server will start and listen on port `6379`.

### Using `redis-cli`

You can connect to the server using the Redis command-line tool:

```sh
redis-cli
```

Once connected, you can test the implemented commands:

```
127.0.0.1:6379> PING
PONG
127.0.0.1:6379> SET hello world
OK
127.0.0.1:6379> GET hello
"world"
127.0.0.1:6379> ECHO "Hello Gemini"
"Hello Gemini"
```

## Project Structure

```
/
├── main.go                 # Application entry point
├── app/                    # Main orchestration logic
├── command/                # Command extraction and handling
├── network/                # TCP server implementation
├── protocol/               # RESP protocol parser
├── redis/                  # Core Redis instance logic
├── replica/                # Replication logic
├── storage/                # In-memory data storage
└── utils/                  # Utility functions (e.g., cryptography)
```
