package commands

import (
	"testing"

	"redis-go/serializer"
)

func BenchmarkSetCommand(b *testing.B) {
	s := &serializer.Serializer{}
	resp := s.SerializeArray("SET", "benchkey", "42")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = HandleCommand(mustDeserialize(s, resp), s)
	}
}

func BenchmarkGetCommand(b *testing.B) {
	sendRawCommand("SET benchkey 42")
	s := &serializer.Serializer{}
	resp := s.SerializeArray("GET", "benchkey")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = HandleCommand(mustDeserialize(s, resp), s)
	}
}

func BenchmarkIncrCommand(b *testing.B) {
	sendRawCommand("SET counter 0")
	s := &serializer.Serializer{}
	resp := s.SerializeArray("INCR", "counter")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = HandleCommand(mustDeserialize(s, resp), s)
	}
}

func BenchmarkLpushCommand(b *testing.B) {
	s := &serializer.Serializer{}
	resp := s.SerializeArray("LPUSH", "mylist", "a", "b", "c")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = HandleCommand(mustDeserialize(s, resp), s)
	}
}
