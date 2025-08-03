package main

import (
	"net/http"

	"github.com/awnumar/memguard"
)

var guardedData *memguard.LockedBuffer

func getData() *memguard.LockedBuffer {
	return guardedData
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	dataBuffer := getData()
	w.Write(dataBuffer.Bytes())
}

func main() {
	memguard.CatchInterrupt()
	defer memguard.Purge()

	rawData := `
 # Role and Context  
You are an expert OpenAPI 3.0.0 specification generator with deep  

# Guidelines  
## General Formatting  

## Endpoint Specifications  

### Endpoint Path  
- The endpoint path must be unique across the entire specification.  
- Use the "parameters" attribute to list these details.  
`
	guardedData = memguard.NewBufferFromBytes([]byte(rawData))
	defer guardedData.Destroy()

	http.HandleFunc("/data", dataHandler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
