Package ws
==========
![Project status](https://img.shields.io/badge/version-0.1.0-green.svg)
[![Build Status](https://semaphoreci.com/api/v1/joeybloggs/ws/branches/master/badge.svg)](https://semaphoreci.com/joeybloggs/ws)
[![Coverage Status](https://coveralls.io/repos/github/go-playground/ws/badge.svg?branch=master)](https://coveralls.io/github/go-playground/ws?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-playground/ws)](https://goreportcard.com/report/github.com/go-playground/ws)
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

Contributing
------------
Make a pull request

Package Versioning
----------
I'm jumping on the vendoring bandwagon, you should vendor this package as I will not
be creating different version with gopkg.in like allot of my other libraries.

Why? because my time is spread pretty thin maintaining all of the libraries I have + LIFE,
it is so freeing not to worry about it and will help me keep pouring out bigger and better
things for you the community.

License
------
Distributed under MIT License, please see license file in code for more details.
