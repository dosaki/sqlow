package helpers

import (
	"log"
)

func CheckError(e error) {
	if e != nil {
		log.Fatalln(e.Error())
	}
}

func CheckWarn(e error) {
	if e != nil {
		log.Println(e.Error())
	}
}
