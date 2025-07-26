package main

import (
	"codigo_reto/model"
	"codigo_reto/processor"
	"codigo_reto/smtp"
)

func main() {
	emailChan := make(chan *model.EmailMessage, 100)
	done := make(chan struct{})
	const numWorkers = 4
	go processor.WorkerPool(numWorkers, emailChan, done)

	// Iniciar el servidor SMTP
	if err := smtp.StartSMTPServer(":2525", emailChan); err != nil {
		panic(err)
	}
}
