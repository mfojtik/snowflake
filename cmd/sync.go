// Copyright © 2016 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/mfojtik/snowflakes/pkg/generator"
	"github.com/mfojtik/snowflakes/pkg/sync"
	"github.com/spf13/cobra"
)

var (
	syncFormat   string
	syncFilename string
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize the flakes database",
	Run: func(cmd *cobra.Command, args []string) {
		controller := &sync.Controller{}
		controller.Run()
		if syncFormat == "json" {
			out := controller.JSONResult()
			if len(syncFilename) > 0 {
				if err := ioutil.WriteFile(syncFilename, out, 0644); err != nil {
					log.Fatalf("error while writing JSON file: %v", err)
				}
			} else {
				fmt.Fprintln(os.Stdout, string(out))
			}
			return
		}
		if syncFormat == "html" {
			out := generator.GenerateHTML(controller.SortedResult())
			if len(syncFilename) > 0 {
				if err := ioutil.WriteFile(syncFilename, []byte(out), 0644); err != nil {
					log.Fatalf("error while writing HTML file: %v", err)
				}
			} else {
				fmt.Fprintln(os.Stdout, out)
			}
			return
		}
		for _, r := range controller.SortedResult() {
			fmt.Printf("[%d|%d]: %s\n", r.Number, r.ReferenceCount, r.Title)
		}
	},
}

func init() {
	syncCmd.Flags().StringVarP(&syncFormat, "output", "o", "", "Output format ('json', 'html' or '')")
	syncCmd.Flags().StringVarP(&syncFilename, "file", "f", "", "Filename to write JSON to")
	RootCmd.AddCommand(syncCmd)
}
