package irpctestpkg

import "time"

// tests the binary.Marshaller and binary.Appender interfaces

type myTime time.Time

type myStructTime struct {
	myTime
}

func (mt myTime) Add(hour time.Duration) myTime {
	return myTime(time.Time(mt).Add(hour))
}

func (mt myTime) MarshalBinary() (data []byte, err error) {
	return time.Time(mt).MarshalBinary()
}

func (mt *myTime) UnmarshalBinary(data []byte) error {
	return (*time.Time)(mt).UnmarshalBinary(data)
}

func (mt myTime) Compare(u myTime) int {
	return time.Time(mt).Compare(time.Time(u))
}

type structContainingBinMarshallable struct {
	t time.Time
}

//go:generate go run ../
type binMarshal interface {
	reflect(t time.Time) time.Time
	addHour(t time.Time) time.Time
	addMyHour(t myTime) myTime
	addMyStructHour(t myStructTime) myStructTime
	structPass(st structContainingBinMarshallable) structContainingBinMarshallable
}

type binMarshalImpl struct{}

// structPass implements binMarshal.
func (b binMarshalImpl) structPass(st structContainingBinMarshallable) structContainingBinMarshallable {
	panic("unimplemented")
}

var _ binMarshal = binMarshalImpl{}

func (b binMarshalImpl) addHour(t time.Time) time.Time {
	return t.Add(time.Hour)
}

func (b binMarshalImpl) reflect(t time.Time) time.Time {
	return t
}

func (b binMarshalImpl) addMyHour(t myTime) myTime {
	return t.Add(1 * time.Hour)
}

func (b binMarshalImpl) addMyStructHour(t myStructTime) myStructTime {
	return myStructTime{t.Add(time.Hour)}
}
