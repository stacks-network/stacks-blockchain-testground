#!/usr/bin/env bash
set -x
set -uo pipefail
STACKS_MASTER_NAME="master"
cd /src/net-test/tests/
source ./config.sh
source ./testlib.sh
source "$__BIN/start.sh"
CONFIG_MODE="$STACKS_MASTER_NAME"

check_chain_quality $1 $2 $3
RET=$?
if [ $RET -ne 0 ]; then
   echo "[FAILED] - Chain quality check failed"
else
  echo "[OK] - Chain quality check passed"
fi
echo "exiting with $RET"
exit $RET
