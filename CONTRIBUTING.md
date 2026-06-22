# Contributing to dbspin

First off, thank you for considering contributing to `dbspin`! It's people like you who make open-source software a great ecosystem.

We want to make contributing to this project as easy and safe as possible, whether it's:

- Reporting a bug
- Discussing the current state of the code
- Submitting a fix
- Proposing new features

---

## 🛠️ Local Development Setup

To work on this project locally, make sure you have:
1. **Go** (1.21+ recommended) installed.
2. **Git** configured.
3. **Docker** installed and running locally.

### Fork and Clone the Repository
1. Fork the repo on GitHub.
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/dbspin.git
   cd dbspin
   ```

### Running Locally
To run and test the program without compiling:
```bash
go run main.go up postgres
```

### Running Tests
Make sure the tests pass before committing:
```bash
go test -v ./...
```

---

## 🎨 Code Style and Quality

To keep the codebase clean and consistent:
- Run `go fmt ./...` before committing.
- Run `go vet ./...` to check for common compiler/runtime errors.
- We recommend installing and registering `pre-commit` hooks so formatting checks run automatically on commit.

---

## 📥 Submitting a Pull Request

1. Create a new branch for your feature or fix:
   ```bash
   git checkout -b feature/AmazingFeature
   ```
2. Commit your changes with clear, descriptive commit messages.
3. Push to your branch:
   ```bash
   git push origin feature/AmazingFeature
   ```
4. Open a Pull Request against our `main` branch. Provide a description of what your PR changes and why it's needed.

---

## 📄 License

By contributing, you agree that your contributions will be licensed under its **MIT License**.
