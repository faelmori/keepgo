# KeepGo

![Version](https://img.shields.io/github/v/release/faelmori/keepgo)
![Build Status](https://img.shields.io/github/actions/workflow/status/faelmori/keepgo/build.yml?branch=main)
![License](https://img.shields.io/github/license/faelmori/keepgo)

A cross-platform Go library for installing, managing, and running services (daemons) on multiple operating systems. KeepGo currently supports:

- **Windows** (XP and later)
- **Linux** (Systemd, Upstart, SysV)
- **macOS** (Launchd)

> **Note:** Please report any issues in the [main repository](https://github.com/faelmori/keepgo).

---

## Table of Contents

- [Features](#features)
- [Installation](#installation)
    - [Supported Platforms](#supported-platforms)
    - [1. Go Installation](#1-go-installation)
    - [2. Build from Source](#2-build-from-source)
- [Usage](#usage)
    - [Basic Example](#basic-example)
- [KeepGo vs Other Libraries](#keepgo-vs-other-libraries)
- [Considerations](#considerations)
- [Contributing](#contributing)
- [License](#license)
- [Acknowledgments](#acknowledgments)

---

## Features

- **Easy Service Management:** Install and uninstall services with minimal effort.
- **Cross-Platform API:** Unified API for different operating systems.
- **Programmatic Control:** Start, stop, and restart services in code.
- **Interactive Mode Detection:** Identify if running in a terminal or service manager.

---

## Installation

### Supported Platforms

- **Windows**
- **Linux**
- **macOS**

### 1. Go Installation

Install KeepGo using Go modules:

```shell
go get github.com/faelmori/keepgo
```

### 2. Build from Source

#### Requirements

- [Git](https://git-scm.com/downloads)
- [Go](https://go.dev/doc/install)

#### Steps

1. **Clone the Repository:**
   ```shell
   git clone https://github.com/faelmori/keepgo.git
   ```

2. **Navigate to the Project Directory:**
   ```shell
   cd keepgo
   ```

3. **Build the Library:**
   ```shell
   go build -o keepgo
   ```

4. **Verify Installation:**
   ```shell
   go test ./...
   ```

---

## Usage

### Basic Example

Here’s a simple example of how to register a service with KeepGo:

```go
package main

import (
	"fmt"
	"github.com/faelmori/keepgo"
)

type program struct{}

func (p *program) Start(s keepgo.Service) error {
	go p.run()
	return nil
}

func (p *program) run() {
	fmt.Println("Service running...")
}

func (p *program) Stop(s keepgo.Service) error {
	fmt.Println("Service stopped.")
	return nil
}

func main() {
	svcConfig := &keepgo.Config{Name: "MyService"}
	prg := &program{}
	srv, err := keepgo.New(prg, svcConfig)
	if err != nil {
		fmt.Println("Error creating service:", err)
		return
	}
	srv.Run()
}
```

---

## KeepGo vs Other Libraries

KeepGo is a fork of [kardianos/service](https://github.com/kardianos/service) with improvements in:

- **Usability:** More intuitive API for developers.
- **Documentation:** Clearer guides and examples.
- **Efficiency:** Lightweight and optimized for Go service execution.

---

## Considerations

- **Linux:** Dependency management is not fully supported.
- **macOS:** `UserService Interactive` mode may have limitations.

---

## Contributing

We welcome contributions from the community! Feel free to open **issues** or submit **pull requests** to enhance KeepGo.

---

## License

KeepGo is licensed under the [MIT License](LICENSE).

---

## Acknowledgments

KeepGo is an independent project based on the original work of [kardianos/service]. We appreciate all contributors to the original project!

---

Thank you for using **KeepGo**! 🚀 If you have suggestions or find issues, please open an issue in the [main repository](https://github.com/faelmori/keepgo).

