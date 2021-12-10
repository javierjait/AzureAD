package util

import (
	"log"
	"os"
)

// AddToLog Add message to log
func AddToLog(tipo string, msg interface{}) {
	f, err := os.OpenFile("archivo.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644) // Open or create file to write
	if err == nil { // if no err
		defer f.Close() // Close file on end function
		logger := log.New(f, "", log.LstdFlags|log.LUTC)
		if tipo != "" {
			logger.Printf("%s: %v", tipo, msg) // Write msg to file
		} else {
			logger.Print(msg)
		}
	} else {
		log.Print(err)
	}
}

func FindInSlice(value string, slice []string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}

	return false
}
