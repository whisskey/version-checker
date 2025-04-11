# Version-Checker

A Go web application that checks the version of a specified package on crates.io using the Fiber framework and Uber FX dependency injection.

## 🛠 Prerequisites

- Go 1.23+

## ⚡️ Quick Start

**Clone the repository**
```bash
git clone https://github.com/whisskey/version-checker.git
```

**Set up environment variables**
```bash
cp .env.example .env
# Edit .env with your configurations
```

**OR**

**Run locally**
```bash
go mod download
go run main.go
```

## 🔄 API Endpoints

### Crate Resource
- `GET /api/crate/:name/:version` - Check if a new version of the specified crate is available

## 🤝 Contributing

- Fork the repository
- Create your feature branch (`git checkout -b feature/new-feature`)
- Commit your changes (`git commit -m 'feat: add new feature'`)
- Push to the branch (`git push origin feature/new-feature`)
- Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details. 