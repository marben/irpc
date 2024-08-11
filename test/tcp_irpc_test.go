package irpctestpkg

import (
	"testing"

	"github.com/marben/irpc/test/testtools"
)

func TestTcpClientServer(t *testing.T) {
	ep1, ep2, err := testtools.CreateLocalTcpEndpoints()
	if err != nil {
		t.Fatalf("failed to create local tcp endpoints: %v", err)
	}

	service := newTcpTestApiIRpcService(tcpTestApiImpl{})
	ep1.RegisterServices(service)

	client, err := newTcpTestApiIRpcClient(ep2)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	res, err := client.Div(4, 2)
	if err != nil {
		t.Fatalf("div failed: %v", err)
	}
	if res != 2 {
		t.Fatalf("wrong result: %f", res)
	}

	if err := ep1.Close(); err != nil {
		t.Fatalf("ep1.Close(): %v", err)
	}
	if err := ep2.Close(); err != nil {
		t.Logf("ep2.Close(): %v", err)
	}
}

// func TestClosingConnection1(t *testing.T) {
// 	c1, c2, err := testtools.CreateLocalTcpConnPipe()
// 	if err != nil {
// 		t.Fatalf("failed to create local tcp pipe")
// 	}
// 	ep1 := irpc.NewEndpoint(c1)
// 	ep2 := irpc.NewEndpoint(c2)

// 	// errC1 := make(chan error)
// 	// errC2 := make(chan error)
// 	// go func() { errC1 <- ep1.Serve(c1) }()
// 	// go func() { errC2 <- ep2.Serve(c2) }()

// 	if err := c1.Close(); err != nil {
// 		t.Fatalf("failed to close connection c1: %v", err)
// 	}

// 	// both endpoint.Serve() errors out
// 	err1 := <-errC1
// 	err2 := <-errC2
// 	t.Logf("err1: %v", err1)
// 	t.Logf("err2: %v", err2)
// }

/*
func TestClosingConnection2(t *testing.T) {
	c1, c2, err := testtools.CreateLocalTcpConnPipe()
	if err != nil {
		t.Fatalf("failed to create local tcp pipe")
	}
	ep1 := irpc.NewEndpoint(c1)
	ep2 := irpc.NewEndpoint(c2)

	errC1 := make(chan error)
	errC2 := make(chan error)
	go func() { errC1 <- ep1.Serve(c1) }()
	go func() { errC2 <- ep2.Serve(c2) }()

	if err := c2.Close(); err != nil {
		t.Fatalf("failed to close connection c2: %v", err)
	}

	// both endpoint.Serve() errors out
	err1 := <-errC1
	err2 := <-errC2
	t.Logf("err1: %v", err1)
	t.Logf("err2: %v", err2)
}
*/
