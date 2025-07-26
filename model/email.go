package model

// EmailMessage representa un email recibido por el servidor SMTP.
type EmailMessage struct {
	From    string
	To      []string
	Subject string
	Headers map[string]string
	Body    string
}
