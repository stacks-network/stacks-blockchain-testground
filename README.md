# Setup

This git repo is just the `plans` dir for a `stacks-node` testground plan.

```
$ cd $TESTGROUND_HOME/plans
$ git clone https://github.com/kantai/stacks-blockchain-testground.git
```

# Making the Docker images

To set up the docker images used by the `stacks-node` plan:

```
$ git clone https://github.com/blockstack/stacks-blockchain.git
$ STACKS_BLOCKCHAIN_DIR="$(pwd)/stacks-blockchain"
$ cd $TESTGROUND_HOME/plans/stacks-node
$ docker build "$STACKS_BLOCKCHAIN_DIR" -f stacks-blockchain-dockerfiles/Dockerfile.buster -t stacks-blockchain:buster
$ docker build "$STACKS_BLOCKCHAIN_DIR" -f stacks-blockchain-dockerfiles/Dockerfile.testground-base -t stacks-blockchain:testground-base
```

# Running the test plan

```bash
testground run single --plan=stacks-node --testcase=node --runner=local:docker  --builder=docker:generic  --instances=4 --tp test_time_mins=5
```
