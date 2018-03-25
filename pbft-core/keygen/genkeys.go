package main

import (
	"fmt"
	"io/ioutil"
	"crypto/ecdsa"
	"path/filepath"
	"crypto/elliptic"
	"path"
	"os"
	"log"
	"crypto/rand"
)

func GetCWD() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return dir
}

func write_new_keys(kcount int) {
	for k:= 0; k < kcount; k++ {
		sk, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		filename := fmt.Sprintf("sign%v.dat", k)
		// FIXME: glob (??)
		err := ioutil.WriteFile(path.Join(GetCWD(), "keys/", filename), sk.D.Bytes(), 0644)
		if err != nil {
			log.Fatal(err)
		}
		
	}
}

func main(){
	N := 1000
	KD := "keys/"
	_, err := os.Stat(KD)
        if os.IsNotExist(err) {
                err := os.Mkdir(KD, 0777)
                if err != nil {
                        log.Fatalln(err)
                }
        }

	write_new_keys(N)
	fmt.Printf("Generated %d keys in keys/ folder..\n", N)
}
