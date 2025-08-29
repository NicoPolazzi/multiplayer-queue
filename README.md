# Multiplayer Queue

[![Go CI on Linux](https://github.com/NicoPolazzi/multiplayer-queue/actions/workflows/ci.yml/badge.svg)](https://github.com/NicoPolazzi/multiplayer-queue/actions/workflows/ci.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/NicoPolazzi/multiplayer-queue)](https://goreportcard.com/report/github.com/NicoPolazzi/multiplayer-queue) [![codecov](https://codecov.io/github/nicopolazzi/multiplayer-queue/graph/badge.svg?token=QHJ6HBD4AG)](https://codecov.io/github/nicopolazzi/multiplayer-queue)

## Introduction

A matchmaking and lobby system for multiplayer games, built as a university project to explore distributed systems concepts. This project utilizes the [gRPC](https://grpc.io/) framework for building RPC services and features a web server built with the [Gin](https://github.com/gin-gonic/gin) framework to handle HTTP requests.

## Architecture

The system is designed with a microservices architecture, composed of several components:

1. Auth Service (gRPC): Manages all user authentication tasks;

2. Lobby Service (gRPC): Handles the creation of game lobbies and the matchmaking queue;

3. Web Server (Gin): A lightweight HTTP server built using the Gin framework. It serves the frontend application and exposes the RESTful API endpoints;

4. gRPC Gateway: Acts as a reverse proxy, translating RESTful JSON API calls from the client into gRPC messages for the backend services.


## Requirements

The application requires [Go](https://go.dev/) version [1.24](https://go.dev/doc/devel/release#go1.24.0) or above and [Make](https://www.gnu.org/software/make/).

## Instructions

### Cloning the repository

```bash
git clone https://github.com/NicoPolazzi/multiplayer-queue.git && cd multiplayer-queue 
```

### Build the application

Downloads dependencies and compiles the source code:

```bash
make build
```

### Run the application

Starts the gRPC services, the gateway, and the Gin web server:

```bash
make run
```

### Access the application

Once the servers are running, you can interact with the application. The web app's default index page is available at http://localhost:8080.


## Configuration

The application is configured using environment variables. If you want to use custom values, you can create a .env file in the project's root directory.

Here is an example configuration file:

```
# Set to "release" for production or "debug" for development
GIN_MODE=debug

HOST=<YOUR_HOST>
GRPC_SERVER_PORT=<YOUR_GRPC_SERVER_PORT>
GRPC_GATEWAY_PORT=<YOUR_GATEWAY_PORT>
GIN_SERVER_PORT=<YOUR_GIN_PORT>

DB_DSN=<YOUR_DATABASE_NAME>

JWT_SECRET=<YOUR_SECRET>
```

## Test suite

To run the entire test suite and generate a code coverage report, use the following command:

```bash
make test
```

## License

MIT
