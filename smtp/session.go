package smtp

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"codigo_reto/model"
)

// SMTPSession representa el estado de una sesión SMTP por conexión.
type SMTPSession struct {
	Conn     net.Conn
	Reader   *bufio.Reader
	Writer   *bufio.Writer
	Helo     bool
	MailFrom string
	RcptTo   []string
	Headers  map[string]string
	Data     bool
}

// handleConnection maneja el ciclo de vida de una sesión SMTP.
func HandleConnection(conn net.Conn, emailChan chan<- *model.EmailMessage) {
	defer conn.Close()
	
	sess := &SMTPSession{
		Conn:   conn,
		Reader: bufio.NewReader(conn),
		Writer: bufio.NewWriter(conn),
		Headers: make(map[string]string),
	}
	sess.writeResponse(220, "Simple Go SMTP Service Ready")

	for {
		line, err := sess.Reader.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimRight(line, "\r\n")
		if len(line) == 0 {
			sess.writeResponse(500, "Empty command")
			continue
		}
		cmd, arg := parseCmd(line)
		switch strings.ToUpper(cmd) {
		case "HELO", "EHLO":
			sess.Helo = true
			sess.writeResponse(250, "Hello")
		case "MAIL":
			if !sess.Helo {
				sess.writeResponse(503, "Bad sequence of commands: HELO/EHLO required")
				continue
			}
			from := parseMailFrom(arg)
			if from == "" {
				sess.writeResponse(501, "Syntax error in parameters or arguments")
				continue
			}
			sess.MailFrom = from
			sess.RcptTo = nil // Reset recipients for new mail
			sess.writeResponse(250, "OK")
		case "RCPT":
			if sess.MailFrom == "" {
				sess.writeResponse(503, "Bad sequence of commands: MAIL FROM required")
				continue
			}
			to := parseRcptTo(arg)
			if to == "" {
				sess.writeResponse(501, "Syntax error in parameters or arguments")
				continue
			}
			sess.RcptTo = append(sess.RcptTo, to)
			sess.writeResponse(250, "OK")
		case "DATA":
			if len(sess.RcptTo) == 0 {
				sess.writeResponse(503, "Bad sequence of commands: RCPT TO required")
				continue
			}
			sess.writeResponse(354, "End data with <CR><LF>.<CR><LF>")
			body, headers, subj, err := readData(sess.Reader)
			if err != nil {
				sess.writeResponse(451, "Error reading DATA")
				continue
			}
			email := &model.EmailMessage{
				From:    sess.MailFrom,
				To:      sess.RcptTo,
				Subject: subj,
				Headers: headers,
				Body:    body,
			}
			emailChan <- email
			sess.writeResponse(250, "OK: Queued")
			// Reset for next message
			sess.MailFrom = ""
			sess.RcptTo = nil
		case "QUIT":
			sess.writeResponse(221, "Bye")
			return
		default:
			sess.writeResponse(500, "Command unrecognized")
		}
	}
}

func (s *SMTPSession) writeResponse(code int, msg string) {
	fmt.Fprintf(s.Writer, "%d %s\r\n", code, msg)
	s.Writer.Flush()
}

func parseCmd(line string) (cmd, arg string) {
	parts := strings.SplitN(line, " ", 2)
	cmd = parts[0]
	if len(parts) > 1 {
		arg = parts[1]
	}
	return
}

func parseMailFrom(arg string) string {
	if strings.HasPrefix(strings.ToUpper(arg), "FROM:") {
		addr := strings.TrimSpace(arg[5:])
		return strings.Trim(addr, "<>")
	}
	return ""
}

func parseRcptTo(arg string) string {
	if strings.HasPrefix(strings.ToUpper(arg), "TO:") {
		addr := strings.TrimSpace(arg[3:])
		return strings.Trim(addr, "<>")
	}
	return ""
}

// readData lee el contenido del correo hasta una línea con solo un punto.
func readData(r *bufio.Reader) (body string, headers map[string]string, subject string, err error) {
	headers = make(map[string]string)
	var lines []string
	readingHeaders := true
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return "", nil, "", err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "." {
			break
		}
		if readingHeaders && line == "" {
			readingHeaders = false
			continue
		}
		if readingHeaders {
			if idx := strings.Index(line, ":"); idx != -1 {
				key := strings.TrimSpace(line[:idx])
				val := strings.TrimSpace(line[idx+1:])
				headers[key] = val
				if strings.ToLower(key) == "subject" {
					subject = val
				}
			}
			continue
		}
		lines = append(lines, line)
	}
	body = strings.Join(lines, "\n")
	return
}
