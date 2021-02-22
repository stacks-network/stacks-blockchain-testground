#!/usr/bin/env bash
# set -x
cd /src/net-test/tests/
source ./config.sh
source "$__BIN/start.sh"

case "$1" in
    master)
        master_config "master" "127.0.0.1" "true"
        STACKS_MASTER_PUBLIC_IP="$2"
        BITCOIN_PUBLIC_IP="127.0.0.1"
        CHAINSTATE_DIR="/tmp/chainstate"
        PROCESS_EXIT_AT_BLOCK_HEIGHT=5000
        start_node &
        PID=$!
        wait_node || exit_error "node failed to boot up"
        tail -f $STACKS_MASTER_LOGFILE
        ;;
    miner)
        miner_config "miner" "20443" "20444"
        STACKS_MINER_PUBLIC_IP="$2"
        BITCOIN_PUBLIC_IP="$3"
        STACKS_MINER_BOOTSTRAP_IP="$3"
        STACKS_MINER_BOOTSTRAP_PORT=20444
        FAUCET_URL="http://$3:$FAUCET_PORT"
        CHAINSTATE_DIR="/tmp/chainstate"
        PROCESS_EXIT_AT_BLOCK_HEIGHT=5000
        start_node &
        PID=$!
        wait_node || exit_error "node failed to boot up"
        tail -f $STACKS_MINER_LOGFILE
        ;;
    follower)
        follower_config "follower" "20443" "20444"
        STACKS_FOLLOWER_PUBLIC_IP="$2"
        STACKS_FOLLOWER_BOOTSTRAP_IP="$3"
        STACKS_FOLLOWER_BOOTSTRAP_PORT=20444
        BITCOIN_PUBLIC_IP="$3"
        FAUCET_URL="http://$3:$FAUCET_PORT"
        CHAINSTATE_DIR="/tmp/chainstate"
        PROCESS_EXIT_AT_BLOCK_HEIGHT=5000
        start_node &
        PID=$!
        wait_node || exit_error "node failed to boot up"
        tail -f $STACKS_FOLLOWER_LOGFILE
        ;;
    miner_nat)
        miner_config "miner" "20443" "20444"
        STACKS_MINER_PUBLIC_IP="$2"
        BITCOIN_PUBLIC_IP="$3"
        STACKS_MINER_BOOTSTRAP_IP="$3"
        STACKS_MINER_BOOTSTRAP_PORT=20444
        FAUCET_URL="http://$3:$FAUCET_PORT"
        CHAINSTATE_DIR="/tmp/chainstate"
        PROCESS_EXIT_AT_BLOCK_HEIGHT=5000
        set_nat "true"
        start_node &
        PID=$!
        wait_node || exit_error "node failed to boot up"
        tail -f $STACKS_MINER_LOGFILE
        ;;
    follower_nat)
        follower_config "follower" "20443" "20444"
        STACKS_FOLLOWER_PUBLIC_IP="$2"
        STACKS_FOLLOWER_BOOTSTRAP_IP="$3"
        STACKS_FOLLOWER_BOOTSTRAP_PORT=20444
        BITCOIN_PUBLIC_IP="$3"
        FAUCET_URL="http://$3:$FAUCET_PORT"
        CHAINSTATE_DIR="/tmp/chainstate"
        PROCESS_EXIT_AT_BLOCK_HEIGHT=5000
        set_nat "true"
        start_node &
        PID=$!
        wait_node || exit_error "node failed to boot up"
        tail -f $STACKS_FOLLOWER_LOGFILE
        ;;
    *)
        exit 1
        ;;
esac
