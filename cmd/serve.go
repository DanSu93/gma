/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start gma server",
	Run: func(cmd *cobra.Command, args []string) {
		startServer()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func startServer() {
	// channel for handling *nix termination commands
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	store := make(map[string]string)

	// create unix socket file
	listener, err := net.Listen("unix", "./tmp/echo.sock")
	if err != nil {
		log.Fatal("listen error:", err)
	}

	defer listener.Close()

	fmt.Println("Gma server has been started successfully")

	listenerChan := make(chan net.Conn)

	go func() {
		for {
			connect, err := listener.Accept()

			if err != nil {
				// TODO: change me
				if _, ok := err.(*net.OpError); ok {
					log.Fatal("Connection refused")
				}

				log.Fatal("accept error: ", err)
			}
			listenerChan <- connect
		}
	}()

	for {
		select {
		case <-signalChan:
			listener.Close()
			os.Exit(1)
		case lol := <-listenerChan:
			go echoServer(lol, store)
		}
	}
}

func echoServer(connect net.Conn, store map[string]string) {
	defer connect.Close()

	for {
		// TODO: change it
		buf := make([]byte, 512)
		lastByte, err := connect.Read(buf)

		if err == io.EOF {
			connect.Close()
			return
		}

		if err != nil {
			log.Fatal("handle connect error:", err)
		}

		data := string(buf[0:lastByte])
		// TODO: strings.Fields(data)
		switch {
		case strings.HasPrefix(data, "SET "):
			handleSet(data, store, connect)
		case strings.HasPrefix(data, "GET "):
			handleGet(data, store, connect)
		case strings.HasPrefix(data, "DEL "):
			handleDel(data, store, connect)
		case strings.HasPrefix(data, "KEYS"):
			handleKeys(data, store, connect)
		}
	}
}

func handleSet(data string, store map[string]string, connect net.Conn) {
	arrayKV := strings.Fields(strings.TrimPrefix(data, "SET "))

	if isArgsCntCorrect(arrayKV, "SET", connect) {
		store[arrayKV[0]] = arrayKV[1]
		response(connect, "OK")
	}
}

func handleGet(data string, store map[string]string, connect net.Conn) {
	arrayKV := strings.Fields(strings.TrimPrefix(data, "GET "))

	if isArgsCntCorrect(arrayKV, "GET", connect) {
		result := store[arrayKV[0]]
		if result == "" {
			result = "(nil)"
		}
		response(connect, result)
	}
}

func handleDel(data string, store map[string]string, connect net.Conn) {
	arrayKeys := strings.Fields(strings.TrimPrefix(data, "DEL "))

	if !isArgsCntCorrect(arrayKeys, "DEL", connect) {
		return
	}
	count := 0
	for _, key := range arrayKeys {
		if _, found := store[key]; found {
			delete(store, key)
			count++
		}
	}
	response(connect, strconv.Itoa(count))
}

func handleKeys(data string, store map[string]string, connect net.Conn) {
	arrayKV := strings.Fields(strings.TrimPrefix(data, "KEYS "))

	if !isArgsCntCorrect(arrayKV, "KEYS", connect) {
		return
	}
	result := ""
	for key := range store {
		if found, err := regexp.MatchString(arrayKV[0], key); found {
			if err != nil {
				response(connect, "(error) Syntax error")
				return
			}
			result += key + "\n"
		}
	}

	if result == "" {
		result = "(nil)"
	}

	response(connect, strings.Trim(result, "\n"))
}

func isArgsCntCorrect(commandArgs []string, commandName string, connect net.Conn) bool {
	var isOk bool

	switch commandName {
	case "SET":
		isOk = len(commandArgs) == 2
	case "GET":
		isOk = len(commandArgs) == 1
	case "DEL":
		isOk = len(commandArgs) > 0
	case "KEYS":
		isOk = len(commandArgs) == 1
	}

	if isOk {
		return true
	}

	response(connect, fmt.Sprintf("(error) ERR wrong number of arguments for %s command", commandName))
	return false
}

func response(connect net.Conn, data string) {
	if _, err := connect.Write([]byte(data)); err != nil {
		fmt.Println("Writing client error: ", err)
	}
}
