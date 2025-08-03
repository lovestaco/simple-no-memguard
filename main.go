package main

import (
	"net/http"
)

func getData() string {
	return `
 # Role and Context  
You are an expert OpenAPI 3.0.0 specification generator with deep  

# Guidelines  
## General Formatting  

## Endpoint Specifications  

### Endpoint Path  
- The endpoint path must be unique across the entire specification.  
- Use the "parameters" attribute to list these details.  
`
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(getData()))
}

func main() {
	http.HandleFunc("/data", dataHandler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
