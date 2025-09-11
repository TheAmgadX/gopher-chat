# 🐹 gopher-chat

A **real-time chat server written in Go**.
`gopher-chat` allows multiple clients to connect, create or join rooms, and exchange messages with simple text commands.

---

## 🚀 Features

* Multiple clients can connect over TCP.
* Supports **chat rooms** — create, join, leave.
* Simple **command-based interface**.
* Concurrency-safe design using goroutines and channels.
* Easily extendable into a distributed system.

---

## 💬 Commands

Once connected, clients can use these commands:

| Command            | Description                                                  |
| ------------------ | ------------------------------------------------------------ |
| `/username <name>` | Set or change your username. Default is `guest-<id>`.        |
| `/join <room>`     | Join a chat room. Creates the room if it doesn’t exist.      |
| `/rooms`           | List all available chat rooms.                               |
| `/quit`            | Leave the current room (you remain connected to the server). |
| `/msg <text>`      | Send a message to all users in your current room.            |

---

## 🛠️ Getting Started

### Prerequisites

* [Go 1.23+](https://go.dev/dl/) installed on your system.

### Clone the repo

```bash
git clone https://github.com/TheAmgadX/gopher-chat.git
cd gopher-chat
```

### Run the server

```bash
go run ./cmd/server
```

The server will start on **`localhost:8080`** (or the configured port).

### Connect with netcat (example client)

In one terminal:

```bash
nc localhost 8080
```

In another terminal:

```bash
nc localhost 8080
```

Now you can chat between the two terminals using the commands above.

---

## 📂 Project Structure

```
gopher-chat/
├── cmd/
│   └── server/
│       └── main.go       # Entry point
├── internal/
│   ├── server/           # Core chat server logic
│   │   ├── client.go
│   │   ├── command.go
│   │   ├── room.go
│   │   └── server.go
│   └── utils/            # Helpers (logging, config, etc.)
├── go.mod
└── README.md
```
---

## 📜 License

MIT License © 2025 [TheAmgadX](https://github.com/TheAmgadX)