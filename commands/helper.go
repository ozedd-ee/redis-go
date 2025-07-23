package commands

import (
	"strconv"
	"time"
)

type ExpiryOption int

type Expiry struct {
	option ExpiryOption
	time   time.Time
}

type Value struct {
	expiry Expiry
	value  any
}

const (
	EX ExpiryOption = iota
	PX
	EXAT
	PXAT
	NONE
)

var expiryStrings = map[ExpiryOption]string{
	EX:    "EX",
	PX:    "PX",
	EXAT:  "EXAT",
	PXAT:  "PXAT",
}

func (e ExpiryOption) String() string {
	option, ok := expiryStrings[e]
	if !ok {
		return "NONE"
	}
	return option
}

func (v  Value) isExpired() bool {
	if v.expiry.time.IsZero() {
		return false
	}
	return time.Now().After(v.expiry.time)
}

func handleSetWithExpiry(key string, val string, e ExpiryOption, t string) string {
	tm, err := strconv.Atoi(t)
	if err != nil {
		return s.SerializeSimpleError("err", "Expiry time  not specified")
	}

	expiryTime := map[ExpiryOption]time.Time{
		EX:    time.Now().Add(time.Second * time.Duration(tm)),
		PX:    time.Now().Add(time.Millisecond * time.Duration(tm)),
		EXAT:  time.Unix(int64(tm), 0),
		PXAT:  time.UnixMilli(int64(tm)),
	}

	exp := Expiry{option: e, time: expiryTime[e]}
	value := Value{expiry: exp, value: val}
	store[key] = value
	return s.SerializeSimpleString("OK")
}

func reverse(slice []string) {
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
}
