package processor

import (
	"codigo_reto/model"
	"fmt"
	"math/rand"
	"net/smtp"
	"strings"
	"time"
)


func WorkerPool(numWorkers int, emailChan <-chan *model.EmailMessage, done chan<- struct{}) {
	for i := 0; i < numWorkers; i++ {
		go worker(i, emailChan, done)
	}
}

func worker(id int, emailChan <-chan *model.EmailMessage, done chan<- struct{}) {
	
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	smtpUser := "edwardmor3h@gmail.com" 
	smtpPass := "toyp nrxf ypgr sncg"   // Genera una contraseña de aplicación en https://myaccount.google.com/apppasswords

	for email := range emailChan {
		fmt.Printf("[Worker %d] Procesando email a %v\n", id, email.To)
		
		delay := rand.Intn(1000) + 500 // 500ms a 1500ms
		time.Sleep(time.Duration(delay) * time.Millisecond)
		fmt.Printf("[Worker %d] --- EMAIL RECIBIDO ---\n", id)
		fmt.Printf("Para: %v\nAsunto: %s\nHeaders: %v\nCuerpo:\n%s\n", email.To, email.Subject, email.Headers, email.Body)
		fmt.Printf("[Worker %d] --- FIN EMAIL ---\n", id)

		toHeader := strings.Join(email.To, ",")
		msg := "From: " + smtpUser + "\r\n" +
			"To: " + toHeader + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/plain; charset=UTF-8\r\n"
		for k, v := range email.Headers {
			if strings.ToLower(k) != "subject" {
				msg += k + ": " + v + "\r\n"
			}
		}
		msg += "Subject: " + email.Subject + "\r\n"
		msg += "\r\n" + email.Body

		auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
		err := smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpUser, email.To, []byte(msg))
		if err != nil {
			fmt.Printf("[Worker %d] Error al enviar: %v\n", id, err)
		} else {
			fmt.Printf("[Worker %d] Email enviado correctamente!\n", id)
		}
		done <- struct{}{}
	}
}
