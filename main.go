package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

// Функция для получения сообщений
func receiveMessages(conn *net.UDPConn) {
	buffer := make([]byte, 1024)
	for {
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error receiving message:", err)
			continue
		}
		fmt.Printf("Received from %s: %s\n", addr, string(buffer[:n]))
	}
}

// Функция для отправки сообщений
func sendMessages(conn *net.UDPConn, remoteAddr *net.UDPAddr) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter message: ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)

		_, err := conn.WriteToUDP([]byte(text), remoteAddr)
		if err != nil {
			fmt.Println("Error sending message:", err)
		}
	}
}

func main() {
	// Локальный адрес и порт
	localAddr, err := net.ResolveUDPAddr("udp", ":4000")
	if err != nil {
		fmt.Println("Error resolving local address:", err)
		return
	}

	// Удаленный адрес и порт (адрес другого узла)
	remoteAddr, err := net.ResolveUDPAddr("udp", "192.168.1.100:4000") // Замените на IP другого устройства
	if err != nil {
		fmt.Println("Error resolving remote address:", err)
		return
	}

	// Создание UDP соединения
	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		fmt.Println("Error creating UDP connection:", err)
		return
	}
	defer conn.Close()

	fmt.Println("UDP P2P connection established. Type messages to send.")

	// Запуск горутины для получения сообщений
	go receiveMessages(conn)

	// Отправка сообщений
	sendMessages(conn, remoteAddr)
}