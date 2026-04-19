package Png

import (
	"encoding/binary"
	"errors"
	"image"
)

const lsbMask = 0b11111110

var messageTooBigError = errors.New("message too big for LSB (use ComputeLSBCapacity)")

func ComputeLSBCapacity(image *image.NRGBA) int {
	return len(image.Pix) / 8
}

func EmbedViaLSB(image *image.NRGBA, message []byte) error {
	if len(message)+4 > len(image.Pix)/8 {
		return messageTooBigError
	}
	// Message length as a big endian uint32
	lenBytes := []byte{
		byte((len(message) >> 24) & 255),
		byte((len(message) >> 16) & 255),
		byte((len(message) >> 8) & 255),
		byte(len(message) & 255),
	}
	// Write the bits into the image.
	for currentBit := 0; currentBit < len(message)*8+32; currentBit++ {
		var bit byte
		if currentBit < 32 { // Writing the first 32 bits, i.e. the message length
			bit = (lenBytes[currentBit>>3] & (1 << (currentBit & 7))) >> (currentBit & 7)
		} else { // Writing the rest of the message
			bit = (message[currentBit>>3-4] & (1 << (currentBit & 7))) >> (currentBit & 7)
		}
		image.Pix[currentBit] = image.Pix[currentBit]&lsbMask | bit
	}
	return nil
}

func ExtractLSB(image *image.NRGBA) []byte {
	var msgBytes []byte
	msgBytes = make([]byte, 4)
	// First, read the message length. It's first 4 bytes (big endian uint32)
	for bitI := 0; bitI < 32; bitI++ {
		msgBytes[bitI>>3] |= (image.Pix[bitI] & 1) << (bitI & 7)
	}
	messageLen := binary.BigEndian.Uint32(msgBytes)
	msgBytes = make([]byte, messageLen)
	for bitI := 0; bitI < int(messageLen*8); bitI++ {
		msgBytes[bitI>>3] |= (image.Pix[bitI+32] & 1) << (bitI & 7)
	}
	return msgBytes
}
