# Network Access Setup

How to access Eskio from other devices on the same local network.

## What was changed

### 1. Next.js allowed dev origins (`client/next.config.ts`)

Next.js 16 blocks cross-origin requests in development by default. When accessing the app from another device (e.g. `192.168.0.38:3000` instead of `localhost:3000`), Next.js treats it as a cross-origin request and blocks it.

Added `allowedDevOrigins` to allow requests from the local network IP:

```ts
const nextConfig: NextConfig = {
  allowedDevOrigins: ["192.168.0.38"],
};
```

### 2. API URL (`client/.env.local`)

The frontend makes API calls to the backend. This URL is set via the `NEXT_PUBLIC_API_URL` environment variable.

Previously it pointed to `localhost:8080`, which only works when the browser is on the same machine as the server. A browser on another device would try to reach its own `localhost`, not the server.

Changed to the network IP:

```
NEXT_PUBLIC_API_URL=http://192.168.0.38:8080/api/v1
```

### 3. Backend CORS (`server/cmd/api/internal/middleware/cors_middleware.go`)

The Go backend had a hardcoded CORS `Access-Control-Allow-Origin` header set to `http://localhost:3000`. Browsers enforce CORS by checking that the server's `Access-Control-Allow-Origin` matches the origin of the request.

When the browser on the other device sends a request from `http://192.168.0.38:3000`, the backend responded with `http://localhost:3000` as the allowed origin. The browser saw a mismatch and blocked the request.

Updated the middleware to accept both origins:

```go
allowedOrigins := map[string]bool{
    "http://localhost:3000":     true,
    "http://192.168.0.38:3000": true,
}
```

## Note

If your local IP changes (e.g. after reconnecting to WiFi), you'll need to update:

1. `client/.env.local` - the API URL
2. `client/next.config.ts` - the allowed dev origin
3. `server/cmd/api/internal/middleware/cors_middleware.go` - the CORS allowed origin
