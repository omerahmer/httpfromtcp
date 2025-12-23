# Custom HTTP/1.1 Server in Go

A fully functional HTTP/1.1 server implemented **from scratch** in Go using raw TCP sockets

## Features
- **Raw TCP Server**  
  Handles TCP connections directly without using the `net/http` package anywhere

- **Request Parsing**  
  Supports GET requests, parses request lines, headers, and handles both plaintext and binary payloads

- **Response Generation**  
  Builds HTTP responses from scratch, including status line, headers, body content, and optionally, trailers

- **Header Management**  
  Allows setting and merging multiple headers for each response

- **Chunked Transfer Encoding**  
  Implements HTTP/1.1 chunked encoding to stream data efficiently without knowing content length in advance

- **Binary Data Support**  
  Correctly reads and responds with binary data (e.g. for videos), not limited to plain text
