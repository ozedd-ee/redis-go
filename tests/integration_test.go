// Package tests contains integration tests that run a live server and exercise
// the full request/response path over TCP using the RESP protocol.
package tests

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"redis-go/server"
)

var testAddr string

func TestMain(m *testing.M) {
	srv, err := server.New(":0") // :0 picks a random free port
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start test server: %v\n", err)
		os.Exit(1)
	}
	_, port, _ := net.SplitHostPort(srv.Addr())
	testAddr = "127.0.0.1:" + port

	go srv.Serve()
	code := m.Run()
	srv.Stop()
	os.Exit(code)
}

// redisClient wraps a single TCP connection and knows how to speak RESP.
type redisClient struct {
	conn net.Conn
	r    *bufio.Reader
}

func dial(t *testing.T) *redisClient {
	t.Helper()
	conn, err := net.DialTimeout("tcp", testAddr, 2*time.Second)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return &redisClient{conn: conn, r: bufio.NewReader(conn)}
}

// Do sends a RESP command and returns the raw RESP response string.
func (c *redisClient) Do(args ...string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "*%d\r\n", len(args))
	for _, a := range args {
		fmt.Fprintf(&b, "$%d\r\n%s\r\n", len(a), a)
	}
	c.conn.Write([]byte(b.String()))
	return c.read()
}

// read reads exactly one complete RESP value from the wire.
func (c *redisClient) read() string {
	line, err := c.r.ReadString('\n')
	if err != nil {
		return ""
	}
	switch line[0] {
	case '+', '-', ':':
		return line
	case '$':
		n, _ := strconv.Atoi(strings.TrimRight(line[1:], "\r\n"))
		if n == -1 {
			return "$-1\r\n"
		}
		data := make([]byte, n+2) // data + CRLF
		io.ReadFull(c.r, data)
		return line + string(data[:n]) + "\r\n"
	case '*':
		count, _ := strconv.Atoi(strings.TrimRight(line[1:], "\r\n"))
		if count == 0 {
			return line
		}
		result := line
		for i := 0; i < count; i++ {
			result += c.read()
		}
		return result
	}
	return line
}

// ── Correctness ────────────────────────────────────────────────────────────────

func TestIntegration_Ping(t *testing.T) {
	c := dial(t)
	if got := c.Do("PING"); got != "+PONG\r\n" {
		t.Fatalf("PING: got %q", got)
	}
}

func TestIntegration_Echo(t *testing.T) {
	c := dial(t)
	if got := c.Do("ECHO", "hello"); got != "+hello\r\n" {
		t.Fatalf("ECHO: got %q", got)
	}
}

func TestIntegration_SetGet(t *testing.T) {
	c := dial(t)
	k := t.Name()
	if got := c.Do("SET", k, "world"); got != "+OK\r\n" {
		t.Fatalf("SET: got %q", got)
	}
	if got, want := c.Do("GET", k), "$5\r\nworld\r\n"; got != want {
		t.Fatalf("GET: want %q got %q", want, got)
	}
}

func TestIntegration_GetMissingKey(t *testing.T) {
	c := dial(t)
	if got := c.Do("GET", t.Name()); got != "$-1\r\n" {
		t.Fatalf("GET missing: got %q", got)
	}
}

func TestIntegration_SetExExpiry(t *testing.T) {
	c := dial(t)
	k := t.Name()
	c.Do("SET", k, "val", "PX", "50") // 50 ms TTL
	if got := c.Do("GET", k); got != "$3\r\nval\r\n" {
		t.Fatalf("GET before expiry: got %q", got)
	}
	time.Sleep(80 * time.Millisecond)
	if got := c.Do("GET", k); got != "$-1\r\n" {
		t.Fatalf("GET after expiry: got %q (expected nil)", got)
	}
}

func TestIntegration_ExistsExpiredKey(t *testing.T) {
	c := dial(t)
	k := t.Name()
	c.Do("SET", k, "v", "PX", "30")
	time.Sleep(60 * time.Millisecond)
	if got := c.Do("EXISTS", k); got != ":0\r\n" {
		t.Fatalf("EXISTS on expired key: got %q", got)
	}
}

func TestIntegration_IncrDecrOnNew(t *testing.T) {
	c := dial(t)
	k1, k2 := t.Name()+":i", t.Name()+":d"
	if got := c.Do("INCR", k1); got != ":1\r\n" {
		t.Fatalf("INCR new: got %q", got)
	}
	if got := c.Do("DECR", k2); got != ":-1\r\n" {
		t.Fatalf("DECR new: got %q", got)
	}
}

func TestIntegration_IncrDecrSequence(t *testing.T) {
	c := dial(t)
	k := t.Name()
	c.Do("SET", k, "10")
	c.Do("INCR", k)
	c.Do("INCR", k)
	c.Do("DECR", k)
	if got := c.Do("GET", k); got != "$2\r\n11\r\n" {
		t.Fatalf("after +2-1: got %q", got)
	}
}

