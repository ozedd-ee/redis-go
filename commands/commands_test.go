package commands

import (
	"strings"
	"testing"

	"redis-go/serializer"
)

func sendRawCommand(cmd string) string {
	s := serializer.Serializer{}
	resp := s.SerializeArray(strings.Split(cmd, " ")...)
	return HandleCommand(mustDeserialize(&s, resp), &s)
}

func mustDeserialize(s *serializer.Serializer, resp string) string {
	cmd, err := s.DeserializeMessage(resp)
	if err != nil {
		panic(err)
	}
	return cmd
}

func TestPing(t *testing.T) {
	got := sendRawCommand("PING")
	want := "+PONG\r\n"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestSetGet(t *testing.T) {
	sendRawCommand("SET mykey hello")
	got := sendRawCommand("GET mykey")
	want := "$5\r\nhello\r\n"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestGetMissingKey(t *testing.T) {
	got := sendRawCommand("GET __no_such_key__")
	if got != "$-1\r\n" {
		t.Errorf("expected null bulk string, got %q", got)
	}
}

func TestIncrDecr(t *testing.T) {
	sendRawCommand("SET count 10")
	got := sendRawCommand("INCR count")
	if got != ":11\r\n" {
		t.Errorf("expected :11, got %q", got)
	}
	got = sendRawCommand("DECR count")
	if got != ":10\r\n" {
		t.Errorf("expected :10, got %q", got)
	}
}

func TestIncrNonExistent(t *testing.T) {
	got := sendRawCommand("INCR __incr_new__")
	if got != ":1\r\n" {
		t.Errorf("expected :1, got %q", got)
	}
}

func TestDecrNonExistent(t *testing.T) {
	got := sendRawCommand("DECR __decr_new__")
	if got != ":-1\r\n" {
		t.Errorf("expected :-1, got %q", got)
	}
}

func TestIncrNegative(t *testing.T) {
	sendRawCommand("SET negcount -5")
	got := sendRawCommand("INCR negcount")
	if got != ":-4\r\n" {
		t.Errorf("expected :-4, got %q", got)
	}
}

func TestIncrNonInteger(t *testing.T) {
	sendRawCommand("SET strval hello")
	got := sendRawCommand("INCR strval")
	if !strings.HasPrefix(got, "-ERR") {
		t.Errorf("expected error, got %q", got)
	}
}

func TestExistsDel(t *testing.T) {
	sendRawCommand("SET x val")
	got := sendRawCommand("EXISTS x")
	if got != ":1\r\n" {
		t.Errorf("expected EXISTS 1, got %q", got)
	}
	sendRawCommand("DEL x")
	got = sendRawCommand("EXISTS x")
	if got != ":0\r\n" {
		t.Errorf("expected EXISTS 0 after delete, got %q", got)
	}
}

func TestLpushRpushLrange(t *testing.T) {
	sendRawCommand("DEL list")
	sendRawCommand("LPUSH list one two")
	sendRawCommand("RPUSH list three")

	got := sendRawCommand("LRANGE list 0 2")
	want := "*3\r\n$3\r\ntwo\r\n$3\r\none\r\n$5\r\nthree\r\n"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestLpushMultipleExisting(t *testing.T) {
	sendRawCommand("DEL lpush_order")
	sendRawCommand("LPUSH lpush_order a b c") // → [c,b,a]
	sendRawCommand("LPUSH lpush_order d e")   // → [e,d,c,b,a]
	got := sendRawCommand("LRANGE lpush_order 0 4")
	want := "*5\r\n$1\r\ne\r\n$1\r\nd\r\n$1\r\nc\r\n$1\r\nb\r\n$1\r\na\r\n"
	if got != want {
		t.Errorf("LPUSH ordering wrong: expected %q, got %q", want, got)
	}
}

func TestLrangeOutOfBounds(t *testing.T) {
	sendRawCommand("DEL oob")
	sendRawCommand("RPUSH oob a b c")
	got := sendRawCommand("LRANGE oob 0 99")
	want := "*3\r\n$1\r\na\r\n$1\r\nb\r\n$1\r\nc\r\n"
	if got != want {
		t.Errorf("expected clamped range %q, got %q", want, got)
	}
}

func TestLrangeNegativeIndex(t *testing.T) {
	sendRawCommand("DEL negrange")
	sendRawCommand("RPUSH negrange a b c")
	got := sendRawCommand("LRANGE negrange 0 -1")
	want := "*3\r\n$1\r\na\r\n$1\r\nb\r\n$1\r\nc\r\n"
	if got != want {
		t.Errorf("expected negative-index range %q, got %q", want, got)
	}
}

func TestErrorInvalidCommand(t *testing.T) {
	s := serializer.Serializer{}
	want := "-ERR invalid command\r\n"
	got := HandleCommand("INVALIDCMD", &s)
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestSetWithExpiry(t *testing.T) {
	resp := sendRawCommand("SET expiring hello PX 10")
	if !strings.HasPrefix(resp, "+OK") {
		t.Errorf("expected OK, got %q", resp)
	}
}
