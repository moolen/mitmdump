package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"

	"github.com/elazarl/goproxy"
	"github.com/moolen/mitmdump"
	uuid "github.com/satori/go.uuid"
)

func main() {
	verbose := flag.Bool("v", true, "should every proxy request be logged to stdout")
	addr := flag.String("l", ":8080", "on which address should the proxy listen")
	flag.Parse()
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = *verbose
	logger, err := mitmdump.NewLogger("/home/moritz/go/src/github.com/moolen/mitmdump/db")
	if err != nil {
		log.Fatal("can't open log file", err)
	}

	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		traceID := req.Header.Get("X-TraceId")
		if len(traceID) == 0 {
			log.Printf("missing X-TraceId header")
		}
		uid := uuid.NewV4().String()
		logCtx := &mitmdump.LogContext{
			TraceID: traceID,
			UUID:    uid,
		}
		ctx.UserData = logCtx
		err := logger.LogReq(req, logCtx)
		if err != nil {
			log.Printf("error logging req: %v", err)
		}
		return req, nil
	})
	proxy.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		logCtx, ok := ctx.UserData.(*mitmdump.LogContext)
		if !ok {
			log.Printf("missing X-TraceId header")
		}
		err := logger.LogRes(resp, logCtx)
		if err != nil {
			log.Printf("error logging res: %v", err)
		}
		return resp
	})

	l, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal("listen:", err)
	}
	lis := mitmdump.NewGracefulListener(l)
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt)
	go func() {
		<-ch
		log.Println("Got SIGINT exiting")
		lis.Add(1)
		lis.Close()
		logger.Close()
		lis.Done()
	}()
	log.Println("Starting Proxy")
	http.Serve(lis, proxy)
	lis.Wait()
	log.Println("All connections closed - exit")
}
