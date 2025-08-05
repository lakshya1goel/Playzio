# Playzio 🎮

A real-time multiplayer word game platform built with Go, featuring WebSocket-based gameplay, chat functionality, and Google OAuth authentication.

## 📱 Frontend Application

The React Native mobile app for this backend is available at: [Playzio-app](https://github.com/lakshya1goel/Playzio-app)

## Features

- 🎯 **Real-time Multiplayer Gameplay**: Turn-based word game with up to 10 players per room
- 💬 **Live Chat**: Real-time chat functionality with Redis support
- 🔐 **Google OAuth Authentication**: Secure user authentication via Google
- 🏠 **Room Management**: Create and join game rooms
- ⏱️ **Timer System**: Configurable time limits for turns (5-20 seconds)
- 🏆 **Scoring System**: Lives and points tracking
- 🔄 **WebSocket Communication**: Real-time bidirectional communication
- 📦 **Dockerized**: Easy deployment with Docker Compose

## Tech Stack

- **Backend**: Go (Gin framework)
- **Database**: PostgreSQL with GORM
- **Cache**: Redis
- **WebSockets**: Gorilla WebSocket
- **Authentication**: Google OAuth 2.0
- **Containerization**: Docker & Docker Compose

## Project Structure

```
Playzio/
├── api/                    # API controllers and routes
│   ├── controller/         # HTTP and WebSocket controllers
│   ├── middleware/         # Authentication middleware
│   └── routes/            # Route definitions
├── bootstrap/             # Application initialization
│   ├── database/          # Database configuration
│   ├── redis/             # Redis configuration
│   └── util/              # Utility functions
├── cmd/                   # Application entry point
├── domain/                # Domain models and DTOs
│   ├── dto/               # Data Transfer Objects
│   └── model/             # Database models
├── repository/            # Data access layer
├── usecase/               # Business logic layer
├── websocket/             # WebSocket handlers and game logic
└── docker-compose.yaml    # Docker services configuration
```

## Prerequisites

- Go 1.24.2 or higher
- PostgreSQL 15+
- Redis 7.2+
- Docker & Docker Compose (optional)
- Google OAuth 2.0 credentials

## Environment Setup

### 1. Clone the Repository

```bash
git clone https://github.com/lakshya1goel/Playzio.git
cd Playzio
```

### 2. Environment Variables

Create a `.env` file in the root directory with the following variables:

```env
# Database Configuration
DB_HOST=your_db_host
DB_PORT=your_db_port
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_NAME=your_db_name

# Redis Configuration
REDIS_HOST=your_redis_host
REDIS_PORT=your_redis_port
REDIS_PASSWORD=your_redis_password
REDIS_DB=your_redis_db_number

# Google OAuth Configuration
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
GOOGLE_REDIRECT_URI=your_google_redirect_uri
```

### Example `.env` for Local Development

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=playzio

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Google OAuth Configuration
GOOGLE_CLIENT_ID=your_actual_google_client_id
GOOGLE_CLIENT_SECRET=your_actual_google_client_secret
GOOGLE_REDIRECT_URI=http://localhost:8000/auth/google/callback
```

### 3. Google OAuth Setup

For detailed Google OAuth setup instructions, follow this guide: [Google OAuth Implementation using React Native and Golang](https://medium.com/@lakshya1234goel/google-oauth-implementation-using-react-native-and-golang-8782752491a7)

## Installation & Running

### Using Docker Compose (Recommended)

```bash
# Start PostgreSQL and Redis services
docker-compose up -d

# Install Go dependencies
go mod tidy

# Run the application
go run cmd/main.go
```

## Game Rules

1. **Room Creation**: Any authenticated user can create a game room
2. **Joining**: Players can join rooms with available slots (max 10 players)
3. **Game Start**: Game begins with a 2-minute countdown when the first player joins
4. **Turns**: Players take turns providing answers within the time limit (5-20 seconds)
5. **Lives**: Each player starts with 3 lives
6. **Scoring**: Points are awarded for correct answers
7. **Game End**: Game ends when only one player remains or all players are eliminated

## Development

### Project Architecture

The project follows a clean architecture pattern:

- **API Layer**: HTTP handlers and WebSocket controllers
- **Use Case Layer**: Business logic implementation
- **Repository Layer**: Data access abstraction
- **Domain Layer**: Core business entities and models

### Adding New Features

1. Define models in `domain/model/`
2. Create DTOs in `domain/dto/`
3. Implement repository interfaces in `repository/`
4. Add business logic in `usecase/`
5. Create controllers in `api/controller/`
6. Define routes in `api/routes/`

## Deployment

1. **Production Environment**:
   - Update `.env` with production values
   - Use a production-ready PostgreSQL instance
   - Configure Redis for production
   - Set up proper Google OAuth redirect URIs

2. **Docker Deployment**:
   ```bash
   # Build and run with Docker Compose
   docker-compose up --build -d
   ```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Contact

**Lakshya Goel**
- Email: lakshya1234goel@gmail.com
- GitHub: [@lakshya1goel](https://github.com/lakshya1goel)

## Support

If you encounter any issues or have questions, please open an issue on GitHub.

---

**Happy Gaming! 🎮** 