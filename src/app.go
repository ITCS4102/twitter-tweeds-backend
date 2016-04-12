package main

import (
    "fmt"
    "net/http"  
    "time"

    "github.com/op/go-logging"
    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader {}
var log = logging.MustGetLogger("Twitter Tweeds Log")


func main() {
    http.HandleFunc("/", home)
    http.HandleFunc("/twitter/stream", twitterStream)
    
    http.ListenAndServe(":8080",nil)
}

func home(res http.ResponseWriter, req *http.Request) {
    http.ServeFile(res,req, "../public/index.html")
}

func twitterStream(res http.ResponseWriter, req *http.Request) {
    log.Info("WS Twitter Request: ",req.URL.Query())
    filter := req.URL.Query()["filter"]
    
    if filter != nil {
        
        // Upgrades the http server connection to the websocket protocol 
        conn, _ := upgrader.Upgrade(res, req, nil)
        
        go wsReader(conn,filter[0])
        go wsWriter(conn,filter[0])
        
    } else {
        fmt.Fprintf(res, "Warning: Filter param is required to stream the data from Twitter")
    }
}

// Read messages from the websocket
func wsReader(conn *websocket.Conn,filter string) {
    for {
        _, _, err := conn.ReadMessage()
        
        //Close the websocket connection
        if err != nil {
            conn.Close()
            log.Info("WS Connection Closed: "+filter)
            // fmt.Println("WS Connection Closed:",filter)
            break
        } else {
            log.Error(err)
        }
    }    
}
// Write messages to the websocket
func wsWriter(conn *websocket.Conn,filter string) {
    for {
        ch := time.Tick(5 *time.Second)
        
        for range ch {
            conn.WriteJSON(myStruct{
                Name: filter,
            })
        }
    }    
}



type myStruct struct {
    Name string
}