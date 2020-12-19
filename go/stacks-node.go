package main

import (
	"time"
	"context"
	"os/exec"
	"github.com/testground/sdk-go/network"
	"github.com/testground/sdk-go/runtime"
	"github.com/testground/sdk-go/sync"
)

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
