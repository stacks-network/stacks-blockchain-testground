# testground

```bash
export TESTGROUND_ROOT=$HOME/Git/blockstack/testgrounds
export TESTGROUND_HOME=$HOME/Git/blockstack/testgrounds/testground
export STACKS_BLOCKCHAIN_DIR=$TESTGROUND_ROOT/stacks-blockchain
mkdir -p $TESTGROUND_HOME
git clone https://github.com/testground/testground.git $TESTGROUND_HOME
cd $TESTGROUND_HOME && make install
cat <<EOF > $TESTGROUND_HOME/.env.toml
["aws"]
region = "us-west-2"
["daemon"]
influxdb_endpoint = "http://testground-influxdb:8086"
EOF
nohup testground daemon &
testground plan import --git --from https://github.com/blockstack/stacks-blockchain-testground --name stacks-node
git clone https://github.com/blockstack/stacks-blockchain.git $STACKS_BLOCKCHAIN_DIR
sed -i -e 's|block_time = 60000|block_time = 120000|' $STACKS_BLOCKCHAIN_DIR/net-test/etc/bitcoin-neon-controller.toml.in

```

## Build test images for local exec
### single image

```bash
docker build \
  -f $TESTGROUND_HOME/plans/stacks-node/stacks-blockchain-dockerfiles/Dockerfile.buster \
  -t blockstack/stacks-blockchain:testground-base \
$STACKS_BLOCKCHAIN_DIR
docker push blockstack/stacks-blockchain:testground-base
```



## K8S
https://docs.testground.ai/runner-library/cluster-k8s/how-to-create-a-kubernetes-cluster-for-testground
https://acloudxpert.com/setup-kubernetes-cluster-on-gcp-using-kops/

### testground env vars
**requires an `awscli` config named `testground`

`~/.aws/config`:
```bash
[profile testground]
region=us-west-1
output=json
```

`~/.aws/credentials`:
```bash
[testground]
aws_access_key_id = <aws_access_key_id>
aws_secret_access_key = <aws_access_key_id>
```


`~/.testground`:
```bash
export TESTGROUND_NAME=name # edit this to your name
export TESTGROUND_ROOT=$HOME/Git/blockstack/testgrounds
export TESTGROUND_HOME=$HOME/Git/blockstack/testgrounds/testground
export STACKS_BLOCKCHAIN_DIR=$TESTGROUND_ROOT/stacks-blockchain
export CLUSTER_NAME=$TESTGROUND_NAME.domain.xyz
export KOPS_STATE_STORE=s3://blockstack-testground-kops
export AWS_REGION=us-west-2
export ZONE_A=us-west-2a
export ZONE_B=us-west-2d
export WORKER_NODES=3
export MASTER_NODE_TYPE=c5.2xlarge
export WORKER_NODE_TYPE=c5.2xlarge
export PUBKEY=$HOME/.ssh/testground_rsa.pub
export TEAM=blockchain
export PROJECT=testground
export AWS_PROFILE=testground
export DEPLOYMENT_NAME=testground
EOF
```

### create ssh key for kops (or use existing key)
*if using existing key, edit `~/.testground` `PUBKEY` to reflect this*
```bash
ssh-keygen -t rsa -b 4096 -C "email address" \
  -f ~/.ssh/testground_rsa -q -P ""
```

### create bucket for testground assets
*this should already exist*
```bash
aws s3api create-bucket \
  --bucket blockstack-testground-kops \
  --region us-west-2 \
  --create-bucket-configuration LocationConstraint=us-west-2
```

### Helm repos
```bash
helm repo add stable https://kubernetes-charts.storage.googleapis.com/
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo add influxdata https://helm.influxdata.com/
helm repo update
```

### Install k8s cluster
*DNS can be problematic here during the first script*
*if `./k8s/01_install_k8s.sh ./k8s/cluster.yaml` fails, skip to DNS issues*

