package main

import (
    "fmt"
    "net/http"
    "os"
    "reflect"
    "strings"

    "github.com/op/go-logging"
    "github.com/gorilla/websocket"
    "github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

// Get Consumer key/secret and Access token/secret from enviroment variables
var consumerKey = os.Getenv("TWITTER_CONSUMER_KEY")
var consumerSecret = os.Getenv("TWITTER_CONSUMER_SECRET")
var accessToken = os.Getenv("TWITTER_ACCESS_TOKEN")
var accessSecret = os.Getenv("TWITTER_ACCESS_SECRET")

var config = oauth1.NewConfig(consumerKey, consumerSecret)
var token = oauth1.NewToken(accessToken, accessSecret)

// Auto Oauth1
var httpClient = config.Client(oauth1.NoContext, token)

// Twitter Client
var client = twitter.NewClient(httpClient)

// Websocket Upgrade buffer 
// Allow any domain can access stream API
var upgrader = websocket.Upgrader {
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool { return true },
}

var filters = []string{};
var conns = make(map[string]*websocket.Conn)

var log = logging.MustGetLogger("Twitter Tweeds Log")

var globalStream *twitter.Stream


// Start the app then Listen and Serve 
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
        
        fmt.Println("CONN:",reflect.TypeOf(conn))
        filters = append(filters,filter[0]);
        conns[filter[0]] = conn;
        fmt.Println("Filters",filters)
        fmt.Println("Connections:",conns)
        
        if globalStream != nil {
            globalStream.Stop()
        }
    	
    	// Read only tweet from twitter stream
    	demux := twitter.NewSwitchDemux()
    	demux.Tweet = func(tweet *twitter.Tweet) {
    	    text := tweet.Text
	    
    	    for _, v := range filters {
    	       if strings.Contains(text,v) && conns[v] != nil {
    	           wsWriter(conns[v],text)
    	       }    
    	    }
    	}
    	
    	fmt.Println("Starting Stream...")
    	
    	// Filter
    	filterParams := &twitter.StreamFilterParams{
    		Track:         filters,
    		StallWarnings: twitter.Bool(true),
    	}
    	
    	stream, err := client.Streams.Filter(filterParams)
    	
    	globalStream = stream
    	
    	fmt.Println(reflect.TypeOf(stream))
    	if err != nil {
    		log.Error(err)
    	}
    	
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
            
            for index, value := range filters {
                if value == filter {
                    // remove filtered word from filter slice
                    filters = append(filters[:index], filters[index+1:]...)
                    break
                }    
            }
            
             // Remove connections from connection list
            delete(conns, filter)      
            
            // Close connection
            conn.Close()
            log.Info("wsReader/WS Connection Closed: "+filter)
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