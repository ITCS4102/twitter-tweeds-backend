package main

import (
    "fmt"
    "net/http"
    "reflect"

    "github.com/op/go-logging"
    "github.com/gorilla/websocket"
    "github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

// Websocket Upgrade buffer
var upgrader = websocket.Upgrader {}

var log = logging.MustGetLogger("Twitter Tweeds Log")

// Listen and Serve 
func main() {
    http.HandleFunc("/", home)
    http.HandleFunc("/twitter/stream", twitterStream)

    http.ListenAndServe(":8080",nil)
}

// Serve home page
func home(res http.ResponseWriter, req *http.Request) {
    res.Header().Set("Access-Control-Allow-Origin", "*")
    http.ServeFile(res,req, "../public/index.html")
}

// Serve Websocket Twitter stream
func twitterStream(res http.ResponseWriter, req *http.Request) {
    log.Info("WS Twitter Request: ",req.URL.Query())
    filter := req.URL.Query()["filter"]
    if filter != nil {
        // Upgrades the http server connection to the websocket protocol 
        conn, _ := upgrader.Upgrade(res, req, nil)
        
        consumerKey := ""
    	consumerSecret := ""
    	accessToken := ""
    	accessSecret := ""
    	
        if consumerKey == "" || consumerSecret == "" || accessToken == "" || accessSecret == "" {
    		log.Error("Consumer key/secret and Access token/secret required")
    	}
    
    	config := oauth1.NewConfig(consumerKey, consumerSecret)
    	token := oauth1.NewToken(accessToken, accessSecret)
    	
    	httpClient := config.Client(oauth1.NoContext, token)
    	
    	client := twitter.NewClient(httpClient)
    	
    	demux := twitter.NewSwitchDemux()
    	demux.Tweet = func(tweet *twitter.Tweet) {
    		fmt.Println("Stream/"+filter[0]+":",tweet.Text)
    		wsWriter(conn,tweet.Text)
    	}
    	
    	fmt.Println("Starting Stream...")
    	
    	// Filter
    	filterParams := &twitter.StreamFilterParams{
    		Track:         []string{filter[0]},
    		StallWarnings: twitter.Bool(true),
    	}
    	
    	stream, err := client.Streams.Filter(filterParams)
    	if err != nil {
    		log.Error(err)
    	}
    	
    	fmt.Println("Type:", reflect.TypeOf(stream))
        
        go wsReader(conn,filter[0],stream)      // Read messages from the websocket or stop websoket and twitter stream
    	go demux.HandleChan(stream.Messages)    // Receive twitter messages until stream quits
        
    } else {
        fmt.Fprintf(res, "Warning: Filter param is required to stream the data from Twitter")
    }
}

// Read messages from the websocket
func wsReader(conn *websocket.Conn,filter string,stream *twitter.Stream) {
    for {
        _, _, err := conn.ReadMessage()
        
        //Close the websocket connection
        if err != nil {
            conn.Close()
            stream.Stop()
            log.Info("wsReader/WS Connection Closed: "+filter)
            log.Info("wsReader/Twitter Stream Closed: "+filter)
            break
        } else {
            log.Error(err)
        }
    }    
}

// Write messages to the websocket
func wsWriter(conn *websocket.Conn,filter string) {
    conn.WriteJSON(tweetStruct{
        Tweet: filter,
    })      
}

type tweetStruct struct {
    Tweet string
}