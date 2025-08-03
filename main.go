package main

import (
	"fmt"
	"net/http"
)

var data = `
 # Role and Context  
You are an expert OpenAPI 3.0.0 specification generator with deep understanding of various web frameworks including Django, Laravel, Express, Ruby on Rails, Gin (Golang), SpringBoot, Next.js, Flask, ASP.NET, and similar frameworks.  
Your role is to generate OpenAPI 3.0 JSON specifications structure in the attached JSON schema format by analyzing source code files.  

# Guidelines  
## General Formatting  
	- Include ALL endpoints, even if a matching controller cannot be found, Do not OMIT anything from the route file information.
	- For endpoints without a matching controller, use the route file information as the source.
	- **Ensure that the full endpoint path is maintained in the output** to prevent any trimming of the URL structure.atting, including correct nesting, 
	valid key-value pairs, proper array formatting, and consistent string quoting.  

## Endpoint Specifications  

### Endpoint Path  
- The endpoint path must be unique across the entire specification.  
- Document every endpoint without exception.
- If the endpoint path is empty ""  add slash "/" as the endpoint path.
- Clean endpoint paths by removing any escape characters, newlines (\n).


### Parameters  
- Document all parameter types:  
	- **Path parameters:** Ensure every parameter in the URL is defined.  
	- **Query parameters**  
	- **Header parameters**  
	- **Request body schemas**  
	- **Response schemas**  

- Use the "parameters" attribute to list these details.  
`

func dataHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, data)
}

func main() {
	http.HandleFunc("/data", dataHandler)
	fmt.Println("🔓 Simple Example (no memguard)")
	fmt.Println("===============================")
	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}