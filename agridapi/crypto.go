package agridapi

import (
	"crypto/aes"
	"crypto/cipher"
)

type gCipher struct {
	key    []byte
	nonce  []byte
	block  cipher.Block
	buffer []byte
}

func (g *gCipher) init(key []byte) error {

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	g.block = block
	g.buffer = make([]byte, aes.BlockSize)
	//g.stream = cipher.NewCFBEncrypter(g.block, g.buffer)
	return nil
}

func (g *gCipher) encrypt(data []byte) ([]byte, error) {

	stream := cipher.NewCFBEncrypter(g.block, g.buffer)
	stream.XORKeyStream(data, data)
	return data, nil
}

func (g *gCipher) decrypt(data []byte) ([]byte, error) {

	stream := cipher.NewCFBDecrypter(g.block, g.buffer)
	stream.XORKeyStream(data, data)
	return data, nil
}