func TestIntegration_IncrNegative(t *testing.T) {
	c := dial(t)
	k := t.Name()
	c.Do("SET", k, "-5")
	if got := c.Do("INCR", k); got != ":-4\r\n" {
		t.Fatalf("INCR negative: got %q", got)
	}
}

func TestIntegration_IncrNonInteger(t *testing.T) {
	c := dial(t)
	k := t.Name()
	c.Do("SET", k, "notanint")
	if got := c.Do("INCR", k); !strings.HasPrefix(got, "-ERR") {
		t.Fatalf("INCR non-int: expected error, got %q", got)
	}
}

func TestIntegration_ExistsMultipleKeys(t *testing.T) {
	c := dial(t)
	k1, k2, k3 := t.Name()+":a", t.Name()+":b", t.Name()+":c"
	c.Do("SET", k1, "1")
	c.Do("SET", k2, "2")
	// k3 not set
	if got := c.Do("EXISTS", k1, k2, k3); got != ":2\r\n" {
		t.Fatalf("EXISTS multi: got %q", got)
	}
}

func TestIntegration_DelMultipleKeys(t *testing.T) {
	c := dial(t)
	k1, k2, k3 := t.Name()+":a", t.Name()+":b", t.Name()+":c"
	c.Do("SET", k1, "1")
	c.Do("SET", k2, "2")
	// k3 not set
	if got := c.Do("DEL", k1, k2, k3); got != ":2\r\n" {
		t.Fatalf("DEL multi: got %q", got)
	}
	if got := c.Do("EXISTS", k1); got != ":0\r\n" {
		t.Fatalf("key should be gone: got %q", got)
	}
}

func TestIntegration_LpushOrdering(t *testing.T) {
	c := dial(t)
	k := t.Name()
	// LPUSH list a b c → [c, b, a]
	if got := c.Do("LPUSH", k, "a", "b", "c"); got != ":3\r\n" {
		t.Fatalf("LPUSH new: got %q", got)
	}
	if got, want := c.Do("LRANGE", k, "0", "-1"),
		"*3\r\n$1\r\nc\r\n$1\r\nb\r\n$1\r\na\r\n"; got != want {
		t.Fatalf("LRANGE after LPUSH new: want %q got %q", want, got)
	}
	// LPUSH list d e → [e, d, c, b, a]
	c.Do("LPUSH", k, "d", "e")
	if got, want := c.Do("LRANGE", k, "0", "-1"),
		"*5\r\n$1\r\ne\r\n$1\r\nd\r\n$1\r\nc\r\n$1\r\nb\r\n$1\r\na\r\n"; got != want {
		t.Fatalf("LRANGE after LPUSH existing: want %q got %q", want, got)
	}
}

func TestIntegration_RpushLrange(t *testing.T) {
	c := dial(t)
	k := t.Name()
	c.Do("RPUSH", k, "x", "y", "z")
	if got, want := c.Do("LRANGE", k, "0", "-1"),
		"*3\r\n$1\r\nx\r\n$1\r\ny\r\n$1\r\nz\r\n"; got != want {
		t.Fatalf("RPUSH/LRANGE: want %q got %q", want, got)
	}
}

func TestIntegration_LrangeOutOfBounds(t *testing.T) {
	c := dial(t)
	k := t.Name()
	c.Do("RPUSH", k, "a", "b", "c")
	// end beyond length — should clamp to last element
	if got, want := c.Do("LRANGE", k, "0", "99"),
		"*3\r\n$1\r\na\r\n$1\r\nb\r\n$1\r\nc\r\n"; got != want {
		t.Fatalf("LRANGE OOB: want %q got %q", want, got)
	}
}

func TestIntegration_LrangeEmptyResult(t *testing.T) {
	c := dial(t)
	k := t.Name()
	c.Do("RPUSH", k, "a", "b")
	// begin > end → empty array
	if got := c.Do("LRANGE", k, "5", "2"); got != "*0\r\n" {
		t.Fatalf("LRANGE empty: got %q", got)
	}
}

func TestIntegration_LrangeOnMissingKey(t *testing.T) {
	c := dial(t)
	if got := c.Do("LRANGE", t.Name(), "0", "-1"); got != "*0\r\n" {
		t.Fatalf("LRANGE missing key: got %q", got)
	}
}

func TestIntegration_InvalidCommand(t *testing.T) {
	c := dial(t)
	if got := c.Do("NOSUCHCMD"); !strings.HasPrefix(got, "-ERR") {
		t.Fatalf("invalid command: expected -ERR, got %q", got)
	}
}

func TestIntegration_MissingArgs(t *testing.T) {
	c := dial(t)
	cases := [][]string{
		{"GET"},
		{"SET", "onlykey"},
		{"LPUSH", "onlykey"},
		{"LRANGE", "k", "0"},
	}
	for _, args := range cases {
		got := c.Do(args...)
		if !strings.HasPrefix(got, "-ERR") {
			t.Errorf("args %v: expected -ERR, got %q", args, got)
		}
	}
}

