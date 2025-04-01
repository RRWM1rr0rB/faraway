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
│   │   ├── cmd -- cmd file for running app.
│   │   │   └── client
│   │   │       └── main.go
│   │   └── internal
│   │       ├── app -- all setup extentions and runner.
│   │       │   ├── app.go -- runner.
│   │       │   └── setup.go -- setup extentions.
│   │       ├── config
│   │       │   ├── config.go -- viper cfg setup(next we can use cobra for cli).
│   │       │   └── constants.go -- global constants for project.
│   │       ├── controller -- logic for communication with server.
│   │       │   └── tcp
│   │       │       └── v1
│   │       │           └── mitigator
│   │       │               ├── controller.go -- Logigic for communication with server.
│   │       │               └── server.go -- Constructor and interface.
│   │       └── policy
│   │           └── mitigator
│   │               ├── dto.go -- Model with struct.
│   │               ├── policy.go -- Constructor and initializer.
│   │               └── policy_migrator.go -- Buisness logic for policy.
│   ├── go.mod
│   └── go.sum
├── app-server
│   ├── app
│   │   ├── cmd -- cmd file for running app.
│   │   │   └── server
│   │   │       └── main.go
│   │   ├── internal
│   │   │   ├── app -- all setup extentions and runner.
│   │   │   │   ├── app.go -- runner.
│   │   │   │   └── setup.go -- setup extentions.
│   │   │   ├── config
│   │   │   │   ├── config.go - viper cfg setup(next we can use cobra for cli).
│   │   │   │   └── constants.go -- global constants for project.
│   │   │   ├── controller
│   │   │   │   └── tcp
│   │   │   │       └── mitigator - logic for communication with client.
│   │   │   │           ├── controller.go -- Logigic for communication with client.
│   │   │   │           └── server.go -- Constructor and interface.
│   │   │   └── policy
│   │   │       └── mitigator
│   │   │           ├── dto.go -- Model with struct.
│   │   │           ├── error.go -- Custom errors.
│   │   │           ├── policy.go -- Constructor and initializer.
│   │   │           └── policy_mitigator.go -- Buisness logic for policy.
│   ├── go.mod
│   └── go.sum
├── configs
│   ├── config.client.local.yaml -- config for client.
│   ├── config.server.local.yaml -- config for server.
|
├── deploy
│   ├── Dockerfile.client.dockerfile -- Dockerfile for client.
│   └── Dockerfile.server.dockerfile -- Dockerfile for server.
```

---

### 📂 Faraway lib Structure.

```
        ├── core
        │   ├── array
        │   │   └── array.go -- don't use
        │   ├── blacklist
        │   │   └── blacklist.go -- logic for blacklist.
        │   ├── bytes
        │   │   └── bytes.go -- logic for bytes.
        │   ├── clock
        │   │   ├── clock.go -- logic for clock.
        │   │   └── interface.go -- interface for clock.
        │   ├── closer
        │   │   └── closer.go -- logic for closer.
        │   ├── encryption
        │   │   └── sha-256 -- logic for encryption.
        │   │       └── sha_256.go
        │   ├── go.mod
        │   ├── go.sum
        │   ├── pointer -- logic for pointer.
        │   │   └── pointer.go
        │   ├── random
        │   │   └── random.go -- logic for random.
        │   ├── repeat
        │   │   ├── repeat.go -- logic for repeat.
        │   │   ├── repeat_http.go -- logic for repeat.
        │   │   └── repeat_ws.go -- logic for repeat.
        │   ├── safe
        │   │   ├── errorgroup
        │   │   │   └── errorgroup.go -- logic for errorgroup.
        │   │   ├── safe.go -- logic for safe.
        │   │   └── waitgroup
        │   │       └── waitgroup.go -- logic for waitgroup.
        │   ├── tcp ------------------- WORK WITH TCP(BLACK LIST, COUNT CONNECTION AND OTHER(CLIENT AND SERVER))
        │   │   ├── client.go
        │   │   ├── error.go
        │   │   ├── middleware.go
        │   │   ├── options.go
        │   │   ├── pool.go
        │   │   ├── pow.go
        │   │   ├── retry.go
        │   │   ├── server.go
        │   │   └── tls.go
        │   ├── time -- logic for time.
        │   │   └── time.go
        │   ├── uuid -- logic for uuid.
        │   │   ├── db -- logic for db.
        │   │   │   └── uuid.go
        │   │   ├── google_uuid -- logic for google_uuid.
        │   │   │   ├── interface.go
        │   │   │   ├── ulid.go
        │   │   │   └── uuid.go
        │   │   ├── network -- logic for network.
        │   │   │   └── uuid.go
        │   │   ├── uuid.go
        │   │   └── uuid_test.go
        │   └── version
        ├── errors -- Custom errors.
        │   ├── errors.go
        │   ├── go.mod
        │   ├── go.sum
        │   └── version
        ├── logging -- Logging on slog with Ctx and other functional.
        │   ├── alias.go
        │   ├── context.go
        │   ├── go.mod
        │   ├── go.sum
        │   ├── logger.go
        │   ├── logger_test.go
        │   ├── middleware.go
        │   └── version
        ├── main.go
        ├── metrics -- Metrics on prometheus with Ctx and other functional.
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
        ├── pprof -- Profiling on pprof with Ctx and other functional.
        │   ├── config.go
        │   ├── go.mod
        │   ├── server.go
        │   ├── server_test.go
        │   └── version
        ├── redis -- Redis on redis with Ctx and other functional.
        │   ├── aliases.go
        │   ├── error.go
        │   ├── go.mod
        │   ├── go.sum
        │   ├── metrics.go
        │   ├── redis.go
        │   └── version
        └── tracing -- Tracing on jaeger with Ctx and other functional.
            ├── attrs.go
            ├── go.mod
            ├── go.sum
            ├── middleware.go
            ├── tracing.go
            ├── tracing_config.go
            └── version
```


This README now includes a more structured description, project features, and an improved file tree display in markdown format. If you need further refinements or explanations, let me know! 🚀
