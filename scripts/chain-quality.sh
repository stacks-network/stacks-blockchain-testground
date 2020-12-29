#!/usr/bin/env bash
STACKS_MASTER_NAME="master"
CONFIG_MODE="$STACKS_MASTER_NAME"
cd /src/net-test/tests/
source ./config.sh
source ./testlib.sh
source "$__BIN/start.sh"
set -xuo pipefail
check_chain_quality $1 $2 $3
if [ $? -ne 0 ]; then
   echo "[FAILED] - Chain quality check failed"
else
  echo "[OK] - Chain quality check passed"
fi
exit $?
