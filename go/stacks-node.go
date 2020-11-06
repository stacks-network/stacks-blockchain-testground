package main

import (
	"context"
	"time"
	"os/exec"
	"fmt"
	"io"
	"bufio"
	"encoding/json"
	"net/http"
	"github.com/testground/sdk-go/network"
	"github.com/testground/sdk-go/runtime"
	"github.com/testground/sdk-go/sync"
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

func NodeStatus(runenv *runtime.RunEnv) {
	client := http.Client{}
	request, err := http.NewRequest("GET", "http://localhost:20443/v2/info", nil)
	if err != nil {
		runenv.RecordMessage(fmt.Sprintf("%s", err))
		return
	}

	resp, err := client.Do(request)
	if err != nil {
		runenv.RecordMessage(fmt.Sprintf("%s", err))
		return
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	runenv.RecordMessage(fmt.Sprintf("Stacks block height => %.0f", result["stacks_tip_height"]))
	runenv.RecordMessage(fmt.Sprintf("Burn block height => %.0f", result["burn_block_height"]))
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
		runenv.RecordMessage(fmt.Sprintf("%s", err))
		return
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	runenv.RecordMessage(fmt.Sprintf("%s", result))
}

func HandleNode(comm_pipe io.Reader, runenv *runtime.RunEnv, c *exec.Cmd) error {
	testLength := runenv.IntParam("test_time_mins")
	startTime := time.Now()

	for {
		time.Sleep(15 * time.Second)
		NodeStatus(runenv)
	// 	timeout, _ := time.ParseDuration("5s")
	// 	data, err := Readln(comm_pipe, timeout)
	// 	if data != "" {
	// 		runenv.RecordMessage(data)
	// 	}
	// 	if err == io.EOF {
	// 		runenv.RecordMessage("%s", err)
	// 		return nil
	// 	} else if err != nil {
	// 		runenv.RecordMessage("%s", err)
	// 		return err
	// 	}

	 	if time.Since(startTime).Minutes() > float64(testLength) {
	 		runenv.RecordMessage("Finished running after %d minutes", testLength)
	 		return nil
	 	}
	 }

	c.Wait()
	return nil

}

func StacksNode(runenv *runtime.RunEnv) error {
    startState := sync.State("start")
    btcState := sync.State("bitcoin-start")
    ctx := context.Background()
    btcInformation := sync.NewTopic("btc-address", "")

	// instantiate a sync service client, binding it to the RunEnv.
	client := sync.MustBoundClient(ctx, runenv)
	defer client.Close()

	// instantiate a network client; see 'Traffic shaping' in the docs.
	netclient := network.NewClient(client, runenv)
	runenv.RecordMessage("waiting for network initialization")

	// wait for the network to initialize; this should be pretty fast.
	netclient.MustWaitNetworkInitialized(ctx)
	runenv.RecordMessage("network initilization complete")

	ip_addr, ip_err := netclient.GetDataNetworkIP()
	if ip_err != nil {
		return ip_err
	}

	// signal entry in the 'enrolled' state, and obtain a sequence number.
	seq := client.MustSignalEntry(ctx, startState)

	runenv.RecordMessage("my sequence ID: %d", seq)

	// if we're the first instance to signal, we'll become the LEADER.
	if seq == 1 {
		client.MustPublish(ctx, btcInformation, ip_addr.String())

		cmd := exec.Command("/scripts/simple-start.sh", "master", ip_addr.String())
		pipe, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}

		err = cmd.Start()
		if err != nil {
			return err
		}

		time.Sleep(5 * time.Second)
		client.MustSignalEntry(ctx, btcState)

		return HandleNode(pipe, runenv, cmd)
	} else {
		// wait until leader has started Bitcoin.
		err := <-client.MustBarrier(ctx, btcState, 1).C
		if err != nil {
			return err
		}

		ch := make(chan string)
		client.MustSubscribe(ctx, btcInformation, ch)
		btc_addr := <-ch

		runenv.RecordMessage("Master started on host address %s", btc_addr)

		cmd := exec.Command("/scripts/simple-start.sh", "miner", ip_addr.String(), btc_addr)
		pipe, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}

		err = cmd.Start()
		if err != nil {
			return err
		}

		time.Sleep(5 * time.Second)
		return HandleNode(pipe, runenv, cmd)
	}
}
