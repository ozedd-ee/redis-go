package commands

import (
	"strconv"
	"strings"
	"time"

	"github.com/ozedd-ee/redis-go/serializer"
)

var s = serializer.Serializer{}

var store = make(map[string]Value)

func ping() string {
	return s.SerializeSimpleString("PONG")
}

func echo(msg string) string {
	return s.SerializeSimpleString(msg)
}

func set(key string, val string, opts ...string) string {
	if len(opts) == 0 {
		var t time.Time
		exp := Expiry{option: NONE, time: t}
		value := Value{expiry: exp, value: val}
		store[key] = value
		return s.SerializeSimpleString("OK")
	}
	for i, v := range opts {
		V := strings.ToUpper(v)
		switch V {
		case "EX":
			return handleSetWithExpiry(key, val, EX, opts[i+1])
		case "PX":
			return handleSetWithExpiry(key, val, PX, opts[i+1])
		case "EXAT":
			return handleSetWithExpiry(key, val, EXAT, opts[i+1])
		case "PXAT":
			return handleSetWithExpiry(key, val, PXAT, opts[i+1])
		}
	}
	return s.SerializeSimpleError("err", "Invalid Expiry option specified")
}

func get(key string) string {
	val, ok := store[key]
	if !ok {
		return s.NullBulkString()
	}
	if val.isExpired() {
		delete(store, key)
		return s.NullBulkString()
	}
	return s.SerializeBulkString(val.value)
}

func exists(keys ...string) string {
	var counter int
	for _, k := range keys {
		_, ok := store[k]
		if ok {
			counter++
		}
	}
	return s.SerializeInteger(counter, true)
}

func deleteKey(keys ...string) string {
	var counter int
	for _, k:= range keys {
		v, ok := store[k]
		if ok {
			delete(store, v.value)
			counter++
		}
	}
	return s.SerializeInteger(counter, true)
}

func increment(key string) string {
	v, ok := store[key]
	// if key does not exist, set key to 0 and increment
	if !ok {
		set(key,"1")
		return s.SerializeInteger(1, true)
	}
	i , err := strconv.Atoi(v.value)
	if err != nil {
		return s.SerializeSimpleError("err", "value is not an integer or out of range")
	}
	i++
	v.value = strconv.Itoa(i)
	store[key] = v
	return s.SerializeInteger(i, true)
}

func decrement(key string) string {
	v, ok := store[key]
	// if key does not exist, set key to 0 and increment
	if !ok {
		set(key,"-1")
		return s.SerializeInteger(1, false)
	}
	i , err := strconv.Atoi(v.value)
	if err != nil {
		return s.SerializeSimpleError("err", "value is not an integer or out of range")
	}
	i--
	v.value = strconv.Itoa(i)
	store[key] = v
	return s.SerializeInteger(i, false)
}
