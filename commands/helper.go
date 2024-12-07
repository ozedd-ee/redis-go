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
	value  string
}

const (
	EX ExpiryOption = iota
	PX
	EXAT
	PXAT
	NONE
)

func (e ExpiryOption) String() string {
	switch e {
	case EX:
		return "EX"
	case PX:
		return "PX"
	case EXAT:
		return "EXAT"
	case PXAT:
		return "PXAT"
	default:
		return "NONE"
	}
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

	var expiryTime time.Time
	switch e {
	case EX:
		expiryTime = time.Now().Add(time.Second * time.Duration(tm)) 
	case PX:
		expiryTime = time.Now().Add(time.Millisecond * time.Duration(tm))
	case EXAT:
		expiryTime = time.Unix(int64(tm), 0) 
	case PXAT:
		expiryTime = time.UnixMilli(int64(tm))
	}

	exp := Expiry{option: e, time: expiryTime}
	value := Value{expiry: exp, value: val}
	store[key] = value
	return s.SerializeSimpleString("OK")
}
