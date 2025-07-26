package processor

import (
	"fmt"
	"math/rand"
	"time"
	"codigo_reto/model"
)

// WorkerPool inicia un grupo de workers para procesar emails concurrentemente.
func WorkerPool(numWorkers int, emailChan <-chan *model.EmailMessage, done chan<- struct{}) {
	for i := 0; i < numWorkers; i++ {
		go worker(i, emailChan, done)
	}
}

func worker(id int, emailChan <-chan *model.EmailMessage, done chan<- struct{}) {
	for email := range emailChan {
		fmt.Printf("[Worker %d] Procesando email de %s a %v\n", id, email.From, email.To)
		// Simula procesamiento variable
		delay := rand.Intn(1000) + 500 // 500ms a 1500ms
		time.Sleep(time.Duration(delay) * time.Millisecond)
		fmt.Printf("[Worker %d] Procesado: Subject='%s'\n", id, email.Subject)
		done <- struct{}{}
	}
}
