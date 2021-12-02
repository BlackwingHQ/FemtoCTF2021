package main

import (
	"FemtoCTF2021/admin"
	"FemtoCTF2021/secret"
)

func main() {
	go admin.Admin()
	secret.Secret()
}
