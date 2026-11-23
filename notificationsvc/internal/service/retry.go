package service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/textproto"
	"time"
)

const (
	maxAttempts  = 3
	initialDelay = time.Second
)

func (s *NotificationService) sendWithRetry(
	ctx context.Context,
	to, subject, body string,
) error {
	delay := initialDelay

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		err := s.repo.Send(ctx, to, subject, body)
		if err == nil {
			return nil
		}

		if !isRetryable(err) || attempt == maxAttempts {
			return fmt.Errorf(
				"send email after %d attempt(s): %w",
				attempt,
				err,
			)
		}

		timer := time.NewTimer(delay)

		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()

		case <-timer.C:
		}

		delay *= 2
	}

	return nil
}

func isRetryable(err error) bool {
	if errors.Is(err, context.Canceled) ||
		errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	var smtpErr *textproto.Error
	if errors.As(err, &smtpErr) {
		return smtpErr.Code >= 400 && smtpErr.Code < 500
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout() || netErr.Temporary()
	}

	return errors.Is(err, net.ErrClosed)
}
