package serializer_test

import (
	"testing"

	"redis-go/serializer"
)

func BenchmarkSerializeArray(b *testing.B) {
	s := &serializer.Serializer{}
	data := []string{"SET", "mykey", "myvalue"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.SerializeArray(data...)
	}
}

func BenchmarkSerializeArrayParallel(b *testing.B) {
	s := &serializer.Serializer{}
	data := []string{"SET", "mykey", "myvalue"}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = s.SerializeArray(data...)
		}
	})
}

func BenchmarkDeserializeMessage(b *testing.B) {
	s := &serializer.Serializer{}
	// RESP array of 3 bulk strings
	input := "*3\r\n$3\r\nGET\r\n$5\r\nmykey\r\n$7\r\nmyvalue\r\n"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = s.DeserializeMessage(input)
	}
}

func BenchmarkDeserializeMessageParallel(b *testing.B) {
	s := &serializer.Serializer{}
	input := "*3\r\n$3\r\nGET\r\n$5\r\nmykey\r\n$7\r\nmyvalue\r\n"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = s.DeserializeMessage(input)
		}
	})
}
