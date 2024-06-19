# TritonHTTP

## Overview

This project builds a simple web server called TritonHTTP, implementing a subset of the HTTP/1.1 protocol. The server will handle client requests, serve files, and manage persistent connections.

## Basic Web Server Functionality

A web server listens for connections on a specific port, processes client requests, and sends back responses. TritonHTTP supports:

* Persistent connections
* Pipelined requests
* A 5-second timeout for idle connections

## TritonHTTP Specification

### HTTP Messages

Both request and response messages are plain-text ASCII with headers and optional body sections.

Initial Request Line: `GET <URL> HTTP/1.1`

Initial Response Line: `HTTP/1.1 <status code> <status description>`

### Headers

Headers provide additional information and follow the format:
`<key>: <value>\r\n`

### Message Body

Request messages do not have a body. Response messages may include a body for a 200 OK status, containing the requested file's bytes.
