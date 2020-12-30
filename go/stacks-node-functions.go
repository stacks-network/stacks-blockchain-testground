package main

import (
	"time"
	"os/exec"
	"fmt"
	"io"
	"strconv"
	"os"
	"net"
	"encoding/json"
	"net/http"
	"github.com/testground/sdk-go/runtime"
)

// NodeNeighbors - Retrieve the node's neighbors
func NodeNeighbors(runenv *runtime.RunEnv) {
	client := http.Client{}
	request, err := http.NewRequest("GET", "http://localhost:20443/v2/neighbors", nil)
	if err != nil {
		runenv.RecordMessage(fmt.Sprintf("%s", err))
		return
	}

	resp, err := client.Do(request)
	if err != nil {
		runenv.RecordMessage(fmt.Sprintf("Waiting for node: [%s]", err))
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

// NodeStatus - Retrieve the node's status
func NodeStatus(runenv *runtime.RunEnv, btcAddr string, seq int64) (result float64, err error) {
	client := http.Client{}
	request, err := http.NewRequest("GET", "http://localhost:20443/v2/info", nil)
	btcPort := []string{"28443"}
	if (len(btcAddr) > 0 && seq != 1) {
		btcConn := btcConnect(runenv, btcAddr, btcPort)
		if !btcConn {
			fakeHeight := (runenv.IntParam("stacks_tip_height")+10) // ensure we'll cross the threshold
			runenv.RecordMessage("BTC Connection is closed -> Stopping this instance")
			runenv.RecordMessage("Setting an artificial stacks_tip_height to: %v", fakeHeight)
			return float64(fakeHeight), nil
		}
	}
	if err != nil {
		runenv.RecordMessage(fmt.Sprintf("%s", err))
		return
	}
	resp, err := client.Do(request)
	if err != nil {
		runenv.RecordMessage(fmt.Sprintf("Waiting for node: [%s]", err))
		return
	}
	var item map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&item)
	runenv.RecordMessage(fmt.Sprintf("Stacks block height => %.0f :: Burn block height => %.0f", item["stacks_tip_height"], item["burn_block_height"]))
	// Extra info to tell us how many neighbors each instance has.
	// NodeNeighbors(runenv)
	return item["stacks_tip_height"].(float64), nil
}

// Check if BTC is accessible
func btcConnect(runenv *runtime.RunEnv, host string, ports []string) bool {
	for _, port := range ports {
		timeout := (5*time.Second)
		// timeout := time.Second
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
		if err != nil {
			// BTC Connection is not open
			runenv.RecordMessage(fmt.Sprintf("Error Connecting to BTC: %s", err))
			return false
		}
		if conn != nil {
			// BTC connection is successful
			// fmt.Println("BTC Connection Open: ", net.JoinHostPort(host, port))
			defer conn.Close()
		}
	}
	return true
}

// Check the chainstate quality
func chainQuality(runenv *runtime.RunEnv, sortitionFraction int, forkFraction int, numBlocks int) bool {
	cmd := exec.Command("/scripts/chain-quality.sh", strconv.Itoa(sortitionFraction), strconv.Itoa(forkFraction), strconv.Itoa(numBlocks))
	outfile, err  := os.Create("/src/net-test/mnt/chain-quality.log")
	if err != nil {
		runenv.RecordMessage("Error Creating Logfile:", err)
	}
	cmd.Stdout = outfile
	cmd.Stderr = outfile
	if err := cmd.Run(); err != nil {
		runenv.RecordMessage("Error Running Command:", cmd)
		runenv.RecordMessage(fmt.Sprintf("%s", err))
		return false
	}
	return true
}

// HandleNode Function
func HandleNode(commPipe io.Reader, runenv *runtime.RunEnv, c *exec.Cmd, btcAddr string, seq int64) error {
	tipHeight := float64(runenv.IntParam("stacks_tip_height"))
	startTime := time.Now()
	runenv.RecordMessage("verify_chain1: %v", runenv.BooleanParam("verify_chain"))
	verify_chain := runenv.BooleanParam("verify_chain")
	runenv.RecordMessage("verify_chain2: %v", verify_chain)
	for {
		time.Sleep(15 * time.Second)
		output,nil := NodeStatus(runenv, btcAddr, seq)
		if ( output >= tipHeight ) {
			if ( seq == 1 && verify_chain) {
				checkChainQuality := chainQuality(runenv, runenv.IntParam("sortition_fraction"),runenv.IntParam("fork_fraction"),runenv.IntParam("num_blocks"))
				if !checkChainQuality {
					runenv.RecordMessage("[ FAIL ] - check_chain_quality did not pass inspection: %v", checkChainQuality)
				} else {
					runenv.RecordMessage("[ PASS ] - check_chain_quality passed inspection %v", checkChainQuality)
				}
			}
			runenv.RecordMessage("Finished running after %v blocks (%v minutes)", output, time.Since(startTime).Minutes())
			return nil
		}
	 }
	c.Wait()
	return nil
}
