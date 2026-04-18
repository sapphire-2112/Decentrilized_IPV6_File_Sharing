package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	// Open file
	file, err := os.Open("sample.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Connect to receiver IPv6 address
	conn, err := net.Dial("tcp6", "[IPV6]:4242")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Get file metadata
	fileInfo, err := file.Stat()
	if err != nil {
		panic(err)
	}

	fileName := fileInfo.Name()
	fileSize := fileInfo.Size()

	fmt.Println("Sending:", fileName)
	fmt.Println("Size:", fileSize, "bytes")

	// Send filename length
	nameBytes := []byte(fileName)
	err = binary.Write(conn, binary.BigEndian, uint16(len(nameBytes)))
	if err != nil {
		panic(err)
	}

	// Send filename
	_, err = conn.Write(nameBytes)
	if err != nil {
		panic(err)
	}

	// Send file size
	err = binary.Write(conn, binary.BigEndian, uint64(fileSize))
	if err != nil {
		panic(err)
	}

	// Send file in chunks
	buffer := make([]byte, 1024*1024) // 1 MB chunk size

	for {
		n, err := file.Read(buffer)

		if err != nil && err != io.EOF {
			panic(err)
		}

		if n == 0 {
			break
		}

		data := buffer[:n]

		// Calculate checksum
		checksum := sha256.Sum256(data)

		// Send chunk size
		err = binary.Write(conn, binary.BigEndian, uint32(n))
		if err != nil {
			panic(err)
		}

		// Send checksum
		_, err = conn.Write(checksum[:])
		if err != nil {
			panic(err)
		}

		// Send actual data
		_, err = conn.Write(data)
		if err != nil {
			panic(err)
		}

		fmt.Println("Sent chunk:", n, "bytes")
	}

	// Send 0-sized chunk to indicate transfer complete
	err = binary.Write(conn, binary.BigEndian, uint32(0))
	if err != nil {
		panic(err)
	}

	fmt.Println("File sent successfully")
}