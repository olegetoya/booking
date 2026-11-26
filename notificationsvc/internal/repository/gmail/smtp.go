package gmail

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Repository struct {
	host     string
	port     string
	username string
	password string
}

func NewRepository(
	host string,
	port string,
	username string,
	password string,
) *Repository {
	return &Repository{
		host:     host,
		port:     port,
		username: username,
		password: password,
	}
}

func (r *Repository) Send(
	ctx context.Context,
	to string,
	subject string,
	body string,
	attempt int,
) (resultErr error) {
	const op = "gmail.Repository.Send"

	ctx, span := otel.Tracer("notificationsvc/smtp").Start(ctx, "SMTP send",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attribute.Int("smtp.attempt", attempt)),
	)
	defer func() {
		if resultErr != nil {
			var reply *textproto.Error
			if errors.As(resultErr, &reply) {
				span.SetAttributes(attribute.Int("smtp.response.code", reply.Code))
			}
			span.RecordError(errors.New("smtp attempt failed"))
			span.SetStatus(codes.Error, "smtp attempt failed")
		}
		span.End()
	}()

	if ctx.Err() != nil {
		return ctx.Err()
	}

	recipient, err := mail.ParseAddress(to)
	if err != nil {
		return fmt.Errorf("%s: invalid recipient: %w", op, err)
	}

	if recipient.Address != to {
		return fmt.Errorf("%s: recipient must be a plain email address", op)
	}

	addr := net.JoinHostPort(r.host, r.port)

	dialer := net.Dialer{
		Timeout: 10 * time.Second,
	}

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("%s: connect smtp: %w", op, err)
	}

	defer conn.Close()
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stop()

	deadline := time.Now().Add(30 * time.Second)

	if ctxDeadline, ok := ctx.Deadline(); ok &&
		ctxDeadline.Before(deadline) {
		deadline = ctxDeadline
	}

	if err := conn.SetDeadline(deadline); err != nil {
		return fmt.Errorf("%s: set deadline: %w", op, err)
	}

	client, err := smtp.NewClient(conn, r.host)
	if err != nil {
		return fmt.Errorf("%s: create smtp client: %w", op, err)
	}

	if ok, _ := client.Extension("STARTTLS"); !ok {
		return fmt.Errorf("%s: server does not support STARTTLS", op)
	}

	tlsConfig := &tls.Config{
		ServerName: r.host,
		MinVersion: tls.VersionTLS12,
	}

	if err := client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("%s: start tls: %w", op, err)
	}

	auth := smtp.PlainAuth(
		"",
		r.username,
		r.password,
		r.host,
	)

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("%s: authenticate: %w", op, err)
	}

	if err := client.Mail(r.username); err != nil {
		return fmt.Errorf("%s: set sender: %w", op, err)
	}

	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("%s: set recipient: %w", op, err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("%s: open message: %w", op, err)
	}

	message := buildMessage(r.username, to, subject, body)

	if _, err := writer.Write([]byte(message)); err != nil {
		_ = writer.Close()
		return fmt.Errorf("%s: write message: %w", op, err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("%s: finish message: %w", op, err)
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	if err := client.Quit(); err != nil {
		return fmt.Errorf("%s: quit smtp: %w", op, err)
	}

	return nil
}

func buildMessage(from, to, subject, body string) string {
	subject = strings.ReplaceAll(subject, "\r", "")
	subject = strings.ReplaceAll(subject, "\n", "")

	body = strings.ReplaceAll(body, "\r\n", "\n")
	body = strings.ReplaceAll(body, "\r", "\n")
	body = strings.ReplaceAll(body, "\n", "\r\n")

	return fmt.Sprintf(
		"From: <%s>\r\n"+
			"To: <%s>\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/plain; charset=UTF-8\r\n"+
			"Content-Transfer-Encoding: 8bit\r\n"+
			"\r\n"+
			"%s\r\n",
		from,
		to,
		mime.QEncoding.Encode("utf-8", subject),
		body,
	)
}
