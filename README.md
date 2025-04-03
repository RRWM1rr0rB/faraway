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

## Feature ...

### 🔥 Enhanced PoW System

#### 1. Sessions for Legitimate Users

• If an IP is not flagged for suspicious activity, then:

• After successfully solving one PoW challenge, they receive 15 minutes of free access.

• This reduces the computational burden on users.

• Sessions are stored in Redis.

#### 2. Dynamic PoW Difficulty

• PoW difficulty adjusts dynamically:

• Under normal traffic, users solve simple challenges (5-10 leading zeros).

• If an IP's activity increases, the difficulty scales up (15-30 leading zeros).

• If the server detects anomalously fast solutions, the IP is banned.

##### 3. Blacklist System

• IPs solving difficult challenges too quickly are added to a blacklist.

• No full blocking: these IPs receive the hardest PoW challenges (30-32 leading zeros), making an attack prohibitively expensive.

#### 4. State Tracking in Redis

• We use Redis to track user activity:

• Key: pow_session:<IP>

• Value: timestamp + last PoW difficulty

• If the session is active, no new PoW is required for 15 minutes.

• ### 🎯 Conclusion

✅ Legitimate users are not affected – 1 PoW every 15 minutes.

✅ Bots must solve PoW constantly, making an attack unprofitable.

✅ The system adapts flexibly to different loads.

• By implementing this system, we achieve a balance between security and usability! 🚀

### 📌 Future Enhancements

• Currently, only a basic PoW system has been implemented. The full adaptive PoW system, as described above, would require approximately one week to develop and integrate.



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

-----------------------------------------------------------------------------------------

## Architecture Choice

- **Client-Server**: A client-server architecture that separates logic between client and server code.

- **Layers**: Layers such as controller and policy interact through interfaces and structures. As the logic scales, this allows for efficient functional expansion. Additionally, a domain folder can be introduced for core logic, with a service handling interactions and database selection, while storage is responsible for implementing database interaction methods.

goos: darwin
goarch: arm64
pkg: app-client/app/internal/policy/mitigator
cpu: Apple M2

|           description           | iterations | nanoseconds per operaion | 
|:-------------------------------:|:-----------|:-------------------------|  
| BenchmarkSolveChallenge_5zeros  | 677421     | 2262 ns/op               | 
| BenchmarkSolveChallenge_10zeros | 155305     | 11870 ns/op              | 
| BenchmarkSolveChallenge_15zeros | 967        | 9537857 ns/op            | 
| BenchmarkSolveChallenge_20zeros | 25         | 60938853 ns/op           | 
| BenchmarkSolveChallenge_25zeros | 1          | 10221830500 ns/op        | 
| BenchmarkSolveChallenge_30zeros | 1          | 23460264833 ns/op        |

This README now includes a more structured description, project features, and an improved file tree display in markdown format. If you need further refinements or explanations, let me know! 🚀
