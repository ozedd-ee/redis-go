package commands

import "strconv"

type ExpiryOption int

type Expiry struct {
	option ExpiryOption
	time   uint
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

func handleSet(key string, val string, e ExpiryOption, t string) string {
	time, err := strconv.Atoi(t)
	if err != nil {
		return s.SerializeSimpleError("err", "Expiry time  not specified")
	}
	exp := Expiry{option: e, time: uint(time)}
	value := Value{expiry: exp, value: val}
	store[key] = value
	return s.SerializeSimpleString("OK")
}
