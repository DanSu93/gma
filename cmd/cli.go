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
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

// cliCmd represents the cli command
var cliCmd = &cobra.Command{
	Use:   "cli",
	Short: "Start gma console",
	Run: func(cmd *cobra.Command, args []string) {
		startCli()
	},
}

func init() {
	rootCmd.AddCommand(cliCmd)
	log.SetFlags(0)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cliCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cliCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func socketReader(r io.Reader) {
	// TODO: change it
	buf := make([]byte, 512)
	for {
		n, err := r.Read(buf[:])
		if err != nil {
			return
		}
		fmt.Println(string(buf[0:n]))
	}
}

func startCli() {
	cliReader := bufio.NewReader(os.Stdin)

	client, err := net.Dial("unix", "./tmp/echo.sock")
	if err != nil {
		if _, ok := err.(*net.OpError); ok {
			log.Fatal("Could not connect to Gma server: Connection refused")
		}

		log.Fatal("Dial error", err)
	}
	defer client.Close()

	go socketReader(client)

	for {
		fmt.Printf("> ")
		text, _ := cliReader.ReadString('\n')
		// split the text into operation strings
		operation := strings.Fields(text)
		switch operation[0] {
		case "exit":
			return
		case "SET", "GET", "DEL", "KEYS":
			handleSocketCommand(client, strings.Trim(text, "\n"))
		//case "DELETE": 		Delete(operation[1], items)
		default:
			fmt.Printf("ERROR: Unrecognised Operation %s\n", operation[0])
		}
	}
}

func handleSocketCommand(client net.Conn, text string) {
	if _, err := client.Write([]byte(text)); err != nil {
		fmt.Println("Write error:", err)
		return
	}
}
