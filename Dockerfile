# Build stage
FROM node:20-alpine AS frontend-builder

WORKDIR /app

# Copy frontend files
COPY frontend/package.json frontend/package-lock.json* frontend/tsconfig.json frontend/tsconfig.node.json frontend/vite.config.ts ./

# Install dependencies
RUN npm ci

# Copy source files
COPY frontend/src/ ./src/
COPY frontend/public/ ./public/
COPY frontend/index.html ./

# Build the frontend
RUN npm run build

# Production stage
FROM golang:1.21-alpine AS backend-builder

WORKDIR /app

# Copy Go files
COPY go.mod go.sum ./
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /learn-french ./cmd/server

# Final stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Copy built files
COPY --from=frontend-builder /app/dist ./frontend/dist
COPY --from=backend-builder /learn-french ./

# Copy static files
COPY frontend/dist/index.html ./frontend/dist/index.html
COPY frontend/public/sw.js ./frontend/dist/sw.js

# Create directory for static files
RUN mkdir -p /app/frontend/dist

# Set timezone
ENV TZ=UTC

# Expose port
EXPOSE 8080

# Set environment variables (will be overridden by deployment)
ENV DATABASE_URL=postgres://user:password@localhost:5432/learn_french?sslmode=disable
ENV PORT=8080
ENV SESSION_SECRET=super-secret-key-change-in-production
ENV CSRF_SECRET=csrf-secret-key-change-in-production
ENV VAPID_PUBLIC_KEY=""
ENV VAPID_PRIVATE_KEY=""
ENV TTS_API_KEY=""
ENV FRONTEND_URL=http://localhost:8080

# Run the application
CMD ["./learn-french"]
