# rayChats 🚀

A real-time chat application built with Go, Gin framework, and modern web technologies. rayChats provides a seamless messaging experience with room-based conversations and user authentication.

## 🌟 Features

- **Real-time messaging** - Instant communication using WebSockets/gRPC
- **Room-based chat** - Join and create chat rooms
- **User authentication** - Secure login system with Google OAuth integration
- **OTP verification** - Additional security layer for user verification  
- **Scalable architecture** - Built with Go's concurrent programming model

## 🛠️ Tech Stack

- **Backend**: Go (Golang) with Gin web framework
- **Authentication**: Google OAuth, OTP verification
- **Database**: Valkey, Postgresql
- **Communication**: gRPC for high-performance service communication
- **Frontend**: HTML, CSS, JavaScript (served from static files)

## 📁 Project Structure

```
rayChats/
├── config/             # Configuration files and settings
├── database/           # Database models, migrations, and connections
├── external_services/  # Third-party service integrations (Google OAuth, etc.)
├── handler/            # HTTP request handlers and middleware
├── models/             # Data models and structs
├── proto/              # Protocol Buffer definitions for gRPC
├── services/           # Business logic and service layer
├── static/             # Static assets (CSS, JS, images)
├── .gitignore         # Git ignore file
├── go.mod             # Go module dependencies
├── go.sum             # Go module checksums
└── main.go            # Application entry point
```

## 🚀 Getting Started

### Prerequisites

- Go 1.19 or higher
- Git

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/ray-27/rayChats.git
   cd rayChats
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up configuration**
   - Copy and configure your environment variables
   - Update database connection settings in `config/`
   - Set up Google OAuth credentials

4. **Run the application**
   ```bash
   go run main.go
   ```

5. **Access the application**
   - Open your browser and navigate to `http://localhost:8080` (or your configured port)

## ⚙️ Configuration

Configure the following environment variables:

```bash
GIN_MODE=debug
PORT=3000
GRPC_PORT=50051

GMAIL_USER=
GMAIL_APP_PASSWORD=

PG_HOST=
PG_PORT=
PG_USER=
PG_PASSWORD=
PG_DBNAME=

VALKEY_ENDPOINT=localhost:6379
VALKEY_PASSWORD=
```

## 🎯 Usage

1. **Sign up/Login**: Create an account or login using Google OAuth
2. **Verify Account**: Complete OTP verification if required
3. **Join Rooms**: Browse and join existing chat rooms
4. **Start Chatting**: Send and receive messages in real-time
5. **Create Rooms**: Create new chat rooms for specific topics

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the Creative Commons Non-Commercial (CC BY-NC)

## 👨‍💻 Author

**Ray** - [@ray-27](https://github.com/ray-27)

## 🙏 Acknowledgments

- Gin framework for the excellent HTTP web framework
- Google for OAuth integration
- The Go community for amazing tools and libraries

---

⭐ Star this repository if you find it helpful!
