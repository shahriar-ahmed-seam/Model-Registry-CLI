package main

import (
	"os"
	
	"ml-reg/internal/cmd"
)

func main() {
	os.Exit(cmd.Execute())
}