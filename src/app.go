package main

import (
    "net/http"  
)

func main() {
    http.HandleFunc("/", home)
    
    http.ListenAndServe(":8080",nil)
}

func home(res http.ResponseWriter, req *http.Request) {
    http.ServeFile(res,req, "../public/index.html")
}


