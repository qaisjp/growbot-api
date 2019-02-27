// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// See https://github.com/gorilla/websocket/blob/master/examples/echo/client.go

package main

import (
	"encoding/base64"
	"flag"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8080", "http service address")
var uuid = flag.String("uuid", "", "uuid to use")

func main() {
	flag.Parse()
	log.SetFlags(0)

	if *uuid == "" {
		log.Fatal("expected uuid, got empty string")
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/stream-video/" + *uuid}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	files, err := ioutil.ReadDir("./")
	if err != nil {
		panic(err)
	}

	frames := [][]byte{}
	for _, info := range files {
		name := info.Name()
		if strings.HasPrefix(name, "frame") {
			b, err := ioutil.ReadFile(name)
			if err != nil {
				panic(err)
			}
			frames = append(frames, b)
		}
	}

	previousLength := len(frames)

	// Now frames contains the data, lets do the same thing in reverse
	for i := len(frames) - 1; i >= 0; i-- {
		frames = append(frames, frames[i])
	}

	if len(frames) != (previousLength * 2) {
		panic("You did an oops with the for loop logic there ^")
	}

	// Replace each frame with a base64 encoded string
	for i := 0; i < len(frames); i++ {
		frames[i] = []byte(base64.StdEncoding.EncodeToString(frames[i]))
	}

	ticker := time.NewTicker(time.Millisecond * 200)
	defer ticker.Stop()

	currentFrame := 0
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			currentFrame++
			err := c.WriteMessage(websocket.TextMessage, frames[currentFrame%len(frames)])
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
