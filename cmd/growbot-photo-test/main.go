// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// See https://github.com/gorilla/websocket/blob/master/examples/echo/client.go

package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

var scheme = flag.String("scheme", "ws", "ws scheme")
var addr = flag.String("addr", "localhost:8080", "http service address")
var uuid = flag.String("uuid", "", "uuid to use")
var photo = flag.String("photo", "", "photo to upload")
var plant = flag.String("plant", "", "plant id")

func main() {
	flag.Parse()
	log.SetFlags(0)

	if *uuid == "" {
		log.Fatal("expected uuid, got empty string")
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: *scheme, Host: *addr, Path: "/stream/" + *uuid}
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

	img := struct {
		Type string `json:"type"`
		Data struct {
			ID    string `json:"id"`
			Image string `json:"image"`
		} `json:"data"`
	}{
		Type: "PLANT_CAPTURE_PHOTO",
	}

	img.Data.ID = *plant

	b, err := ioutil.ReadFile(*photo)
	if err != nil {
		panic(err)
	}
	img.Data.Image = base64.StdEncoding.EncodeToString(b)

	err = c.WriteJSON(img)
	if err != nil {
		panic(err)
	}

	fmt.Println("Job done.")
	if err := c.Close(); err != nil {
		panic(err)
	}
	// c.Subprotocol

	for {
		select {
		case <-done:
			return
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
