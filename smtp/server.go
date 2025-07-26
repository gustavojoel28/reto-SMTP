package smtp

import (
	"codigo_reto/model"
	"fmt"
	"net"
)


func StartSMTPServer(address string, emailChan chan<- *model.EmailMessage) error {
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("Error al iniciar el listener: %w", err)
	}
	fmt.Printf("Servidor SMTP escuchando en %s\n", address)
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("Error aceptando conexión: %v\n", err)
			continue
		}
		go HandleConnection(conn, emailChan)
	}
}
