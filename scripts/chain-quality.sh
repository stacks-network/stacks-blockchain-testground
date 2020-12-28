#!/usr/bin/env bash
# set -x
STACKS_MASTER_NAME="master"
cd /src/net-test/tests/
source ./config.sh
source ./testlib.sh
# PROCESS_EXIT_AT_BLOCK_HEIGHT=450
source "$__BIN/start.sh"
set -uo pipefail
CONFIG_MODE="$STACKS_MASTER_NAME"
# check_chain_quality 90 90 100
check_chain_quality $1 $2 $3
if [ $? -ne 0 ]; then
   echo "[FAILED] - Chain quality check failed"
else
  echo "[OK] - Chain quality check passed"
fi
exit 0
