package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/olekukonko/tablewriter"
)

const (
	greenColor = "\033[32m"
	resetColor = "\033[0m"
)

func readASCIIArt(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func scanPort(protocol, hostname string, port int, results chan<- []string, wg *sync.WaitGroup) {
	defer wg.Done()
	address := hostname + ":" + strconv.Itoa(port)

	conn, err := net.DialTimeout(protocol, address, 1*time.Second)
	if err == nil {
		results <- []string{
			strconv.Itoa(port),
			strings.ToUpper(protocol),
			greenColor + "OPEN" + resetColor,
		}
		conn.Close()
	}
}

func displaySpinner(done chan struct{}) {
	chars := `|/-\`
	i := 0
	for {
		select {
		case <-done:
			fmt.Print("\rDone! \n")
			return
		default:
			fmt.Printf("\rPlease wait %s", string(chars[i]))
			i = (i + 1) % len(chars)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func main() {
	asciiArt, err := readASCIIArt("ascii_art.txt")
	if err != nil {
		fmt.Println("Error reading ASCII art file:", err)
		return
	}

	fmt.Println(asciiArt)

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter hostname or IP address (or type 'exit' to quit): ")
		hostname, _ := reader.ReadString('\n')
		hostname = strings.TrimSpace(hostname)

		if strings.ToLower(hostname) == "exit" {
			fmt.Println("Exiting program.")
			break
		}

		startTime := time.Now()

		fmt.Printf("Scanning ports on %s...\n", hostname)

		done := make(chan struct{})
		go displaySpinner(done)

		var wg sync.WaitGroup
		results := make(chan []string, 65535)

		for port := 1; port <= 65535; port++ {
			wg.Add(1)
			go scanPort("tcp", hostname, port, results, &wg)
		}

		// Wait for all goroutines to finish for better performance
		go func() {
			wg.Wait()
			close(results)
			close(done)
		}()

		var openPorts [][]string

		for result := range results {
			openPorts = append(openPorts, result)
		}

		fmt.Printf("\n%sWe were able to find %d open TCP ports%s\n", greenColor, len(openPorts), resetColor)

		//table settings
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Port", "Protocol", "Status"})
		table.SetBorder(true)
		table.SetHeaderLine(true)
		table.SetCenterSeparator("|")
		table.SetColumnSeparator("|")
		table.SetRowSeparator("-")
		table.SetTablePadding("\t")

		table.SetColumnAlignment([]int{
			tablewriter.ALIGN_LEFT,   // Right align numbers (Port)
			tablewriter.ALIGN_CENTER, // Center align text (Protocol)
			tablewriter.ALIGN_CENTER, // Center align text (Status)
		})

		for _, port := range openPorts {
			table.Append(port)
		}

		table.Render()

		duration := time.Since(startTime).Round(time.Second)

		fmt.Printf("\nPort scanning completed in %s.\n", duration)

		fmt.Print("\nDo you want to scan another IP address? (yes/no): ")
		again, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(again)) != "yes" {
			fmt.Println("Exiting program.")
			break
		}
	}
}
