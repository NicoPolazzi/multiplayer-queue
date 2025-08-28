# Multiplayer Queue

[![Go CI on Linux](https://github.com/NicoPolazzi/multiplayer-queue/actions/workflows/ci.yml/badge.svg)](https://github.com/NicoPolazzi/multiplayer-queue/actions/workflows/ci.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/NicoPolazzi/multiplayer-queue)](https://goreportcard.com/report/github.com/NicoPolazzi/multiplayer-queue) [![codecov](https://codecov.io/github/nicopolazzi/multiplayer-queue/graph/badge.svg?token=QHJ6HBD4AG)](https://codecov.io/github/nicopolazzi/multiplayer-queue)

## Introduction

TODO: insert description of the project

## Requirements

The application requires [Go](https://go.dev/) version [1.24](https://go.dev/doc/devel/release#go1.24.0) or above.  


## Instructions

### Cloning the repository

Clone this repository locally and enter the cloned folder:

```bash
git clone https://github.com/NicoPolazzi/multiplayer-queue.git && cd multiplayer-queue 
```

### Build the application

Download the required dependencies and build the application:

```bash
make build
```

### Run the application

To start the application, you need to run:

```bash
make run
```

After the servers are running, you can start using the web app visiting the default [index page](http://localhost:8080).



## Configuration

TODO: suggestion for a possible .env file and explain that the app uses default values 

## Test suite

You can run the full test suite, that includes also the code coverage computation, with:

```bash
make test
```

## License

MIT
