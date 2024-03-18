# Go Connection Pool

A lightweight, efficient, and thread-safe connection pool management library in Go. It's designed to maintain a pool of reusable connection objects, reducing the overhead of establishing connections frequently. The library provides mechanisms to obtain and release connections efficiently, handle connection timeouts, and automatically clean up idle connections.

## Features

- **Thread-Safe Connection Handling**: Ensures that connections are safely shared among multiple goroutines.
- **Automatic Connection Cleanup**: Periodically removes idle connections to free up resources.
- **Flexible Connection Creation**: Supports custom connection creation logic.
- **Panic Handling**: Includes mechanisms to deal with panics that may occur during connection use or initialization.
- **Configurable Parameters**: Offers options to configure the pool size, idle timeout, auto-cleanup interval, and more.

![](./png/pool.png)

## Components

The library consists of several key components:

- **Connector Interface**: Defines the methods for a connection handler.
- **AtomicConnector**: An implementation of the `Connector` interface, managing individual connection states atomically.
- **ConnectorSet Interface**: Manages a set of `Connector` objects, providing methods to add, retrieve, and clean up connectors.
- **AutoClearConnectorSet**: An implementation of the `ConnectorSet` interface, adding automatic cleanup capabilities.
- **ConnectPool Interface**: Represents the overall connection pool, offering methods to register new connections, obtain connection statistics, and configure the pool.

## Getting Started

### Installation

To use this library, first ensure you have Go installed on your system. Then, include the library in your project by copying the provided code into your project's directory.

```bash
go get github.com/HuXin0817/ConnectPool
```

### Usage

To create a new connection pool, you need to define a connection creation method and optionally customize the pool with various configuration options. Here's a simple example:

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    // Define a method to create new connections. In this example, it's a dummy function.
    connectMethod := func() any {
        return "new connection" // Replace this with actual connection logic
    }

    // Create a new connection pool with default settings
    pool := NewConnectPool(connectMethod)

    // Register a new connection from the pool
    newConnect, cancelFunc := pool.Register()
    fmt.Println(newConnect) // Use the connection

    // Once done, release the connection
    cancelFunc()
}
```

### Configuration Options

Customize your connection pool using the following options:

- **WithCap(cap int)**: Set the maximum number of connections the pool can hold.
- **WithMaxFreeTime(maxFreeTime time.Duration)**: Define the maximum time a connection can stay idle before being automatically cleaned up.
- **WithAutoClearInterval(autoClearInterval time.Duration)**: Set the interval for the automatic cleanup task.
- **WithDealPanicMethod(dealPanicMethod func(panicInfo any))**: Provide a custom method to handle panic scenarios.
- **WithCloseMethod(closeMethod func(connect any))**: Specify a method to be called before closing a connection.

## Contributing

Contributions to improve the library are welcome. Please follow the standard fork-and-pull request workflow on GitHub.

## License

Specify your license here or state that the project is unlicensed and available for free use.