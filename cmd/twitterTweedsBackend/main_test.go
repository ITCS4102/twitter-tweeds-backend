package main

import "testing"

func TestHelloWorld(t *testing.T) {
    result := HelloTweetDeck()
    if result != "Tweet Deck Test" {
        t.Error("Expected Tweet Deck Test, got ", result)
    }
}