# Setup

This git repo is just the `plans` dir for a `stacks-node` testground plan.

## Local
Setup local directories:
```bash
export TESTGROUND_ROOT=/opt/testgrounds
export TESTGROUND_HOME=$TESTGROUND_ROOT/testground
export STACKS_BLOCKCHAIN_DIR=$TESTGROUND_ROOT/stacks-blockchain
mkdir -p $TESTGROUND_ROOT
git clone https://github.com/blockstack/stacks-blockchain.git $STACKS_BLOCKCHAIN_DIR
```

Download testground and install/configure:
```bash
git clone https://github.com/testground/testground.git $TESTGROUND_HOME
cd $TESTGROUND_HOME \
  && make install

cat <<EOF > $TESTGROUND_HOME/.env.toml
["aws"]
region = "us-west-2"
["daemon"]
influxdb_endpoint = "http://testground-influxdb:8086"
EOF
```
* *`env.toml` in this repo can also be used*

Start testground in daemon mode:
```bash
cd $TESTGROUND_HOME && nohup testground daemon &
```

Import this test plan:
```bash
testground plan import \
  --git \
  --from https://github.com/blockstack/stacks-blockchain-testground \
  --name stacks-node
```

## Debian VM
Setup local directories:
```bash
curl -L https://golang.org/dl/go1.15.6.linux-amd64.tar.gz -o /tmp/go1.15.6.linux-amd64.tar.gz
tar -C /usr/local -xzf /tmp/go1.15.6.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
go version
apt-get update -y && apt-get install -y software-properties-common
curl -fsSL https://download.docker.com/linux/debian/gpg | sudo apt-key add -
add-apt-repository \
   "deb [arch=amd64] https://download.docker.com/linux/debian \
   $(lsb_release -cs) \
   stable"
apt-get update -y && apt-get install -y git docker-ce docker-ce-cli containerd.io build-essential
export TESTGROUND_ROOT=/opt/testgrounds
export TESTGROUND_HOME=$TESTGROUND_ROOT/testground
export STACKS_BLOCKCHAIN_DIR=$TESTGROUND_ROOT/stacks-blockchain
mkdir -p $TESTGROUND_ROOT
git clone https://github.com/blockstack/stacks-blockchain.git $STACKS_BLOCKCHAIN_DIR
sed -i -e 's|block_time = 60000|block_time = 120000|' $STACKS_BLOCKCHAIN_DIR/net-test/etc/bitcoin-neon-controller.toml.in
```

Download testground and install/configure:
```bash
git clone https://github.com/testground/testground.git $TESTGROUND_HOME
cd $TESTGROUND_HOME \
  && make install \
  && cp /root/go/bin/testground /usr/local/bin/

cat <<EOF > $TESTGROUND_HOME/.env.toml
["aws"]
region = "us-west-2"
["daemon"]
influxdb_endpoint = "http://testground-influxdb:8086"
EOF
```
* *`env.toml` in this repo can also be used*

Start testground in daemon mode:
```bash
cd $TESTGROUND_HOME && nohup testground daemon &
```

Import this test plan:
```bash
testground plan import \
  --git \
  --from https://github.com/kantai/stacks-blockchain-testground \
  --name stacks-node
```

## Build Local test images
* *if you change the tag, you'll need to also update the `Dockerfile` in the [repo root](https://github.com/blockstack/stacks-blockchain-testground/blob/main/Dockerfile#L7)*

```bash
docker build \
  -f $TESTGROUND_HOME/plans/stacks-node/stacks-blockchain-dockerfiles/Dockerfile.buster \
  -t blockstack/stacks-blockchain:testground-base \
$STACKS_BLOCKCHAIN_DIR
```


# Running the test plan
*data stored at `$TESTGROUND_HOME/data`*
*logs can be tailed with `testground logs -f -t <ID>`

* 2 instance test to a height of `10`
```bash
testground \
    --vv run single \
    --plan=stacks-node \
    --testcase=blocks \
    --runner=local:docker \
    --builder=docker:generic \
    --instances=2 \
    --collect \
    --tp stacks_tip_height=10
```

* 30 instance test to a height of `150`
```bash
testground \
    --vv run single \
    --plan=stacks-node \
    --testcase=blocks \
    --runner=local:docker \
    --builder=docker:generic \
    --instances=30 \
    --collect \
    --tp stacks_tip_height=150
```

* 30 instance test to a height of `150` **without** a chain quality test
```bash
testground \
    --vv run single \
    --plan=stacks-node \
    --testcase=blocks \
    --runner=local:docker \
    --builder=docker:generic \
    --instances=30 \
    --collect \
    --tp stacks_tip_height=150 \
    --tp verify_chain=false
```
