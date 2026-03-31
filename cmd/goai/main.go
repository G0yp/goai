package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"net/http"
)

func main() {
	repl()
}

func repl() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Enter input (type exit to end process): ")
	for scanner.Scan() {
		text := scanner.Text()

		text = strings.TrimSpace(text)
		if text == "exit" {
			return
		}
		fmt.Println(text)
		fmt.Print("Enter input (type exit to end process): ")
	}
}

func makeRequest() {

}
