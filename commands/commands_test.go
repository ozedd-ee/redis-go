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

func TestIncrDecr(t *testing.T) {
	sendRawCommand("SET count 10")
	got := sendRawCommand("INCR count")
	if !strings.HasPrefix(got, ":11") {
		t.Errorf("expected incremented integer, got %q", got)
	}

	got = sendRawCommand("DECR count")
	if !strings.HasPrefix(got, ":10") {
		t.Errorf("expected decremented integer, got %q", got)
	}
}

func TestExistsDel(t *testing.T) {
	sendRawCommand("SET x val")
	got := sendRawCommand("EXISTS x")
	if !strings.HasPrefix(got, ":1") {
		t.Errorf("expected EXISTS 1, got %q", got)
	}

	sendRawCommand("DEL x")
	got = sendRawCommand("EXISTS x")
	if !strings.HasPrefix(got, ":0") {
		t.Errorf("expected EXISTS 0 after delete, got %q", got)
	}
}

func TestLpushRpushLrange(t *testing.T) {
	sendRawCommand("LPUSH list one two")
	sendRawCommand("RPUSH list three")

	got := sendRawCommand("LRANGE list 0 2")
	want := "*3\r\n$3\r\ntwo\r\n$3\r\none\r\n$5\r\nthree\r\n"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
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
