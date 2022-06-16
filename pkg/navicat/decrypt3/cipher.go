package decrypt3

import (
	"encoding/hex"
	"log"

	"github.com/andreburgaud/crypt2go/ecb"
	"golang.org/x/crypto/blowfish"
)

var (
	pt  []byte = []byte("\xFF\xFF\xFF\xFF\xFF\xFF\xFF\xFF")
	key []byte = []byte("B\xce\xb2q\xa5\xe4X\xb7J\xea\x93\x94y\"5C\x91\x873@")
	cv  []byte = encrypt(pt, key)
)

func encrypt(pt, key []byte) []byte {
	ct := make([]byte, len(pt))
	block, err := blowfish.NewCipher(key)
	if err != nil {
		log.Println(err)
		return ct
	}
	mode := ecb.NewECBEncrypter(block)
	/*padder := padding.NewPkcs5Padding()
	pt, err = padder.Pad(pt) // pad last block of plaintext if block size less than block cipher size
	if err != nil {
		panic(err.Error())
	}*/
	mode.CryptBlocks(ct, pt)
	return ct
}

func decrypt(ct, key []byte) []byte {
	pt := make([]byte, len(ct))
	block, err := blowfish.NewCipher(key)
	if err != nil {
		log.Println(err)
		return pt
	}
	mode := ecb.NewECBDecrypter(block)
	mode.CryptBlocks(pt, ct)
	/*padder := padding.NewPkcs5Padding()
	pt, err = padder.Unpad(pt) // unpad plaintext after decryption
	if err != nil {
		panic(err.Error())
	}*/
	return pt
}

func XorBytes(b ...[]byte) []byte {
	b_len := len(b[0])
	br := make([]byte, b_len)
	for _, m := range b {
		if len(m) != b_len {
			log.Println("XorBytes length mismatch!")
			return br
		}
	}
	for i := range b[0] {
		br[i] = 0
		for _, m := range b {
			br[i] = br[i] ^ m[i]
		}
	}
	return br
}

func EncryptString(data string) string {
	/*	fmt.Printf("Ciphertext: %x\n", a)
		fmt.Println(hex.EncodeToString(a))
		fmt.Println(hex.EncodeToString(XorBytes(a,[]byte("abcdefgh"))))*/
	full_round := len(data) / 8
	left_length := len(data) % 8
	plaintext := string("")
	cv1 := cv
	for i := 0; i < full_round; i++ {
		//fmt.Println(data[i*8:i*8+8])
		r := XorBytes([]byte(data[i*8:i*8+8]), cv1)
		r = encrypt(r, key)
		plaintext += hex.EncodeToString(r)
		cv1 = XorBytes(cv1, r)
		//fmt.Println(plaintext)
	}
	if left_length != 0 {
		cv1 = encrypt(cv1, key)
		plaintext += hex.EncodeToString(XorBytes([]byte(data[8*full_round:]), cv1[:left_length]))
		//fmt.Println(plaintext)
	}
	return plaintext
}

func DecryptString(data1 string) string {
	data, _ := hex.DecodeString(data1)
	cv1 := cv
	full_round := len(data) / 8
	left_length := len(data) % 8
	plaintext := string("")
	for i := 0; i < full_round; i++ {
		r := decrypt(data[i*8:i*8+8], key)
		r = XorBytes(r, cv1)
		plaintext += hex.EncodeToString(r)
		cv1 = XorBytes(cv1, data[i*8:i*8+8])
	}
	if left_length != 0 {
		cv1 = encrypt(cv1, key)
		plaintext += hex.EncodeToString(XorBytes([]byte(data[8*full_round:]), cv1[:left_length]))
	}
	result, _ := hex.DecodeString(plaintext)
	return string(result)
}
