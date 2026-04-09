#!/bin/bash
# Esbio Startup Script
# Starts the database, backend server, and frontend client

echo "🚀 Starting Esbio..."
echo ""

# Get the script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Check if Docker daemon is running, start it if not
echo "🐳 Checking Docker..."
if ! docker info > /dev/null 2>&1; then
    echo "🔄 Docker is not running. Starting Docker Desktop..."
    open -a Docker
    # Wait for Docker daemon to be ready
    echo "⏳ Waiting for Docker to start..."
    while ! docker info > /dev/null 2>&1; do
        sleep 2
    done
    echo "✅ Docker is running"
else
    echo "✅ Docker is already running"
fi
echo ""

# Fix docker-credential-desktop not found in PATH (common Docker Desktop issue)
if [ -f "$HOME/.docker/config.json" ]; then
    if grep -q '"credsStore": "desktop"' "$HOME/.docker/config.json" && \
       ! command -v docker-credential-desktop > /dev/null 2>&1; then
        echo "⚠️  Fixing Docker credential helper (docker-credential-desktop not in PATH)..."
        # Temporarily remove credsStore so docker compose doesn't fail on pull
        sed -i.bak 's/"credsStore": "desktop"/"credsStore": ""/g' "$HOME/.docker/config.json"
        DOCKER_CREDS_FIXED=true
    fi
fi

# Check if PostgreSQL container is running
echo "📦 Checking PostgreSQL container..."
if docker ps | grep -q esbio-postgres; then
    echo "✅ PostgreSQL container is already running"
else
    echo "🔄 Starting PostgreSQL container..."
    cd "$SCRIPT_DIR/server" && docker compose up -d
    echo "⏳ Waiting for PostgreSQL to be ready..."
    # Wait for healthcheck to pass
    until docker exec esbio-postgres pg_isready -U postgres > /dev/null 2>&1; do
        sleep 2
    done
    echo "✅ PostgreSQL is ready"
fi

# Restore docker config if we modified it
if [ "$DOCKER_CREDS_FIXED" = true ] && [ -f "$HOME/.docker/config.json.bak" ]; then
    mv "$HOME/.docker/config.json.bak" "$HOME/.docker/config.json"
fi
echo ""

# Start the backend server
echo "🔧 Starting Go backend server on :8080..."
cd "$SCRIPT_DIR/server/cmd"
go run api/main.go &
BACKEND_PID=$!
echo "✅ Backend started (PID: $BACKEND_PID)"
echo ""

# Wait a moment for backend to initialize
sleep 2

# Start Stripe webhook listener (if stripe CLI is installed)
if command -v stripe > /dev/null 2>&1; then
    echo "💳 Starting Stripe webhook listener..."
    stripe listen --forward-to localhost:8080/api/v1/stripe/webhook &
    STRIPE_PID=$!
    echo "✅ Stripe webhook listener started (PID: $STRIPE_PID)"
else
    echo "⚠️  Stripe CLI not installed — webhook listener skipped"
    echo "   Install with: brew install stripe/stripe-cli/stripe"
    STRIPE_PID=""
fi
echo ""

# Start the frontend client
echo "⚛️  Starting Next.js frontend on :3000..."
cd "$SCRIPT_DIR/client"
npm run dev &
FRONTEND_PID=$!
echo "✅ Frontend started (PID: $FRONTEND_PID)"
echo ""

# Save PIDs for the stop script
echo "$BACKEND_PID" > "$SCRIPT_DIR/.backend.pid"
echo "$FRONTEND_PID" > "$SCRIPT_DIR/.frontend.pid"
[ -n "$STRIPE_PID" ] && echo "$STRIPE_PID" > "$SCRIPT_DIR/.stripe.pid"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✨ Esbio is now running!"
echo ""
echo "📱 Frontend:  http://localhost:3000"
echo "🔌 Backend:   http://localhost:8080"
echo "🗄️  Database:  postgres://localhost:5432"
[ -n "$STRIPE_PID" ] && echo "💳 Stripe:    webhook listener active"
echo ""
echo "To stop all services, run: ./stop.sh"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Keep script running and show logs
echo "📋 Showing logs (Ctrl+C to exit logs, services will keep running)..."
echo ""
wait
