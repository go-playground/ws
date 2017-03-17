Package ws
==========
![Project status](https://img.shields.io/badge/version-0.1.0-green.svg)
[![GoDoc](https://godoc.org/github.com/go-playground/ws?status.svg)](https://godoc.org/github.com/go-playground/ws)
![License](https://img.shields.io/dub/l/vibe-d.svg)

Package ws creates a hub for WebSocket connections and abstracts away allot of the boilerplate code for managing connections using [Gorilla WebSocket](https://github.com/gorilla/websocket). The design of this library was inspired by [Dave Cheney's Talk on first Class Functions](https://www.youtube.com/watch?v=5buaPyJ0XeQ)

Installation
-------------
```shell
go get -u github.com/go-playground/ws
```

Examples
---------
- [Basic Usage](https://github.com/go-playground/ws/blob/master/examples/basic/main.go)
- [Custom Client Logic](https://github.com/go-playground/ws/blob/master/examples/custom/main.go)