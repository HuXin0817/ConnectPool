# Connect Pool

Connect Pool is a Go package designed to efficiently manage a pool of connections. It abstracts the complexity of handling multiple concurrent connections, ensuring optimal resource utilization and simplifying the process of creating, using, and disposing of connections. This document outlines the Connect Pool's functionalities and provides a guide on how to use it in your Go applications.

## Features

- **Dynamic Connection Management:** Automatically creates and closes connections based on demand, up to a specified maximum pool size.
- **Connection Reuse:** Allows for the reuse of idle connections, reducing the overhead of establishing new connections.
- **Concurrent Safe:** Designed to be safe for concurrent use by multiple goroutines.
- **Auto-Clean:** Periodically clears idle connections that exceed a specified maximum free time, helping to release unused resources.
- **Panic Handling:** Provides a mechanism to handle panics that occur during connection operations, enhancing the robustness of your application.
- **Customizable Connection Logic:** Supports custom connection creation, closing, and panic handling functions to fit specific requirements.

![](png/pool.png)

## Usage

Below is a basic guide on how to integrate and use the Connect Pool in a Go application.

### Installation

Ensure you have Go installed on your machine (visit [Go's official site](https://golang.org/dl/) for installation instructions). Then, input the order on your Terminal:

```shell
$ go get -u github.com/HuXin0817/ConnectPool
```

import the Connect Pool package into your project:

```go
import "github.com/HuXin0817/ConnectPool"
```

### Creating a New Connection Pool

To create a new connection pool, you need to specify the maximum size of the pool and a function that defines how each connection is established:

```go
pool := pool.NewConnectPool(100, mockConnectMethod)
```

`mockConnectMethod` is a function that returns a new connection object. This method is called whenever the pool needs to create a new connection.

### Registering and Using a Connection

To use a connection from the pool:

```go
conn, cancelFunc := pool.Register()
// Use the connection
cancelFunc() // Release the connection back to the pool
```

### Setting Up Custom Handlers

You can set custom methods for handling connection closure and panic scenarios:

```go
pool.SetCloseMethod(mockCloseFunc)
pool.SetDealPanicMethod(mockDealPanicMethod)
```

### Configuring Auto-Clean Parameters

The pool can be configured to automatically clean up idle connections that exceed a certain duration:

```go
pool.SetMaxFreeTime(30 * time.Second) // Set the maximum idle time before a connection is closed
pool.SetAutoClearInterval(1 * time.Minute) // Set the interval for the auto-clean process
```

### Closing the Pool

To close the pool and release all its resources:

```go
pool.Close()
```
