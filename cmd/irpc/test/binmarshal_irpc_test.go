package irpctestpkg

import (
	"testing"
	"time"

	"github.com/marben/irpc/cmd/irpc/test/testtools"
)

func TestTime(t *testing.T) {
	localEp, remoteEp, err := testtools.CreateLocalTcpEndpoints()
	if err != nil {
		t.Fatalf("create enpoints: %v", err)
	}
	defer localEp.Close()
	defer remoteEp.Close()

	remoteEp.RegisterService(newBinMarshalIrpcService(binMarshalImpl{}))

	c, err := newBinMarshalIrpcClient(localEp)
	if err != nil {
		t.Fatalf("new client(): %+v", err)
	}

	start := time.Now()
	res := c.addHour(start)
	if res.Sub(start) != time.Hour {
		t.Errorf("got wrong time back: %s", res)
	}

	reflectRes := c.reflect(start)
	if reflectRes.Compare(start) != 0 {
		t.Errorf("c.reflect(): %s != %s", reflectRes, start)
	}

	myTimeRes := c.addMyHour(myTime(start))
	if time.Time(myTimeRes).Compare(start.Add(time.Hour)) != 0 {
		t.Fatalf("myTimeRes: %s", time.Time(myTimeRes))
	}

	myStructTimeRes := c.addMyStructHour(myStructTime{myTime(start)})
	if myStructTimeRes.Compare(myTime(start).Add(time.Hour)) != 0 {
		t.Fatalf("myStructTimeRes: %+v", myStructTimeRes)
	}
}
