
# Logging & Server Configuration Guide

This guide explains how to configure logging and server settings for different environments using `.env` variables.

---

## Server Mode (`APP_SERVER_MODE`)

Controls **Gin Framework** behavior:

| Value   | Description                                | When to Use     |
|---------|--------------------------------------------|-----------------|
| `debug` | Shows route registration, detailed errors  | Development     |
| `release` | Production optimized, minimal output     | Production      |
| `test`  | For testing, disables console colors       | Testing/CI      |

---

## Logger Level (`APP_LOGGER_LEVEL`)

Controls **what logs are shown**:

| Value   | Shows                                 | When to Use     |
|---------|----------------------------------------|-----------------|
| `debug` | All logs (debug, info, warn, error)    | Development     |
| `info`  | info, warn, error (hides debug)        | Staging         |
| `warn`  | warn, error only                       | Production      |
| `error` | error only                             | Critical Prod   |

---

## Logger Format (`APP_LOGGER_FORMAT`)

Controls **how logs look**:

| Value         | Output                 | When to Use         |
|---------------|------------------------|---------------------|
| `console`     | Colored, human-readable| Development         |
| `json`        | Single-line JSON       | Production          |
| `json-pretty` | Multi-line JSON        | Dev / Debugging     |

---

## Recommended Combinations

### ✅ Development

```env
APP_SERVER_MODE=debug
APP_LOGGER_LEVEL=debug
APP_LOGGER_FORMAT=console
APP_LOGGER_DISABLE_GIN=false
```

### 🧪 Staging

```env
APP_SERVER_MODE=debug
APP_LOGGER_LEVEL=info
APP_LOGGER_FORMAT=json-pretty
APP_LOGGER_DISABLE_GIN=true
```

### 🚀 Production

```env
APP_SERVER_MODE=release
APP_LOGGER_LEVEL=warn
APP_LOGGER_FORMAT=json
APP_LOGGER_DISABLE_GIN=true
```

---

## Log Output Examples

### Debug Level

```
DEBUG: Database query executed
INFO:  User logged in
WARN:  Rate limit approaching
ERROR: Database connection failed
```

### Info Level

```
INFO:  User logged in
WARN:  Rate limit approaching  
ERROR: Database connection failed
```

### Warn Level

```
WARN:  Rate limit approaching
ERROR: Database connection failed
```

### Error Level

```
ERROR: Database connection failed
```

---

## Log Format Output Samples

### 1. Console Format (Development)
Set in `.env`:
```env
APP_LOGGER_FORMAT=console
```

Example output:
```
2025-08-02T00:58:59.392+0530  INFO  server/main.go:53  Starting API Server  {"version": "1.0.0", "environment": "debug"}
```

---

### 2. JSON Format (Production - One Line)
Set in `.env`:
```env
APP_LOGGER_FORMAT=json
```

Example output:
```json
{"level":"info","timestamp":"2025-08-02T00:58:59.392+0530","caller":"server/main.go:53","message":"Starting API Server","version":"1.0.0"}
```

---

### 3. Pretty JSON Format (Development - Multi Line)
Set in `.env`:
```env
APP_LOGGER_FORMAT=json-pretty
```

Example output:
```json
{
  "level": "info",
  "timestamp": "2025-08-02T00:58:59.392+0530",
  "caller": "server/main.go:53",
  "message": "Starting API Server",
  "version": "1.0.0",
  "environment": "debug"
}
```

---

### 4. Disable GIN Debug Logs
Set in `.env`:
```env
APP_LOGGER_DISABLE_GIN=true
```

This will hide logs like:
```
[GIN-debug] GET /health --> handler.Health (6 handlers)
```

---

## Quick Test Commands

### Console format (recommended for development)
```bash
echo "APP_LOGGER_FORMAT=console" >> .env
make dev
```

### Pretty JSON format (for dev debugging)
```bash
sed -i 's/APP_LOGGER_FORMAT=console/APP_LOGGER_FORMAT=json-pretty/' .env
make dev
```

### Regular JSON (for production)
```bash
sed -i 's/APP_LOGGER_FORMAT=json-pretty/APP_LOGGER_FORMAT=json/' .env
make dev
```

### Clean GIN logs
```bash
echo "APP_LOGGER_DISABLE_GIN=true" >> .env
make dev
```

---