```bash
cd $TESTGROUND_ROOT/infra
source ~/.testground
./01_install_k8s.sh ./cluster.yaml
./02_efs.sh ./cluster.yaml
./03_ebs.sh ./cluster.yaml
```

#### DNS issues
- running `./01_install_k8s.sh ./cluster.yaml` can fail if DNS hasn't propagated quickly enough
- failure:
```
...
Cluster nodes are Ready

Install default container limits

Unable to connect to the server: dial tcp: lookup api.name.domain.xyz: no such host
Error on line 78
```
- verify the cluster api is responding (it will work once DNS is propagated)
```bash
$ kubectl get nodes
NAME                                          STATUS   ROLES    AGE     VERSION
ip-172-20-47-191.us-west-2.compute.internal   Ready    node     5m24s   v1.18.10
...
```

- once `kubectl get nodes` returns the nodes, run:
```
./01_install_k8s_secondary.sh
./02_efs.sh ./cluster.yaml
./03_ebs.sh ./cluster.yaml
```


### Delete k8s cluster
```
cd $TESTGROUND_ROOT/infra
source ~/.testground
./k8s/delete_ebs.sh
./k8s/delete_efs.sh
./k8s/delete_kops.sh
```


### Download k8s context
** NOTE: testground uses ~/.kube/config, not `$KUBECONFIG` env var**

it may be necessary to symlink your `$KUBECONFIG` file to ~/.kube/config
you can check with `echo $KUBECONFIG` to verify where your contexts are stored

```bash
kops export kubecfg --state $KOPS_STATE_STORE --name=$CLUSTER_NAME --admin
```

### Extra setup steps for metrics
** requires `influxdb` cli installed locally. ex: `brew install influxdb`
```bash
$ kubectl port-forward service/influxdb 8086:8086 &
$ influx
Connected to http://localhost:8086 version 1.8.3
InfluxDB shell version: 1.8.3
> create database testground
> show databases
name: databases
name
----
_internal
testground
> exit
$ kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
# Error from server (ServiceUnavailable): the server is currently unable to handle the request (get nodes.metrics.k8s.io)
```

### SSH to host
```bash
ssh -i ~/.ssh/id_rsa ubuntu@api.testground.domain.xyz
```


### Connect to testground services
**Grafana:** login: `admin:admin`
```bash
$ kubectl port-forward service/prometheus-operator-grafana 3000:80 &
```

**Redis**:
```bash
$ kubectl port-forward service/testground-infra-redis-master 6379:6379 &
```

**InfluxDB**:
```bash
$ kubectl port-forward service/influxdb 8086:8086 &
```

### Run Plans
```bash
cd $TESTGROUND_HOME && nohup testground daemon &
```

#### Restart script:
```
#!/bin/sh
source ~/.testground
PID=$(ps -ef | grep "testground daemon" | grep -v grep | awk {'print $2'})
if [ "$PID" != "" ]; then
  echo "Killing testground daemon $PID"
  kill -9 $PID
fi
if [ $? -eq 0 ]; then
  echo "Removing old data"
  rm -rf nohup.out
  rm -rf data/*
  rm -rf tasks.db/*
  echo "Starting testground daemon"
  nohup testground daemon &
  exit 0
else
  echo "error stopping testground"
fi
```

#### Plans
**height=>10, instances=>1, verify_chain=>false**:
```
testground -vv run single \
  --plan=stacks-node \
  --testcase=blocks \
  --runner=local:docker  \
  --builder=docker:generic  \
  --instances=1 \
  --collect \
  --tp stacks_tip_height=10 \
  --tp verify_chain=false
```

**height=>10, instances=>2, verify_chain=>false**:
```
testground --vv run single \
  --plan=stacks-node \
  --testcase=blocks \
  --runner=cluster:k8s \
  --builder=docker:generic \
  --instances=2 \
  --collect \
  --tp stacks_tip_height=10 \
  --tp verify_chain=false
```

**height=>100, instances=>10, verify_chain=>false, wait=>true**:
```
testground --vv run single \
  --plan=stacks-node \
  --testcase=blocks \
  --runner=cluster:k8s \
  --builder=docker:generic \
  --instances=10 \
  --collect \
  --tp stacks_tip_height=150
```

