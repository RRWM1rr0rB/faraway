# Word of Window

## Test task for Server Engineer (Roman Zaitsau)

### 📜 Description
Design and implement “Word of Wisdom” TCP server.

• TCP server should be protected from DDoS attacks with Proof of Work (https://en.wikipedia.org/wiki/Proof_of_work), using a challenge-response protocol.
• The choice of the PoW algorithm should be explained.
• After Proof of Work verification, the server should send one of the quotes from the “Word of Wisdom” book or another collection of quotes.
• Docker files should be provided for both the server and the client that solves the PoW challenge.

### ✨ Features:

🔐 **PoW protection**: Prevents brute-force and bot-based DDoS attacks.
⚡ **SHA-256 based challenge**: Adjustable difficulty depending on server load.
📖 **Quote delivery**: Once PoW is verified, the server sends a quote from the "Word of Wisdom" collection.
🐳 **Docker support**: Dockerized setup for both server and client.

### 🛡️ Proof of Work (PoW) Algorithm Choice

When selecting a PoW algorithm, the main consideration is the type of attackers we aim to defend against. Typically, small hacker groups, competitors, or individual attackers cannot afford high-cost servers. Even if they do, they need to migrate quickly from one server to another, making an efficient and adaptable PoW crucial.

❓ **Why SHA-256?**
✔️ **Security & Performance**: Unlike SHA-1, which is outdated and insecure, SHA-256 offers strong cryptographic security.
✔️ **Avoiding Self-DDoS**: Algorithms like Scrypt and Argon2, while effective against bots, are too resource-intensive and could overload our own server.
✔️ **Dynamic Difficulty Adjustment**:
📈 Increased load → higher difficulty.
🚫 High requests from a single IP → adaptive difficulty increase.
❌ Persistent offenders → temporary IP ban (e.g., 24 hours).

| ⚙️ Algorithm | 🏗️ Type         | 📱 Mobile-Friendly?       |   ⚖️ Balance of Difficulty   | 🛡️ Protection Against Bots      |
|:----------:|:-------------|:-------------------------|:----------------------------------:|:----------------------------|  
|  SHA-256   | 🖥️ CPU-bound   | ✅ Yes                    |  ⚠️ Requires manual tuning | ❌ ASIC miners can bypass |
|   Scrypt   | 🧠 Memory-bound | ⚠️ High memory use |               ✅ Yes                | ✅ Strong protection         |
|   Argon2   | 🧠 Memory-bound | ❌ Heavy on phones     |    ✅ Excellent              | ✅ Best protection      |
|  Hashcash  | 🖥️ CPU-bound | ✅ Yes                    |  ✅ Easy to adjust  | ⚠️ Moderate protection        |

## 📂 Project Structure

```
├── app-client
│   ├── app
│   │   ├── cmd
│   │   │   └── client
│   │   │       └── main.go
│   │   └── internal
│   │       ├── app
│   │       │   ├── app.go
│   │       │   └── setup.go
│   │       ├── config
│   │       │   ├── config.go
│   │       │   └── constants.go
│   │       ├── controller
│   │       │   └── tcp
│   │       │       └── v1
│   │       │           └── mitigator
│   │       │               ├── controller.go
│   │       │               └── server.go
│   │       └── policy
│   │           └── mitigator
│   │               ├── dto.go
│   │               ├── policy.go
│   │               └── policy_migrator.go
│   ├── go.mod
│   └── go.sum
├── app-server
│   ├── app
│   │   ├── cmd
│   │   │   └── server
│   │   │       └── main.go
│   │   ├── internal
│   │   │   ├── app
│   │   │   │   ├── app.go
│   │   │   │   └── setup.go
│   │   │   ├── config
│   │   │   │   ├── config.go
│   │   │   │   └── constants.go
│   │   │   ├── controller
│   │   │   │   └── tcp
│   │   │   │       └── mitigator
│   │   │   ├── domain
│   │   │   │   └── mitigator
│   │   │   │       ├── model
│   │   │   │       │   └── model.go
│   │   │   │       ├── service
│   │   │   │       │   └── service.go
│   │   │   │       └── storage
│   │   │   │           └── redis
│   │   │   │               └── storage.go
│   │   │   └── policy
│   │   │       ├── base.go
│   │   │       └── mitigator
│   │   │           ├── dto.go
│   │   │           ├── error.go
│   │   │           ├── policy.go
│   │   │           └── policy_mitigator.go
│   │   └── pkg
│   │       └── pow
│   │           └── algo.go
│   ├── go.mod
│   └── go.sum
├── configs
│   ├── config.client.local.yaml
│   ├── config.server.local.yaml
│   └── docker-compose
│       ├── docker-compose.client.local.yaml
│       └── docker-compose.server.local.yaml
├── deploy
│   ├── Dockerfile.client.dockerfile
│   └── Dockerfile.server.dockerfile
```

---

### 📂 Faraway lib Structure.

```
        ├── core
        │   ├── array
        │   │   └── array.go
        │   ├── blacklist
        │   │   └── blacklist.go
        │   ├── bytes
        │   │   └── bytes.go
        │   ├── clock
        │   │   ├── clock.go
        │   │   └── interface.go
        │   ├── closer
        │   │   └── closer.go
        │   ├── encryption
        │   │   └── sha-256
        │   │       └── sha_256.go
        │   ├── go.mod
        │   ├── go.sum
        │   ├── pointer
        │   │   └── pointer.go
        │   ├── random
        │   │   └── random.go
        │   ├── repeat
        │   │   ├── repeat.go
        │   │   ├── repeat_http.go
        │   │   └── repeat_ws.go
        │   ├── safe
        │   │   ├── errorgroup
        │   │   │   └── errorgroup.go
        │   │   ├── safe.go
        │   │   └── waitgroup
        │   │       └── waitgroup.go
        │   ├── tcp
        │   │   ├── client.go
        │   │   ├── error.go
        │   │   ├── middleware.go
        │   │   ├── options.go
        │   │   ├── pool.go
        │   │   ├── pow.go
        │   │   ├── retry.go
        │   │   ├── server.go
        │   │   └── tls.go
        │   ├── time
        │   │   └── time.go
        │   ├── uuid
        │   │   ├── db
        │   │   │   └── uuid.go
        │   │   ├── google_uuid
        │   │   │   ├── interface.go
        │   │   │   ├── ulid.go
        │   │   │   └── uuid.go
        │   │   ├── network
        │   │   │   └── uuid.go
        │   │   ├── uuid.go
        │   │   └── uuid_test.go
        │   └── version
        ├── errors
        │   ├── errors.go
        │   ├── go.mod
        │   ├── go.sum
        │   └── version
        ├── logging
        │   ├── alias.go
        │   ├── context.go
        │   ├── go.mod
        │   ├── go.sum
        │   ├── logger.go
        │   ├── logger_test.go
        │   ├── middleware.go
        │   └── version
        ├── main.go
        ├── metrics
        │   ├── config.go
        │   ├── go.mod
        │   ├── go.sum
        │   ├── grpc_middleware.go
        │   ├── handler.go
        │   ├── http_middleware.go
        │   ├── metrics.go
        │   ├── metrics_grpc_availability.go
        │   ├── metrics_test.go
        │   └── version
        ├── pprof
        │   ├── config.go
        │   ├── go.mod
        │   ├── server.go
        │   ├── server_test.go
        │   └── version
        ├── redis
        │   ├── aliases.go
        │   ├── error.go
        │   ├── go.mod
        │   ├── go.sum
        │   ├── metrics.go
        │   ├── redis.go
        │   └── version
        └── tracing
            ├── attrs.go
            ├── go.mod
            ├── go.sum
            ├── middleware.go
            ├── tracing.go
            ├── tracing_config.go
            └── version
```


This README now includes a more structured description, project features, and an improved file tree display in markdown format. If you need further refinements or explanations, let me know! 🚀
