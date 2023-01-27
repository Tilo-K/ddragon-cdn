package main

import (
	"fmt"
	"os"
	"time"
)

func checkError(err error) {
	if err == nil {
		return
	}

	fmt.Println("[!] ", err)
	f, _ := os.OpenFile("error.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	v, _ := time.Now().UTC().MarshalText()
	f.WriteString(fmt.Sprintf("[%s] %s\n", string(v), err))
}