**height=>150, instances=>30, verify_chain=>true**:
```
testground --vv run single \
  --plan=stacks-node \
  --testcase=blocks \
  --runner=cluster:k8s \
  --builder=docker:generic \
  --instances=30 \
  --collect \
  --tp stacks_tip_height=150
```


**Track progress**
Once a plan is running, run `kubectl get po`
And notice the pods with a name like: `tg-stacks-node-bvs8j6lnf4q15bf1p620-single-0`

the basename is `tg-stacks-node`, followed by task id `bvs8j6lnf4q15bf1p620`, followed by run type `single`, and finally the pod id `0`

to track the progress: `kubectl logs -f tg-stacks-node-bvs8j6lnf4q15bf1p620-single-0`

```
{"ts":1610123915624507763,"msg":"","group_id":"single","run_id":"bvs8j6lnf4q15bf1p620","event":{"message_event":{"message":"registering default http handler at: http://[::]:6060/ (pprof: http://[::]:6060/debug/pprof/)"}}}
{"ts":1610123915624540093,"msg":"","group_id":"single","run_id":"bvs8j6lnf4q15bf1p620","event":{"start_event":{"runenv":{"plan":"stacks-node","case":"blocks","run":"bvs8j6lnf4q15bf1p620","params":{"fork_fraction":"90","num_blocks":"100","sortition_fraction":"90","stacks_tip_height":"150","verify_chain":"true"},"instances":10,"outputs_path":"/outputs/bvs8j6lnf4q15bf1p620/single/0","network":"18.59.0.0/16","group":"single","group_instances":10}}}}
{"ts":1610123915630919220,"msg":"","group_id":"single","run_id":"bvs8j6lnf4q15bf1p620","event":{"message_event":{"message":"waiting for network initialization"}}}
{"ts":1610123918632751896,"msg":"","group_id":"single","run_id":"bvs8j6lnf4q15bf1p620","event":{"message_event":{"message":"network initialisation successful"}}}
{"ts":1610123918634730359,"msg":"","group_id":"single","run_id":"bvs8j6lnf4q15bf1p620","event":{"message_event":{"message":"network initilization complete"}}}
{"ts":1610123918634870993,"msg":"","group_id":"single","run_id":"bvs8j6lnf4q15bf1p620","event":{"message_event":{"message":"127.0.0.1 not in data subnet 18.59.0.0/16, ignoring"}}}
{"ts":1610123918634901350,"msg":"","group_id":"single","run_id":"bvs8j6lnf4q15bf1p620","event":{"message_event":{"message":"100.96.1.8 not in data subnet 18.59.0.0/16, ignoring"}}}
{"ts":1610123918634934881,"msg":"","group_id":"single","run_id":"bvs8j6lnf4q15bf1p620","event":{"message_event":{"message":"detected data network IP: 18.59.24.0/16"}}}
{"ts":1610123918635557526,"msg":"","group_id":"single","run_id":"bvs8j6lnf4q15bf1p620","event":{"message_event":{"message":"my sequence ID: 8"}}}
{"ts":1610123923637763835,"msg":"","group_id":"single","run_id":"bvs8j6lnf4q15bf1p620","event":{"message_event":{"message":"Master started on host address 18.59.128.1"}}}
{"ts":1610123943646538161,"msg":"","group_id":"single","run_id":"bvs8j6lnf4q15bf1p620","event":{"message_event":{"message":"Waiting for node: [Get \"http://localhost:20443/v2/info\": dial tcp 127.0.0.1:20443: connect: connection refused]"}}}
...
{"ts":1610124273711088734,"msg":"","group_id":"single","run_id":"bvs8j6lnf4q15bf1p620","event":{"message_event":{"message":"Stacks block height => 0 :: Burn block height => 11"}}}
```


### Terminate Plan
```bash
testground terminate --runner cluster:k8s
```
