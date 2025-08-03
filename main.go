package main

import (
	"net/http"

	"github.com/awnumar/memguard"
)

var protectedData *memguard.LockedBuffer


func dataHandler(w http.ResponseWriter, r *http.Request) {
	dataCopy := memguard.NewBuffer(protectedData.Size())
	copy(dataCopy.Buffer.Data(), protectedData.Buffer.Data())
	defer dataCopy.Destroy()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write(dataCopy.Buffer.Data())
}

func getData() string {
	// Simulate reading sensitive data (could be from env or secure file)
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
return rawData
}


func main() {
	// Important: Ensure memguard cleans up on interrupt or panic
	memguard.CatchInterrupt()
	defer memguard.Purge()


	// Create a protected LockedBuffer
	protectedData = memguard.NewBufferFromBytes([]byte(getData()))

	// Set up HTTP handler
	http.HandleFunc("/data", dataHandler)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
