package main

import (
	"time"
	"os/exec"
	"fmt"
	"io"
	"strconv"
	"os"
	"net"
	"bufio"
	"encoding/json"
	"net/http"
	"github.com/testground/sdk-go/runtime"
)

func Readln(in io.Reader, timeout time.Duration) (string, error) {
	s := make(chan string)
	e := make(chan error)

	go func() {
		reader := bufio.NewReader(in)
		line, err := reader.ReadString('\n')
		if err != nil {
			e <- err
		} else {
			s <- line
		}
		close(s)
		close(e)
	}()

	select {
	case line := <-s:
		return line, nil
	case err := <-e:
		return "", err
	case <-time.After(timeout):
		return "", nil
	}
}

func NodeNeighbors(runenv *runtime.RunEnv) {
	client := http.Client{}
	request, err := http.NewRequest("GET", "http://localhost:20443/v2/neighbors", nil)
	if err != nil {
		runenv.RecordMessage(fmt.Sprintf("%s", err))
		return
	}

	resp, err := client.Do(request)
	if err != nil {
		runenv.RecordMessage(fmt.Sprintf("Waiting for miner: [%s]", err))
		return
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
  lenInbound := len(result["inbound"].([]interface{}))
  lenOutbound := len(result["inbound"].([]interface{}))
  if lenInbound > 0 || lenOutbound > 0 {
    //runenv.RecordMessage(fmt.Sprintf("Neighbors => Inbound [ %v ] :: Outbound [ %v ]", lenInbound, lenOutbound))
    runenv.RecordMessage(fmt.Sprintf("Inbound/Outbound Neighbors => %v/%v", lenInbound, lenOutbound))
  }
  return
}

func NodeStatus(runenv *runtime.RunEnv, btcAddr string) (result float64, err error) {
	client := http.Client{}
	request, err := http.NewRequest("GET", "http://localhost:20443/v2/info", nil)
	btcPort := []string{"28443"}
	if len(btcAddr) > 0 {
		btcConn := btcConnect(btcAddr, btcPort)
		if !btcConn {
			runenv.RecordMessage("BTC Connection is closed -> Stopping this instance")
			runenv.RecordMessage("Setting an artificial stacks_tip_height to: 1000000")
			return float64(1000000), nil
		}
	}
	if err != nil {
		runenv.RecordMessage(fmt.Sprintf("%s", err))
		return
	}
	resp, err := client.Do(request)
	if err != nil {
		runenv.RecordMessage(fmt.Sprintf("Waiting for miner: [%s]", err))
		return
	}
	var item map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&item)
	runenv.RecordMessage(fmt.Sprintf("Stacks block height => %.0f :: Burn block height => %.0f", item["stacks_tip_height"], item["burn_block_height"]))
	// Extra info to tell us how many neighbors each instance has.
	// NodeNeighbors(runenv)
	return item["stacks_tip_height"].(float64), nil
}

func btcConnect(host string, ports []string) bool {
	for _, port := range ports {
		timeout := (5*time.Second)
		// timeout := time.Second
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
		if err != nil {
			// fmt.Println("BTC Connection error:", err)
			return false
		}
		if conn != nil {
			defer conn.Close()
			//fmt.Println("BTC Connection Open: ", net.JoinHostPort(host, port))
		}
	}
	return true
}

func chainQuality(runenv *runtime.RunEnv, sortitionFraction int, forkFraction int, numBlocks int) bool {
  runenv.RecordMessage("sortitionFraction: ", sortitionFraction)
  runenv.RecordMessage("forkFraction:      ", forkFraction)
  runenv.RecordMessage("numBlocks:         ", numBlocks)
  cmd := exec.Command("/scripts/chain-quality.sh", strconv.Itoa(sortitionFraction), strconv.Itoa(forkFraction), strconv.Itoa(numBlocks))
  outfile, err  := os.Create("/src/net-test/mnt/chain-quality.log")
  if err != nil {
    runenv.RecordMessage("Error Creating logfile: ", cmd, err)
  }
  cmd.Stdout = outfile
  cmd.Stderr = outfile
  runenv.RecordMessage("Running: ", cmd)
  if err := cmd.Run(); err != nil {
      fmt.Println( "Error:", err )
      return false
  }
  return true
}

func HandleNode(commPipe io.Reader, runenv *runtime.RunEnv, c *exec.Cmd, btcAddr string) error {
	tipHeight := float64(runenv.IntParam("stacks_tip_height"))
	startTime := time.Now()
	for {
		time.Sleep(15 * time.Second)
		output,nil := NodeStatus(runenv, btcAddr)
		if ( output > tipHeight ) {
			checkChainQuality := chainQuality(runenv, runenv.IntParam("sortition_fraction"),runenv.IntParam("fork_fraction"),runenv.IntParam("num_blocks"))
			if !checkChainQuality {
				runenv.RecordMessage("[ FAIL ] - check_chain_quality did not pass inspection")
			} else {
				runenv.RecordMessage("[ PASS ] - check_chain_quality passed inspection")
			}
			runenv.RecordMessage("Finished running after %v blocks (%v minutes)", output, time.Since(startTime).Minutes())
			return nil
		}
	 }
	c.Wait()
	return nil
}
