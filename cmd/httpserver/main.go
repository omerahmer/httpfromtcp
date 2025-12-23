package main

import (
	"crypto/sha256"
	"fmt"
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

func toStr(bytes []byte) string {
	var out strings.Builder
	for _, b := range bytes {
		fmt.Fprintf(&out, "%02x", b)
	}
	return out.String()
}

const port = 42069

func respond400() []byte {
	return []byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)
}

func respond500() []byte {
	return []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)
}

func respond200() []byte {
	return []byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)
}

func main() {
	server, err := server.Serve(port, func(w *response.Writer, req *request.Request) {
		h := response.GetDefaultHeaders(0)
		body := respond200()
		status := response.StatusOK

		if req.RequestLine.RequestTarget == "/yourproblem" {
			body = respond400()
			status = response.StatusBadRequest

		} else if req.RequestLine.RequestTarget == "/myproblem" {
			body = respond500()
			status = response.StatusInternalServerError

		} else if req.RequestLine.RequestTarget == "/video" {
			f, _ := os.ReadFile("assets/vim.mp4")

			h.Replace("Content-Type", "video/mp4")
			h.Replace("Content-Length", fmt.Sprintf("%d", len(f)))

			w.WriteStatusLine(response.StatusOK)
			w.WriteHeaders(*h)
			w.WriteBody(f)

		} else if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
			target := req.RequestLine.RequestTarget
			url := "https://httpbin.org/" + target[len("/httpbin/"):]

			cmd := exec.Command("curl", "-s", "-N", "-X", "GET", "-H", "Accept: application/json", url)
			cmd.Env = os.Environ()

			stdout, err := cmd.StdoutPipe()
			if err != nil {
				body = respond500()
				status = response.StatusInternalServerError
			} else if err := cmd.Start(); err != nil {
				body = respond500()
				status = response.StatusInternalServerError
			} else {
				w.WriteStatusLine(response.StatusOK)

				h.Delete("Content-Length")
				h.Set("Transfer-Encoding", "chunked")
				h.Replace("Content-Type", "text/plain")
				h.Set("Trailers", "X-Content-SHA256")
				h.Set("Trailers", "X-Content-Length")

				w.WriteHeaders(*h)

				fullBody := []byte{}
				buf := make([]byte, 32)
				for {
					n, err := stdout.Read(buf)
					if n > 0 {
						fullBody = append(fullBody, buf[:n]...)
						w.WriteBody(fmt.Appendf(nil, "%x\r\n", n))
						w.WriteBody(buf[:n])
						w.WriteBody([]byte("\r\n"))
					}
					if err != nil {
						break
					}
				}
				w.WriteBody([]byte("0\r\n"))

				trailer := headers.NewHeaders()
				out := sha256.Sum256(fullBody)
				trailer.Set("X-Content-SHA256", toStr(out[:]))
				trailer.Set("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))
				w.WriteHeaders(*trailer)

				cmd.Wait()
				return
			}
		}
		h.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
		h.Replace("Content-Type", "text/html")
		w.WriteStatusLine(status)
		w.WriteHeaders(*h)
		w.WriteBody(body)
	})
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
