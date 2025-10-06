package irpctestpkg

import (
	"fmt"
	"testing"
	"time"

	"github.com/marben/irpc"
	"github.com/marben/irpc/cmd/irpc/test/testtools"
)

func BenchmarkBinaryMarshal(b *testing.B) {
	// b.Log("running binmarshall test")

	p1, p2, err := testtools.CreateLocalTcpConnPipe()
	if err != nil {
		b.Fatalf("failed to create local tcp connection: %v", err)
	}

	serviceEp := irpc.NewEndpoint(p2)

	crw := &CountingReadWriteCloser{rwc: p1}
	clientEp := irpc.NewEndpoint(crw)

	serviceEp.RegisterService(newBinMarshalIRpcService(binMarshalImpl{}))

	c, err := newBinMarshalIRpcClient(clientEp)
	if err != nil {
		b.Fatalf("newBinMarshalIRpcClient: %+v", err)
	}

	callsN := []int{1, 10, 100}

	now := time.Now()
	nowPlusHour := now.Add(1 * time.Hour)

	b.ResetTimer()
	crw.Reset()
	for _, cn := range callsN {
		b.Logf("making %d calls", cn)
		b.Run(fmt.Sprintf("runs-%d", cn), func(b *testing.B) {
			crw.Reset()
			for range b.N {
				for range cn {
					if res := c.addHour(now); !res.Equal(nowPlusHour) {
						b.Fatalf("now: %q. addHour(): %q", now, res)
					}
				}
			}
			b.ReportMetric(float64(crw.RBytes())/float64(b.N), "rBytes/rpc")
			b.ReportMetric(float64(crw.WBytes())/float64(b.N), "wBytes/rpc")
		})
	}

	if err := clientEp.Close(); err != nil {
		b.Fatalf("clientEp.Close(): %v", err)
	}
}

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
	serviceEp.RegisterService(service)

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
