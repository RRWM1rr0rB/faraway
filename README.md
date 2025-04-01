# Word of Window

## Test task for Server Engineer(Roman Zaitsau)

### 📜 Description
Design and implement “Word of Wisdom” tcp server.

• TCP server should be protected from DDOS attacks with the Prof of Work (https://en.wikipedia.org/wiki/Proof_of_work), the challenge-response protocol should be used.

• The choice of the POW algorithm should be explained.

• After Prof Of Work verification, server should send one of the quotes from “word of wisdom” book or any other collection of the quotes.

• Docker file should be provided both for the server and for the client that solves the POW challenge.

### ✨ Features:

🔐 PoW protection: Prevents brute-force and bot-based DDoS attacks.

⚡ SHA-256 based challenge: Adjustable difficulty depending on server load.

📖 Quote delivery: Once PoW is verified, the server sends a quote from the "Word of Wisdom" collection.

🐳 Docker support: Dockerized setup for both server and client.

### 🛡️ Proof of Work (PoW) Algorithm Choice

When selecting a PoW algorithm, the main consideration is the type of attackers we aim to defend against. Typically, small hacker groups, competitors, or individual attackers cannot afford high-cost servers. Even if they do, they need to migrate quickly from one server to another, making an efficient and adaptable PoW crucial.

❓ Why SHA-256?

✔️ Security & Performance: Unlike SHA-1, which is outdated and insecure, SHA-256 offers strong cryptographic security.

✔️ Avoiding Self-DDOS: Algorithms like Scrypt and Argon2, while effective against bots, are too resource-intensive and could overload our own server.

✔️ Dynamic Difficulty Adjustment:

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
README.md
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