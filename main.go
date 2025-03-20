package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"os"
	"strings"

	"github.com/gordonklaus/portaudio"
)

func main() {
	// Локальный адрес и порт
	localAddr, err := net.ResolveUDPAddr("udp", ":4000")
	if err != nil {
		fmt.Println("Error resolving local address:", err)
		return
	}

	// Удаленный адрес и порт (адрес другого узла)
	remoteAddr, err := net.ResolveUDPAddr("udp", "192.168.0.135:4000") // Замените на IP другого устройства
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

	sendAudio(conn, remoteAddr)
	// receiveAudio(conn, remoteAddr)

	// Запуск горутины для получения сообщений
	// go receiveMessages(conn)

	// Отправка сообщений
	// sendMessages(conn, remoteAddr)
}

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

func sendAudio(conn *net.UDPConn, remoteAddr *net.UDPAddr) {
	// Инициализация PortAudio
	err := portaudio.Initialize()
	if err != nil {
		fmt.Println("Ошибка инициализации PortAudio:", err)
		return
	}
	defer portaudio.Terminate()

	// Параметры аудио
	sampleRate := 44100
	framesPerBuffer := 1024
	numChannels := 1

	// Открытие потока для захвата звука
	streamIn, err := portaudio.OpenDefaultStream(numChannels, 0, float64(sampleRate), framesPerBuffer, func(in []float32) {
		for {
			buf := make([]byte, len(in)*4)
			for i := range in {
				bits := math.Float32bits(in[i])
				binary.LittleEndian.PutUint32(buf[i*4:], bits)
			}
			_, err := conn.WriteToUDP(buf, remoteAddr)
			if err != nil {
				fmt.Println("Error sending voice:", err)
			}
		}
	})
	if err != nil {
		fmt.Println("Ошибка при открытии входного потока:", err)
		return
	}
	defer streamIn.Close()

	// Запуск потока захвата
	fmt.Println("Запись начата. Нажмите Enter для завершения записи...")
	if err := streamIn.Start(); err != nil {
		fmt.Println("Ошибка при запуске входного потока:", err)
		return
	}
	defer streamIn.Stop()

	// Ожидание завершения записи (нажатие Enter)
	waitForEnter()

	// Остановка записи
	if err := streamIn.Stop(); err != nil {
		fmt.Println("Ошибка при остановке входного потока:", err)
		return
	}
}

func receiveAudio(conn *net.UDPConn, remoteAddr *net.UDPAddr) {
	// Инициализация PortAudio
	err := portaudio.Initialize()
	if err != nil {
		fmt.Println("Ошибка инициализации PortAudio:", err)
		return
	}
	defer portaudio.Terminate()

	// Параметры аудио
	sampleRate := 44100
	framesPerBuffer := 1024
	numChannels := 1

	// Открытие потока для воспроизведения звука
	streamOut, err := portaudio.OpenDefaultStream(0, numChannels, float64(sampleRate), framesPerBuffer, func(out []float32) {
		buffer := make([]byte, 1024)
		for {
			n, _, err := conn.ReadFromUDP(buffer)
			if err != nil {
				fmt.Println("Error receiving message:", err)
				continue
			}
			var audioBuffer []float32
			for i := 0; i < n; i += 4 {
				bits := binary.LittleEndian.Uint32(buffer[i:])
				audioBuffer = append(audioBuffer, math.Float32frombits(bits))
			}
			// Воспроизведение аудиоданных из буфера
			for i := range out {
				if len(audioBuffer) > 0 {
					out[i] = audioBuffer[0]
					audioBuffer = audioBuffer[1:] // Удаляем воспроизведенный сэмпл из буфера
				} else {
					out[i] = 0 // Если буфер пуст, воспроизводим тишину
				}
			}
		}
	})
	if err != nil {
		fmt.Println("Ошибка при открытии выходного потока:", err)
		return
	}
	defer streamOut.Close()

	// Запуск потока воспроизведения
	if err := streamOut.Start(); err != nil {
		fmt.Println("Ошибка при запуске выходного потока:", err)
		return
	}
	defer streamOut.Stop()

	fmt.Println("Воспроизведение начато. Нажмите Enter для завершения...")
	waitForEnter()
}

// Функция для ожидания нажатия Enter
func waitForEnter() {
	fmt.Scanln() // Ожидание нажатия Enter
}
