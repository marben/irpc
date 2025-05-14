package irpctestpkg

import (
	"github.com/marben/irpc"
	"github.com/marben/irpc/cmd/irpc/test/testtools"
	"testing"
)

/*
	func BenchmarkClientRegister2(b *testing.B) {
		// SERVER
		server := irpc.NewServer()

		sl, err := net.Listen("tcp", ":")
		if err != nil {
			b.Fatalf("failed to create listener: %v", err)
		}

		serverErrC := make(chan error, 1)
		go func() { serverErrC <- server.Serve(sl) }()

		// // CLIENT
		// clientConn, err := net.Dial("tcp", sl.Addr().String())
		// if err != nil {
		// 	b.Fatalf("net.Dial(%s): %v", sl.Addr().String(), err)
		// }
		// clientEp := irpc.NewEndpoint()
		// clientErrC := make(chan error)
		// go func() { clientErrC <- clientEp.Serve(clientConn) }()

		skew := 2
		service := newBasicAPIIRpcService(basicApiImpl{skew: skew})

		server.RegisterService(service)

		// b.ResetTimer()
		for range b.N {
			// CLIENT CONN
			log.Println("connecting")
			clientConn, err := net.Dial("tcp", sl.Addr().String())
			if err != nil {
				b.Fatalf("net.Dial(%s): %v", sl.Addr().String(), err)
			}
			log.Println("connected")
			clientEp := irpc.NewEndpoint(clientConn)

			// CLIENT REGISTER
			log.Println("registering")
			_, err = newBasicAPIIRpcClient(clientEp)
			if err != nil {
				b.Fatalf("newBasicAPIIrpcClient(): %+v", err)
			}
			log.Println("registered")

			if err := clientEp.Close(); err != nil {
				b.Fatalf("clientEp.Close(): %+v", err)
			}

			log.Println("another iter")
		}
		if err := server.Close(); err != nil {
			b.Fatalf("server.Close(): %+v", err)
		}
		serverServeErr := <-serverErrC
		if serverServeErr != irpc.ErrServerClosed {
			b.Fatalf("serverServeErr: %+v", serverServeErr)
		}
	}

	func BenchmarkClientRegister(b *testing.B) {
		var rb, wb int // read/write bytes
		for range b.N {
			// b.StopTimer()
			p1, p2, err := testtools.CreateLocalTcpConnPipe()
			if err != nil {
				b.Fatalf("create tcp pipe: %v", err)
			}

			clientEp := irpc.NewEndpoint(p1)

			crw := &CountingReadWriteCloser{rwc: p2}
			serviceEp := irpc.NewEndpoint(p2)

			skew := 2
			service := newBasicAPIIRpcService(basicApiImpl{skew: skew})
			serviceEp.RegisterServices(service)

			// b.StartTimer()
			// register client (network communication)
			_, err = newBasicAPIIRpcClient(clientEp)
			if err != nil {
				b.Fatalf("failed to create client: %v", err)
			}
			// b.StopTimer()

			if err := clientEp.Close(); err != nil {
				b.Fatalf("clientEp.Close(): %v", err)
			}
			// if err := serviceEp.Close(); err != nil {
			// 	b.Logf("serviceEp.Close(): %v", err)
			// }

			// rb, wb = crw.rBytes, crw.wBytes
			rb += crw.rBytes
			wb += crw.wBytes
			// b.StartTimer()

			if err := serviceEp.Close(); err != nil {
				b.Fatalf("serviceEp.Close(): %+v", err)
			}
			if err := clientEp.Close(); err != nil {
				b.Fatalf("clientEp.Close(): %+v", err)
			}
		}
		b.ReportMetric(float64(rb)/float64(b.N), "rBytes/rpc")
		b.ReportMetric(float64(wb)/float64(b.N), "wBytes/rpc")
	}
*/
func BenchmarkAddInt64(b *testing.B) {
	p1, p2, err := testtools.CreateLocalTcpConnPipe()
	if err != nil {
		b.Fatalf("failed to create local tcp connection: %v", err)
	}

	serviceEp := irpc.NewEndpoint(p2)

	crw := &CountingReadWriteCloser{rwc: p1}
	clientEp := irpc.NewEndpoint(crw)

	skew := 2
	service := newBasicAPIIRpcService(basicApiImpl{skew: skew})
	serviceEp.RegisterServices(service)

	c, err := newBasicAPIIRpcClient(clientEp)
	if err != nil {
		b.Fatalf("failed to create client: %v", err)
	}

	b.ResetTimer()
	crw.Reset()
	for range b.N {
		var x, y int64 = 1, 10
		if res := c.addInt64(x, y); res != x+y+int64(skew) {
			b.Fatalf("%d + %d != %d", x, y, res)
		}
	}
	// b.Logf("N: %d, rBytes: %d, wBytes: %d", b.N, crw.RBytes(), crw.WBytes())
	b.ReportMetric(float64(crw.RBytes())/float64(b.N), "rBytes/rpc")
	b.ReportMetric(float64(crw.WBytes())/float64(b.N), "wBytes/rpc")

	if err := clientEp.Close(); err != nil {
		b.Fatalf("clientEp.Close(): %v", err)
	}
	// if err := serviceEp.Close(); err != nil {
	// 	b.Fatalf("serviceEp.Close(): %+v", err)
	// }
	// if err := clientEp.Close(); err != nil {
	// 	b.Fatalf("clientEp.Close(): %+v", err)
	// }
}
