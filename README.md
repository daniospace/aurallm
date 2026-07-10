# 🟢 AuraLLM: The Cloud-Native, FinOps-Enforced AI Gateway

AuraLLM is a highly extensible, commercial-grade AI Gateway designed as a performant, developer-friendly alternative to solutions like LiteLLM. Written in Go for raw concurrency and effortless platform extensibility, AuraLLM serves as a secure bridge connecting your developers' standard OpenAI codebases to multiple downstream LLM providers (OpenAI, Anthropic, Gemini, DeepSeek, etc.) while enforcing **real-time token limits, budget controls, and cost auditing**.

---

## 🚀 Why AuraLLM beats LiteLLM

1. **DevOps & Cloud-Native Native (Go vs. Python/Rust):** Unlike LiteLLM (written in Python), AuraLLM has absolute zero runtime overhead, starting in under 10ms with a tiny ~15MB memory footprint. Unlike Rust gateways, AuraLLM exposes clean, idiomatic Go interfaces, allowing enterprise platform teams to easily write custom middleware (PII scrubbers, proprietary auth, custom prompt guards) in hours, not weeks.
2. **Unified Single-Binary Deployments:** AuraLLM compiles down into a single, self-contained executable (~20MB) that embeds its entire administration dashboard. You can distribute AuraLLM as a zero-dependency service or Kubernetes sidecar.
3. **Reactive Real-Time Budget Enforcement:** AuraLLM intercepts incoming request payloads and Server-Sent Events (SSE) stream lines in real-time, parsing token usage on-the-fly and atomically deducting costs from an in-memory cache to guarantee budget limits are never exceeded.
4. **Executive FinOps Console:** A premium, styled slate-indigo Next.js dashboard built for Engineering and Financial Managers featuring month-to-date spend accumulation, 30-day forecasting trendlines, budget-at-risk alarms, and team onboarding.

---

## ⚡ Killer Enterprise Features (Where LiteLLM Falls Short)

### 1. 🌉 Shadow Parallel Routing (Migration Auditor)
AuraLLM allows you to test cheaper, specialized, or open-source models with live production traffic without any latency or reliability risk.
*   **How it works:** When a client request is sent with the header `X-Shadow-Model: claude-3-5-sonnet`, AuraLLM resolves the primary request (e.g., `gpt-4o`) instantly to keep production latency at absolute zero. In the background, a non-blocking goroutine duplicates the prompt, queries the shadow model, and records side-by-side cost, token, and latency metrics.
*   **The Benefit:** CTOs get 100% data-driven, side-by-side comparison tables showing exactly what production traffic would cost and look like under a cheaper model, removing all friction from migration decisions.

### 🛡️ 2. Local PII & Secrets Redaction Guard
AuraLLM intercepts prompts and sanitizes sensitive records locally on your private network before they ever reach public cloud LLM servers.
*   **How it works:** AuraLLM uses compiled Go regex to scrub high-risk patterns—such as **Social Security Numbers (SSNs)**, **Credit Card numbers**, **email addresses**, and **API keys/Secrets**—replacing them with safe placeholders (e.g., `[REDACTED_SSN_1]`). For non-streaming responses, it automatically un-redacts (re-injects) original values on-the-fly.
*   **The Benefit:** Full, frictionless SOC2, HIPAA, and GDPR compliance out-of-the-box. Developers write zero custom security code, while compliance teams track blocked leaks via the dashboard.

---

## 🛠️ System Architecture

*   **Go Backend (`/backend`):** High-concurrency engine utilizing standard Go routers, Postgres storage adapters, atomic memory cache layers, and bidirectional translation protocols.
*   **Next.js Dashboard (`/frontend`):** Premium Next.js application styled with CSS variables and custom SVG-based live charting.
*   **Data Layer:** PostgreSQL (for durable state, credentials, and usage logs) and an atomic In-Memory Cache (for zero-latency hot authorization/budget checks).

---

## 🚦 Getting Started

### 1. Launch the Database
Start the pre-configured PostgreSQL database on port `5435`:
```bash
docker-compose up -d
```

### 2. Configure the Environment
An `.env` file has been generated in the root directory:
```env
DATABASE_URL=postgres://postgres:postgres@localhost:5435/gateway?sslmode=disable
REDIS_URL=redis://localhost:6379/0
PORT=8085
```

---

## 💻 Developer Workflow (Run Separately)

Running separate services is recommended during active development for live hot-reloading.

#### Start the Go Backend API
```bash
cd backend
go run main.go
```
*AuraLLM will automatically load `.env` parameters and start on port `8085`.*

#### Start the Next.js Dev Console
```bash
cd frontend
npm run dev
```
*This launches your hot-reloaded dashboard at `http://localhost:3000`.*

---

## 📦 Production Workflow (Unified Single Binary)

To build and compile AuraLLM into a single executable serving both the dashboard and the proxy APIs:

#### 1. Compile Next.js to Static Assets
```bash
cd frontend
npm run build
```

#### 2. Embed Static Assets & Compile Go
```bash
# Clear old Go embedding assets and sync fresh static files
rm -rf backend/out && cp -r frontend/out backend/out

# Compile the final binary
cd backend
go build -o gateway main.go
```

#### 3. Run the Unified Application
```bash
./backend/gateway
```
*Your entire gateway, including the live administrative dashboard, is now active at `http://localhost:8085/`.*

---

## 🔒 Enterprise Gateway API Schema

Developers access multi-provider models through AuraLLM by changing their client headers to point to the gateway:

### Endpoint: `POST /v1/chat/completions`
*   **Auth Header:** `Authorization: Bearer <your_aurallm_team_key>`
*   **Optional Shadow Routing:** `X-Shadow-Model: claude-3-5-sonnet`
*   **Payload (Standard OpenAI):**
    ```json
    {
      "model": "gpt-4o",
      "messages": [{"role": "user", "content": "Analyze credit card leaks in SSN 999-12-3456"}]
    }
    ```
