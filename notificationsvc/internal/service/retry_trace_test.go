package service

import (
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/olegetoya/booking/notificationsvc/internal/events"
	"github.com/olegetoya/booking/notificationsvc/internal/repository/gmail"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestSMTPRetrySpans(t *testing.T) {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	replies := []int{451, 451, 550}
	serverErr := make(chan error, 1)
	go func() {
		for _, code := range replies {
			conn, err := listener.Accept()
			if err != nil {
				serverErr <- err
				return
			}
			_, err = io.WriteString(conn, fmt.Sprintf("%d recipient@example.test\r\n", code))
			_ = conn.Close()
			if err != nil {
				serverErr <- err
				return
			}
		}
		serverErr <- nil
	}()

	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	previous := otel.GetTracerProvider()
	otel.SetTracerProvider(provider)
	t.Cleanup(func() {
		otel.SetTracerProvider(previous)
		_ = provider.Shutdown(context.Background())
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ctx, processing := provider.Tracer("test").Start(ctx, "event processing")
	defer processing.End()

	host, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	repo := gmail.NewRepository(host, port, "sender@example.test", "test")
	service := NewNotificationService(repo, "recipient@example.test")
	err = service.BookingCreated(ctx, events.BookingCreated{BookingID: 1})
	if err == nil || !strings.Contains(err.Error(), "after 3 attempt(s)") {
		t.Fatalf("expected three failed attempts, got %v", err)
	}
	select {
	case err := <-serverErr:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}

	spans := recorder.Ended()
	if len(spans) != 4 {
		t.Fatalf("got %d ended spans, want 3 SMTP attempts and 1 notification", len(spans))
	}
	notification := spans[3]
	if notification.Name() != "booking.created notify" ||
		notification.Parent().SpanID() != processing.SpanContext().SpanID() ||
		notification.Status().Code != codes.Error {
		t.Fatalf("unexpected notification span: %s", notification.Name())
	}
	for i, span := range spans[:3] {
		if span.Name() != "SMTP send" || span.Parent().SpanID() != notification.SpanContext().SpanID() ||
			span.SpanContext().TraceID() != processing.SpanContext().TraceID() || span.Status().Code != codes.Error {
			t.Fatalf("invalid span for SMTP attempt %d", i+1)
		}

		attrs := make(map[string]int64)
		for _, attr := range span.Attributes() {
			attrs[string(attr.Key)] = attr.Value.AsInt64()
		}
		if attrs["smtp.attempt"] != int64(i+1) || attrs["smtp.response.code"] != int64(replies[i]) {
			t.Fatalf("invalid attributes for SMTP attempt %d: %v", i+1, attrs)
		}
		events := span.Events()
		if len(events) != 1 {
			t.Fatalf("got %d events for SMTP attempt %d, want one error event", len(events), i+1)
		}
		for _, attr := range events[0].Attributes {
			if strings.Contains(attr.Value.AsString(), "recipient@example.test") {
				t.Fatalf("SMTP error event contains the recipient")
			}
		}
	}
}
