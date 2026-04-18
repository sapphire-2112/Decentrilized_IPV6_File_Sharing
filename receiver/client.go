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
	// Listen on IPv6 TCP
	ln, err := net.Listen("tcp6", "[::]:4242")
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	fmt.Println("Waiting for sender on port 4242...")

	conn, err := ln.Accept()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Println("Sender connected from:", conn.RemoteAddr())

	// ---------- Read metadata ----------

	// Read filename length
	var nameLen uint16
	err = binary.Read(conn, binary.BigEndian, &nameLen)
	if err != nil {
		panic(err)
	}

	// Read filename
	nameBytes := make([]byte, nameLen)
	_, err = io.ReadFull(conn, nameBytes)
	if err != nil {
		panic(err)
	}

	fileName := string(nameBytes)

	// Read file size
	var fileSize uint64
	err = binary.Read(conn, binary.BigEndian, &fileSize)
	if err != nil {
		panic(err)
	}

	fmt.Println("Receiving file:", fileName)
	fmt.Println("Expected size:", fileSize, "bytes")

	// Create output file
	outFile, err := os.Create("received_" + fileName)
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	var received uint64

	// ---------- Receive chunks ----------
	for {
		// Read chunk size
		var chunkSize uint32
		err = binary.Read(conn, binary.BigEndian, &chunkSize)
		if err != nil {
			panic(err)
		}

		// 0 means transfer finished
		if chunkSize == 0 {
			break
		}

		// Read checksum (32 bytes)
		checksum := make([]byte, 32)
		_, err = io.ReadFull(conn, checksum)
		if err != nil {
			panic(err)
		}

		// Read actual data
		data := make([]byte, chunkSize)
		_, err = io.ReadFull(conn, data)
		if err != nil {
			panic(err)
		}

		// Verify checksum
		calculated := sha256.Sum256(data)

		if string(calculated[:]) != string(checksum) {
			fmt.Println("Checksum mismatch! Chunk corrupted.")
			return
		}

		// Write chunk to file
		_, err = outFile.Write(data)
		if err != nil {
			panic(err)
		}

		received += uint64(chunkSize)
		fmt.Printf("Received %d / %d bytes\n", received, fileSize)
	}

	fmt.Println("File received successfully as:", "received_"+fileName)
}
