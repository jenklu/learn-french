# Learn French - Listening Practice App

A complete French listening practice application with Go backend and React frontend. Features AI-generated quizzes using Mistral API, progress tracking, push notifications, and admin management.

## Features

### Backend (Go)
- **Authentication**: Secure user registration/login with bcrypt password hashing
- **Database**: PostgreSQL with automatic migrations
- **AI Integration**: Mistral API for generating French listening exercises
- **Progress Tracking**: Track user sessions, scores, and improvement over time
- **Push Notifications**: VAPID-based browser push notifications
- **Reminders**: Timezone-aware daily reminders (10AM and 9PM)
- **Admin Dashboard**: Full user CRUD operations
- **Rate Limiting**: Protect against abuse

### Frontend (React)
- **Interactive Practice**: Listen to French audio (text-based) and answer questions
- **Session History**: View past practice sessions with detailed results
- **Progress Charts**: Visualize your improvement over time
- **Settings**: Configure timezone and notification preferences
- **Admin Panel**: Manage users and content
- **PWA Support**: Installable as a Progressive Web App

## Architecture

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Frontend   │───▶│   Backend   │───▶│  PostgreSQL  │
│   (React)    │    │    (Go)     │    │  Database    │
└─────────────┘    └─────────────┘    └─────────────┘
                          │
                          ▼
                    ┌─────────────┐
                    │ Mistral API │
                    │ (AI Quizzes)│
                    └─────────────┘
```

## Quick Start

### Prerequisites
- Go 1.21+
- Node.js 18+
- PostgreSQL
- Mistral API key (for AI-generated quizzes)

### Environment Variables

Create a `.env` file in the project root:

```bash
# Database
DATABASE_URL=postgres://user:password@localhost:5432/learn_french?sslmode=disable

# Server
PORT=8080
SESSION_SECRET=your-super-secret-session-key
CSRF_SECRET=your-csrf-secret-key

# Mistral AI (for quiz generation)
MISTRAL_API_KEY=your-mistral-api-key

# Push Notifications (VAPID)
VAPID_PUBLIC_KEY=your-vapid-public-key
VAPID_PRIVATE_KEY=your-vapid-private-key

# Frontend
FRONTEND_URL=http://localhost:5173
```

### Run the Application

#### Backend
```bash
# Install dependencies
go mod download

# Run the server
go run ./cmd/server
```

#### Frontend
```bash
cd frontend
npm install
npm run dev
```

The backend will be available at `http://localhost:8080` and the frontend at `http://localhost:5173`.

## Docker Deployment

### Build and Run with Docker
```bash
docker build -t learn-french .
docker run -p 8080:8080 \
  -e DATABASE_URL=postgres://user:password@db:5432/learn_french \
  -e MISTRAL_API_KEY=your-mistral-api-key \
  -e SESSION_SECRET=your-secret \
  -e CSRF_SECRET=your-csrf-secret \
  learn-french
```

### Docker Compose (with PostgreSQL)
```yaml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://user:password@db:5432/learn_french?sslmode=disable
      - MISTRAL_API_KEY=${MISTRAL_API_KEY}
      - SESSION_SECRET=${SESSION_SECRET}
      - CSRF_SECRET=${CSRF_SECRET}
    depends_on:
      - db

  db:
    image: postgres:15
    environment:
      - POSTGRES_USER=user
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=learn_french
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

volumes:
  postgres_data:
```

Run with:
```bash
docker-compose up -d
```

## DigitalOcean App Platform

1. Create a new app in DigitalOcean App Platform
2. Connect your GitHub repository
3. Set the following environment variables:
   - `DATABASE_URL`: Your DigitalOcean PostgreSQL connection string
   - `MISTRAL_API_KEY`: Your Mistral API key
   - `SESSION_SECRET`: A random secret key
   - `CSRF_SECRET`: A random CSRF secret
   - `VAPID_PUBLIC_KEY`: Your VAPID public key
   - `VAPID_PRIVATE_KEY`: Your VAPID private key
4. Deploy!

## API Endpoints

### Authentication
- `POST /api/auth/register` - Register a new user
- `POST /api/auth/login` - Login
- `POST /api/auth/logout` - Logout
- `GET /api/auth/me` - Get current user

### Exercises
- `GET /api/exercises?level=A2&count=5` - Get exercises for a level
- `POST /api/exercises/submit` - Submit exercise answer
- `POST /api/sessions/complete` - Complete a practice session

### Sessions & Progress
- `GET /api/sessions` - Get user's practice sessions
- `GET /api/sessions/results?session_id=1` - Get results for a session
- `GET /api/progress` - Get user's progress
- `GET /api/stats/daily` - Get daily statistics

### Reminders & Notifications
- `GET /api/reminders` - Get reminder settings
- `POST /api/reminders` - Update reminder settings
- `POST /api/push/subscribe` - Subscribe to push notifications
- `GET /api/push/subscriptions` - Get push subscriptions
- `DELETE /api/push/subscriptions` - Delete push subscription

### Admin (Requires Admin Privileges)
- `GET /api/admin/users` - List all users
- `POST /api/admin/users` - Create a user
- `PUT /api/admin/users` - Update a user
- `DELETE /api/admin/users` - Delete a user
- `PUT /api/admin/users/password` - Update user password
- `POST /api/admin/ai/exercises/generate` - Generate new exercises
- `POST /api/admin/ai/exercises/regenerate` - Regenerate exercises
- `GET /api/admin/ai/status` - Check AI service status

## AI Quiz Generation

The app uses Mistral API to generate French listening exercises on-demand. When a user requests exercises:

1. The backend first checks the database for existing exercises
2. If there aren't enough, it calls Mistral API to generate new ones
3. Generated exercises are saved to the database for future use
4. Exercises are served to the user

### Prompt Engineering

The AI uses a carefully crafted prompt to generate high-quality exercises:
- Exactly 4 multiple-choice options
- One correct answer
- Appropriate difficulty for the level (A2 or B1)
- Common French vocabulary and grammar
- Includes both the question and transcript text

### Fallback

If Mistral API is not configured or fails, the app falls back to hardcoded exercises to ensure functionality.

## Database Schema

### Tables
- `users`: User accounts with timezone and admin status
- `sessions`: Practice sessions with duration and scores
- `exercise_results`: Individual exercise results
- `progress`: User progress tracking
- `reminders`: Daily reminder settings
- `push_subscriptions`: Browser push notification subscriptions
- `exercises`: AI-generated and hardcoded exercises

## Configuration

### VAPID Keys for Push Notifications

Generate VAPID keys for push notifications:
```bash
# Using web-push-cli
npx web-push generate-vapid-keys

# Or manually
openssl ecparam -name prime256v1 -genkey -noout -out vapid_private.pem
openssl ec -in vapid_private.pem -pubout -out vapid_public.pem
```

Then set:
```bash
VAPID_PUBLIC_KEY=$(cat vapid_public.pem | base64 -w 0)
VAPID_PRIVATE_KEY=$(cat vapid_private.pem | base64 -w 0)
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT
