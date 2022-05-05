// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"sync"

	_ "github.com/google/uuid"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

type Server struct {
	*Hub
}

func newServer() *Server{
	return &Server{
		Hub: newHub(),
	}
}

var lock = &sync.Mutex{}
var singleInstance *Server

func GetInstanceOfServer() *Server {
    if singleInstance == nil {
        lock.Lock()
        defer lock.Unlock()
        if singleInstance == nil {
            fmt.Println("Creating single instance now.")
            singleInstance = newServer()
        } else {
            fmt.Println("Single instance already created.")
        }
    } else {
        fmt.Println("Single instance already created.")
    }
    return singleInstance
}


func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			if client.Id == "" {
				client.Id = "asd"//uuid.Must(uuid.NewRandom()).String()
				fmt.Println("hubrun: " + client.Id)
				h.clients[client] = true
				fmt.Println(len(h.clients))
			}
		case client := <-h.unregister:
			fmt.Println("unregister")
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func (s *Server) GetClientById(id string) (*Client, error) {
	fmt.Println("test1")
	for k := range s.Hub.clients {
		fmt.Println("test")
		if k.Id == id {
			return k, nil
		}
	}
	return nil, fmt.Errorf("Client with id: '" + id + "' doesn't exist.")
}
