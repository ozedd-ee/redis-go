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
	switch val.value.(type) {
	case string:
		return s.SerializeBulkString(val.value.(string))
	default:
		return s.SerializeSimpleError("err", "type stored at key is not a string")
	}
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
		_, ok := store[k]
		if ok {
			delete(store, k)
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
	str, ok := v.value.(string)
	if !ok {
		return s.SerializeSimpleError("err", "value is not an integer")
	}
	i , err := strconv.Atoi(str)
	if err != nil {
		return s.SerializeSimpleError("err", "value is not an integer or out of range")
	}
	i++
	v.value = strconv.Itoa(i)
	store[key] = v
	if i < 0 {
		return s.SerializeInteger(i, false)
	}
	return s.SerializeInteger(i, true)
}

func decrement(key string) string {
	v, ok := store[key]
	// if key does not exist, set key to 0 and decrement
	if !ok {
		set(key,"-1")
		return s.SerializeInteger(1, false)
	}
	str, ok := v.value.(string)
	if !ok {
		return s.SerializeSimpleError("err", "value is not an integer")
	}
	i , err := strconv.Atoi(str)
	if err != nil {
		return s.SerializeSimpleError("err", "value is not an integer or out of range")
	}
	i--
	v.value = strconv.Itoa(i)
	store[key] = v
	if i < 0 {
		return s.SerializeInteger(i, false)
	}
	return s.SerializeInteger(i, true)
}

func lpush(key string, elements ...string) string {
	v, ok := store[key]
	if !ok {
		reverse(elements)
		var t time.Time
		exp := Expiry{option: NONE, time: t}
		value := Value{expiry: exp, value: elements}
		store[key] = value
		return s.SerializeInteger(len(value.value.([]string)), true)
	}
	initial, ok := v.value.([]string)
	if !ok {
		return s.SerializeSimpleError("err", "value stored at key is not a list")
	}
	var val []string
	val = append(val, elements...)
	val = append(val, initial...)
	v.value = val
	store[key] = v
	return s.SerializeInteger(len(v.value.([]string)), true)
}

func rpush(key string, elements ...string) string {
	v, ok := store[key]
	if !ok {
		var t time.Time
		exp := Expiry{option: NONE, time: t}
		value := Value{expiry: exp, value: elements}
		store[key] = value
		return s.SerializeInteger(len(value.value.([]string)), true)
	}
	initial, ok := v.value.([]string)
	if !ok {
		return s.SerializeSimpleError("err", "value stored at key is not a list")
	}
	initial = append(initial, elements...)
	v.value = initial
	store[key] = v
	return s.SerializeInteger(len(v.value.([]string)), true)
}

func lrange(key string, start string, stop string) string {
	val, ok := store[key]
	if !ok {
		return s.SerializeArray()
	}
	v, ok := val.value.([]string)
	if !ok {
		return s.SerializeSimpleError("err", "value stored at key is not a list")
	}

	begin, err := strconv.Atoi(start)
	if err != nil {
		return s.SerializeSimpleError("err", "invalid start index specified")
	}
	end, err := strconv.Atoi(stop)
	if err != nil {
		return s.SerializeSimpleError("err", "invalid end index specified")
	}

	var itemBuffer []string
	if end < 0 {
		end = len(v) + end
	}
	for i, j := begin, end; i <= j; i++ {
		itemBuffer = append(itemBuffer, v[i])
	}
	return s.SerializeArray(itemBuffer...)
}
