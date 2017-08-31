# mitmdump
This is a HTTP proxy that acts as a man-in-the-mittle that hijacks https connections instead of transparently forwarding them to you. 

## Testing with curl (easy)
Run the proxy with `go run cmd/main.go` and issue an http using curl: `curl -vikx http://localhost:8080 https://github.com`.

## Testing with a browser (you trust me?)
First, import [this certificate](https://github.com/elazarl/goproxy/blob/master/ca.pem) into your browser certificate chain. Start the proxy with `go run cmd/main.go` and start your browser: `google-chrome-stable --proxy-server=http://localhost:8080 https://github.com`