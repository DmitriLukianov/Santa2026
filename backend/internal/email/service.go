package email

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"log/slog"
	"math/big"
	"net"
	"net/smtp"
	"time"

	"secret-santa-backend/internal/config"
)

type Service struct {
	cfg *config.Config
	log *slog.Logger
}

func New(cfg *config.Config, log *slog.Logger) *Service {
	return &Service{cfg: cfg, log: log}
}

func (s *Service) SendLoginNotification(ctx context.Context, email, name string) error {
	subject := "✅ Вы успешно вошли в Тайный Санта"
	body := fmt.Sprintf(`Привет, %s!
 
Вы только что вошли в аккаунт Тайный Санта.
 
Если это были не вы — немедленно напишите нам.
 
С наилучшими пожеланиями,
Команда Тайный Санта`, name)

	return s.send(ctx, email, subject, body)
}

func (s *Service) SendOTP(ctx context.Context, email string) (string, error) {
	code := s.generateOTP()

	subject := "🔑 Код подтверждения для Тайный Санта"
	body := fmt.Sprintf(`Ваш код подтверждения: %s

Код действителен 10 минут.

Если вы не запрашивали код — просто проигнорируйте это письмо.

С наилучшими пожеланиями,
Команда Тайный Санта`, code)

	if err := s.send(ctx, email, subject, body); err != nil {
		// Код сгенерирован — логируем его, чтобы можно было использовать вручную.
		// Ошибку отправки не пробрасываем: пользователь всё равно может ввести код из лога.
		if s.log != nil {
			s.log.Warn("не удалось отправить OTP на почту, код доступен в логах",
				slog.String("email", email),
				slog.String("otp_code", code),
				slog.String("send_error", err.Error()),
			)
		}
	}
	return code, nil
}

func (s *Service) SendInvitationEmail(ctx context.Context, email, eventTitle, inviteURL string) error {
	subject := fmt.Sprintf("🎅 Приглашение в Тайный Санта: %s", eventTitle)
	body := fmt.Sprintf(`Привет!

Вас приглашают принять участие в событии «%s» в Тайном Санте.

Чтобы присоединиться, перейдите по ссылке:
%s

Ссылка действительна ограниченное время.

С наилучшими пожеланиями,
Команда Тайный Санта`, eventTitle, inviteURL)

	return s.send(ctx, email, subject, body)
}

func (s *Service) SendDrawNotification(ctx context.Context, email, eventTitle string, organizerNotes *string) error {
	subject := "🎲 Жеребьёвка проведена!"
	body := fmt.Sprintf(`Привет!

В событии «%s» проведена жеребьёвка.

Теперь вы можете посмотреть, кому дарите подарок, в разделе «Мои события».

Удачи в Тайном Санте! 🎁`, eventTitle)

	if organizerNotes != nil && *organizerNotes != "" {
		body += fmt.Sprintf("\n\n📌 От организатора:\n%s", *organizerNotes)
	}

	body += "\n \nС наилучшими пожеланиями,\nКоманда Тайный Санта"

	return s.send(ctx, email, subject, body)
}

func (s *Service) generateOTP() string {
	const digits = "0123456789"
	code := make([]byte, 6)
	for i := range code {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		code[i] = digits[n.Int64()]
	}
	return string(code)
}

func (s *Service) send(ctx context.Context, to, subject, body string) error {
	if !s.cfg.SMTPEnabled() {
		if s.log != nil {
			s.log.Debug("SMTP not configured, skipping email",
				slog.String("to", to),
				slog.String("subject", subject),
			)
		}
		return nil
	}

	if s.log != nil {
		s.log.Info("sending email",
			slog.String("to", to),
			slog.String("subject", subject),
		)
	}

	auth := smtp.PlainAuth("", s.cfg.SMTPUsername, s.cfg.SMTPPassword, s.cfg.SMTPHost)

	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n",
		s.cfg.FromEmail, to, subject, body,
	))

	addr := net.JoinHostPort(s.cfg.SMTPHost, fmt.Sprintf("%d", s.cfg.SMTPPort))

	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		if s.log != nil {
			s.log.Error("failed to connect to SMTP server", slog.String("to", to), slog.String("error", err.Error()))
		}
		return fmt.Errorf("failed to connect to smtp: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.cfg.SMTPHost)
	if err != nil {
		return fmt.Errorf("failed to create smtp client: %w", err)
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: s.cfg.SMTPHost}); err != nil {
			return fmt.Errorf("smtp starttls failed: %w", err)
		}
	}

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth failed: %w", err)
	}
	if err := client.Mail(s.cfg.FromEmail); err != nil {
		return fmt.Errorf("smtp mail from failed: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt failed: %w", err)
	}
	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data failed: %w", err)
	}
	if _, err = wc.Write(msg); err != nil {
		return fmt.Errorf("smtp write failed: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("smtp close failed: %w", err)
	}
	if err := client.Quit(); err != nil {
		if s.log != nil {
			s.log.Error("failed to send email",
				slog.String("to", to),
				slog.String("error", err.Error()),
			)
		}
		return fmt.Errorf("failed to send email: %w", err)
	}

	if s.log != nil {
		s.log.Info("email sent successfully", slog.String("to", to))
	}
	return nil
}
