package mailer

import (
	"bufio"
	"context"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/opsybot/opsybot/internal/config"
)

// fakeSMTP accepts one plaintext SMTP session and captures the DATA payload.
func fakeSMTP(t *testing.T) (host string, port int, data func() string) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	var mu sync.Mutex
	var captured string
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		br := bufio.NewReader(conn)
		writeln := func(s string) { _, _ = conn.Write([]byte(s + "\r\n")) }
		writeln("220 fake ESMTP")
		inData := false
		var buf strings.Builder
		for {
			line, err := br.ReadString('\n')
			if err != nil {
				return
			}
			if inData {
				if line == ".\r\n" {
					inData = false
					mu.Lock()
					captured = buf.String()
					mu.Unlock()
					writeln("250 OK")
					continue
				}
				buf.WriteString(line)
				continue
			}
			switch cmd := strings.ToUpper(strings.TrimSpace(line)); {
			case strings.HasPrefix(cmd, "EHLO"), strings.HasPrefix(cmd, "HELO"):
				writeln("250-fake.localhost")
				writeln("250 OK")
			case cmd == "DATA":
				writeln("354 End data with <CR><LF>.<CR><LF>")
				inData = true
			case cmd == "QUIT":
				writeln("221 Bye")
				return
			default:
				writeln("250 OK")
			}
		}
	}()

	addr := ln.Addr().(*net.TCPAddr)
	return "127.0.0.1", addr.Port, func() string {
		mu.Lock()
		defer mu.Unlock()
		return captured
	}
}

func samplePage() PageData {
	return PageData{
		Severity: "critical", Service: "payments", Title: "checkout is down",
		StartedAt: "2026-07-23 12:00", PolicySlug: "prod", Level: 1,
		AlertURL: "https://opsy.test/acme/alerts/al-1",
		AckURL:   "https://opsy.test/v1/act/atk", ResolveURL: "https://opsy.test/v1/act/rtk",
	}
}

func TestSendPageRendersSubjectBodyAndActionLinks(t *testing.T) {
	host, port, data := fakeSMTP(t)
	c, err := New(config.Mailer{Host: host, Port: port, Encryption: "none", From: "opsy@opsy.test", FromName: "Opsybot", Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("new mailer: %v", err)
	}
	if err := c.SendPage(context.Background(), "on-call@acme.test", samplePage()); err != nil {
		t.Fatalf("send page: %v", err)
	}
	got := data()
	for _, want := range []string{
		"[CRITICAL] payments: checkout is down",
		"checkout is down",
		"https://opsy.test/v1/act/atk",
		"https://opsy.test/v1/act/rtk",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("sent message is missing %q\n---\n%s", want, got)
		}
	}
}

func TestSendPageDisabledWithoutHost(t *testing.T) {
	c, err := New(config.Mailer{})
	if err != nil {
		t.Fatalf("new mailer: %v", err)
	}
	if err := c.SendPage(context.Background(), "on-call@acme.test", samplePage()); err != ErrDisabled {
		t.Fatalf("a mailer with no host should return ErrDisabled, got %v", err)
	}
}
