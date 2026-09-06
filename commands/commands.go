package commands

import (
	"strconv"
	"strings"
	"time"

	"redis-go/serializer"
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
	str, ok := val.value.(string)
	if !ok {
		return s.SerializeSimpleError("WRONGTYPE", "value stored at key is not a string")
	}
	return s.SerializeBulkString(str)
}

func exists(keys ...string) string {
	var counter int
	for _, k := range keys {
		v, ok := store[k]
		if !ok {
			continue
		}
		if v.isExpired() {
			delete(store, k)
			continue
		}
		counter++
	}
	return s.SerializeInteger(counter)
}

func deleteKey(keys ...string) string {
	var counter int
	for _, k := range keys {
		_, ok := store[k]
		if ok {
			delete(store, k)
			counter++
		}
	}
	return s.SerializeInteger(counter)
}

func increment(key string) string {
	v, ok := store[key]
	if !ok {
		set(key, "1")
		return s.SerializeInteger(1)
	}
	str, ok := v.value.(string)
	if !ok {
		return s.SerializeSimpleError("err", "value is not an integer")
	}
	i, err := strconv.Atoi(str)
	if err != nil {
		return s.SerializeSimpleError("err", "value is not an integer or out of range")
	}
	i++
	v.value = strconv.Itoa(i)
	store[key] = v
	return s.SerializeInteger(i)
}

func decrement(key string) string {
	v, ok := store[key]
	if !ok {
		set(key, "-1")
		return s.SerializeInteger(-1)
	}
	str, ok := v.value.(string)
	if !ok {
		return s.SerializeSimpleError("err", "value is not an integer")
	}
	i, err := strconv.Atoi(str)
	if err != nil {
		return s.SerializeSimpleError("err", "value is not an integer or out of range")
	}
	i--
	v.value = strconv.Itoa(i)
	store[key] = v
	return s.SerializeInteger(i)
}

func lpush(key string, elements ...string) string {
	v, ok := store[key]
	if !ok {
		cp := make([]string, len(elements))
		copy(cp, elements)
		reverse(cp)
		var t time.Time
		exp := Expiry{option: NONE, time: t}
		value := Value{expiry: exp, value: cp}
		store[key] = value
		return s.SerializeInteger(len(cp))
	}
	initial, ok := v.value.([]string)
	if !ok {
		return s.SerializeSimpleError("WRONGTYPE", "value stored at key is not a list")
	}
	// Reverse incoming elements so LPUSH list a b c → [c,b,a,...] (same as pushing one at a time)
	cp := make([]string, len(elements))
	copy(cp, elements)
	reverse(cp)
	val := append(cp, initial...)
	v.value = val
	store[key] = v
	return s.SerializeInteger(len(val))
}

func rpush(key string, elements ...string) string {
	v, ok := store[key]
	if !ok {
		var t time.Time
		exp := Expiry{option: NONE, time: t}
		value := Value{expiry: exp, value: elements}
		store[key] = value
		return s.SerializeInteger(len(elements))
	}
	initial, ok := v.value.([]string)
	if !ok {
		return s.SerializeSimpleError("WRONGTYPE", "value stored at key is not a list")
	}
	initial = append(initial, elements...)
	v.value = initial
	store[key] = v
	return s.SerializeInteger(len(initial))
}

func lrange(key string, start string, stop string) string {
	val, ok := store[key]
	if !ok {
		return s.SerializeArray()
	}
	v, ok := val.value.([]string)
	if !ok {
		return s.SerializeSimpleError("WRONGTYPE", "value stored at key is not a list")
	}

	begin, err := strconv.Atoi(start)
	if err != nil {
		return s.SerializeSimpleError("err", "invalid start index specified")
	}
	end, err := strconv.Atoi(stop)
	if err != nil {
		return s.SerializeSimpleError("err", "invalid end index specified")
	}

	n := len(v)
	if begin < 0 {
		begin = n + begin
		if begin < 0 {
			begin = 0
		}
	}
	if end < 0 {
		end = n + end
	}
	if begin > end || begin >= n {
		return s.SerializeArray()
	}
	if end >= n {
		end = n - 1
	}

	var itemBuffer []string
	for i := begin; i <= end; i++ {
		itemBuffer = append(itemBuffer, v[i])
	}
	return s.SerializeArray(itemBuffer...)
}
