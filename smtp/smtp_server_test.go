package smtp_test

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
	"codigo_reto/model"
	smtp "codigo_reto/smtp"
)

// Arranca el servidor SMTP en background para pruebas.
func startTestSMTPServer(emailChan chan *model.EmailMessage) (addr string, stop func()) {
	ln, err := net.Listen("tcp", ":0") // puerto aleatorio
	if err != nil {
		panic(err)
	}
	addr = ln.Addr().String()
	stopped := make(chan struct{})
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				break
			}
			go smtp.HandleConnection(conn, emailChan)
		}
		close(stopped)
	}()
	return addr, func() { ln.Close(); <-stopped }
}



func TestSMTPFlow_OK(t *testing.T) {
	emailChan := make(chan *model.EmailMessage, 10)
	addr, stop := startTestSMTPServer(emailChan)
	defer stop()
	time.Sleep(100 * time.Millisecond) // Espera a que el server esté listo

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("No se pudo conectar: %v", err)
	}
	defer conn.Close()
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)
	readLine := func() string {
		l, _ := r.ReadString('\n')
		return l
	}
	write := func(s string) {
		fmt.Fprintf(w, "%s\r\n", s)
		w.Flush()
	}
	// Secuencia SMTP válida
	if !strings.HasPrefix(readLine(), "220") {
		t.Fatal("Falta banner 220")
	}
	write("HELO localhost")
	if !strings.HasPrefix(readLine(), "250") {
		t.Fatal("Falta respuesta HELO")
	}
	write("MAIL FROM:<test@dom.com>")
	if !strings.HasPrefix(readLine(), "250") {
		t.Fatal("Falta respuesta MAIL FROM")
	}
	write("RCPT TO:<dest@dom.com>")
	if !strings.HasPrefix(readLine(), "250") {
		t.Fatal("Falta respuesta RCPT TO")
	}
	write("DATA")
	if !strings.HasPrefix(readLine(), "354") {
		t.Fatal("Falta respuesta DATA")
	}
	write("Subject: Prueba\r\n\r\nHola mundo!\r\n.")
	if !strings.HasPrefix(readLine(), "250") {
		t.Fatal("Falta respuesta tras DATA")
	}
	write("QUIT")
	if !strings.HasPrefix(readLine(), "221") {
		t.Fatal("Falta respuesta QUIT")
	}
	// Verifica que el email llegó al canal
	select {
	case email := <-emailChan:
		if email.Subject != "Prueba" || email.Body != "Hola mundo!" {
			t.Errorf("Email parseado incorrectamente: %+v", email)
		}
	case <-time.After(time.Second):
		t.Error("No llegó el email al canal")
	}
}

func TestSMTPFlow_BadSequence(t *testing.T) {
	emailChan := make(chan *model.EmailMessage, 10)
	addr, stop := startTestSMTPServer(emailChan)
	defer stop()
	time.Sleep(100 * time.Millisecond)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("No se pudo conectar: %v", err)
	}
	defer conn.Close()
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)
	readLine := func() string {
		l, _ := r.ReadString('\n')
		return l
	}
	write := func(s string) {
		fmt.Fprintf(w, "%s\r\n", s)
		w.Flush()
	}
	readLine() // 220
	write("DATA")
	resp := readLine()
	if !strings.HasPrefix(resp, "503") {
		t.Errorf("DATA fuera de secuencia no da 503, da: %s", resp)
	}
}

func TestSMTPFlow_Concurrent(t *testing.T) {
	emailChan := make(chan *model.EmailMessage, 100)
	addr, stop := startTestSMTPServer(emailChan)
	defer stop()
	time.Sleep(100 * time.Millisecond)

	n := 10
	done := make(chan struct{})
	for i := 0; i < n; i++ {
		go func(i int) {
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				t.Errorf("No conecta: %v", err)
				return
			}
			defer conn.Close()
			r := bufio.NewReader(conn)
			w := bufio.NewWriter(conn)
			readLine := func() string { l, _ := r.ReadString('\n'); return l }
			write := func(s string) { fmt.Fprintf(w, "%s\r\n", s); w.Flush() }
			readLine() // 220
			write("HELO localhost")
			readLine()
			write(fmt.Sprintf("MAIL FROM:<c%d@dom.com>", i))
			readLine()
			write(fmt.Sprintf("RCPT TO:<d%d@dom.com>", i))
			readLine()
			write("DATA")
			readLine()
			write(fmt.Sprintf("Subject: S%d\r\n\r\nBody%d\r\n.", i, i))
			readLine()
			write("QUIT")
			readLine()
			done <- struct{}{}
		}(i)
	}
	// Espera a que todos terminen
	for i := 0; i < n; i++ {
		<-done
	}
	// Verifica que llegaron todos los emails
	count := 0
	for {
		select {
		case <-emailChan:
			count++
		case <-time.After(500 * time.Millisecond):
			if count != n {
				t.Errorf("Esperaba %d emails, llegaron %d", n, count)
			}
			return
		}
	}
}
