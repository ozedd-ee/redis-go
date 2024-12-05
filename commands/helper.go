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
// current state supports only EX option. Add support for others
func handleSet(key string, val string, e ExpiryOption, t string) string {
	tm, err := strconv.Atoi(t)
	expiryTime := time.Now().Add(time.Second * time.Duration(tm)) 
	if err != nil {
		return s.SerializeSimpleError("err", "Expiry time  not specified")
	}
	exp := Expiry{option: e, time: expiryTime}
	value := Value{expiry: exp, value: val}
	store[key] = value
	return s.SerializeSimpleString("OK")
}