// ── Connection lifecycle ────────────────────────────────────────────────────────

func TestIntegration_ClientDisconnect(t *testing.T) {
	// Server must not crash when a client disconnects mid-session.
	conn, err := net.DialTimeout("tcp", testAddr, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	c := &redisClient{conn: conn, r: bufio.NewReader(conn)}
	c.Do("PING") // establish session
	conn.Close() // abrupt close

	// Small pause then verify server still accepts new connections.
	time.Sleep(20 * time.Millisecond)
	c2 := dial(t)
	if got := c2.Do("PING"); got != "+PONG\r\n" {
		t.Fatalf("server unhealthy after client disconnect: got %q", got)
	}
}

func TestIntegration_Reconnect(t *testing.T) {
	k := t.Name()
	c1 := dial(t)
	c1.Do("SET", k, "persistent")
	c1.conn.Close()

	time.Sleep(10 * time.Millisecond)

	c2 := dial(t)
	if got, want := c2.Do("GET", k), "$10\r\npersistent\r\n"; got != want {
		t.Fatalf("data after reconnect: want %q got %q", want, got)
	}
}

// ── Concurrent access ──────────────────────────────────────────────────────────

func TestIntegration_ConcurrentClients(t *testing.T) {
	const numClients = 20
	const opsPerClient = 50

	var wg sync.WaitGroup
	errs := make(chan string, numClients*opsPerClient)

	for id := 0; id < numClients; id++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			conn, err := net.DialTimeout("tcp", testAddr, 2*time.Second)
			if err != nil {
				errs <- fmt.Sprintf("client %d dial: %v", id, err)
				return
			}
			defer conn.Close()
			c := &redisClient{conn: conn, r: bufio.NewReader(conn)}
			k := fmt.Sprintf("concurrent:%d", id)

			for j := 0; j < opsPerClient; j++ {
				val := strconv.Itoa(j)
				if got := c.Do("SET", k, val); got != "+OK\r\n" {
					errs <- fmt.Sprintf("client %d SET: got %q", id, got)
					return
				}
				got := c.Do("GET", k)
				want := fmt.Sprintf("$%d\r\n%s\r\n", len(val), val)
				if got != want {
					errs <- fmt.Sprintf("client %d GET op %d: want %q got %q", id, j, want, got)
					return
				}
			}
		}(id)
	}

	wg.Wait()
	close(errs)
	for msg := range errs {
		t.Error(msg)
	}
}

func TestIntegration_ConcurrentIncrSameKey(t *testing.T) {
	// All goroutines increment the same counter; final value must equal numGoroutines.
	const numGoroutines = 50
	k := t.Name()

	c0 := dial(t)
	c0.Do("SET", k, "0")

	var wg sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, err := net.DialTimeout("tcp", testAddr, 2*time.Second)
			if err != nil {
				return
			}
			defer conn.Close()
			c := &redisClient{conn: conn, r: bufio.NewReader(conn)}
			c.Do("INCR", k)
		}()
	}
	wg.Wait()

	got := c0.Do("GET", k)
	want := fmt.Sprintf("$%d\r\n%d\r\n", len(strconv.Itoa(numGoroutines)), numGoroutines)
	if got != want {
		t.Fatalf("counter after concurrent INCR: want %q got %q", want, got)
	}
}

// ── Sustained load ─────────────────────────────────────────────────────────────

func TestIntegration_SustainedLoad(t *testing.T) {
	c := dial(t)
	k := t.Name()
	const ops = 500

	for i := 0; i < ops; i++ {
		val := strconv.Itoa(i)
		if got := c.Do("SET", k, val); got != "+OK\r\n" {
			t.Fatalf("op %d SET: got %q", i, got)
		}
		got := c.Do("GET", k)
		want := fmt.Sprintf("$%d\r\n%s\r\n", len(val), val)
		if got != want {
			t.Fatalf("op %d GET: want %q got %q", i, want, got)
		}
	}
}

func TestIntegration_SustainedMixedLoad(t *testing.T) {
	c := dial(t)
	k := t.Name()
	c.Do("SET", k+":ctr", "0")

	const ops = 200
	for i := 0; i < ops; i++ {
		c.Do("INCR", k+":ctr")
		c.Do("SET", k+":val", strconv.Itoa(i))
		c.Do("RPUSH", k+":list", strconv.Itoa(i))
		if got := c.Do("PING"); got != "+PONG\r\n" {
			t.Fatalf("PING failed at op %d: got %q", i, got)
		}
	}

	// Counter must equal ops
	got := c.Do("GET", k+":ctr")
	want := fmt.Sprintf("$%d\r\n%d\r\n", len(strconv.Itoa(ops)), ops)
	if got != want {
		t.Fatalf("sustained counter: want %q got %q", want, got)
	}
}
