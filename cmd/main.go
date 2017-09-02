package main

import (
	"bytes"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"

	"github.com/elazarl/goproxy/transport"

	"github.com/elazarl/goproxy"
	"github.com/moolen/mitmdump"
	uuid "github.com/satori/go.uuid"
)

// hack for now
var fingerprintCheck bool

func main() {
	headerDebug := *flag.String("debug-header", "X-Debug", "if a request has a debug-header and a trace-id-header it will be logged")
	traceHeader := *flag.String("trace-id-header", "X-Trace-ID", "both trace-id-header and debug-header are needed in order to dump the request")
	fingerprintCheck = *flag.Bool("fingerprint-check", false, "enables the fingerprint check")
	ignoreMissingTrace := *flag.Bool("ignore-missing-trace", false, "missing trace-ids are logged aswell")
	logAll := *flag.Bool("log-all", false, "logs all request even if trace-id-header or debug-header is missing")
	verbose := *flag.Bool("v", false, "should every proxy request be logged to stdout")
	addr := flag.String("l", ":8080", "on which address should the proxy listen")
	flag.Parse()

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = verbose
	logger, err := mitmdump.NewLogger("/home/moritz/go/src/github.com/moolen/mitmdump/db")
	if err != nil {
		log.Fatal("can't open log file", err)
	}

	tr := transport.Transport{
		Proxy: transport.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify:    true,
			VerifyPeerCertificate: verifyCert,
		}}

	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		ctx.RoundTripper = goproxy.RoundTripperFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (res *http.Response, err error) {
			_, res, err = tr.DetailedRoundTrip(req)
			return res, err
		})
		if req.Header.Get(headerDebug) == "" && !logAll {
			return req, nil
		}
		traceID := req.Header.Get(traceHeader)
		if traceID == "" {
			if !ignoreMissingTrace {
				log.Printf("missing trace-Id header")
				return req, nil
			}
			traceID = uuid.NewV4().String()
		}
		logCtx := &mitmdump.LogContext{
			TraceID: traceID,
			UUID:    uuid.NewV4().String(),
		}
		ctx.UserData = logCtx
		err := logger.LogReq(req, logCtx)
		if err != nil {
			log.Printf("error logging req: %v", err)
		}
		return req, nil
	})
	proxy.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		if ctx.UserData == nil {
			return resp
		}
		logCtx, ok := ctx.UserData.(*mitmdump.LogContext)
		if !ok {
			log.Printf("could not get log context")
			return resp
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

var trustedCertsSHA1 = [][]byte{
	// github.com
	[]byte{0xd7, 0x9f, 0x07, 0x61, 0x10, 0xb3, 0x92, 0x93, 0xe3, 0x49, 0xac, 0x89, 0x84, 0x5b, 0x03, 0x80, 0xc1, 0x9e, 0x2f, 0x8b},
}

func verifyCert(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	if !fingerprintCheck {
		return nil
	}
	first := rawCerts[0]
	for _, v := range trustedCertsSHA1 {
		fingerprint := sha1.Sum(first)
		if bytes.Compare(v, fingerprint[:]) == 0 {
			return nil
		}
	}
	return errors.New("certificate not trusted")
}
