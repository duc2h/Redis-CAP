# Testing consistency level in Scylladb cluster.

We will test the consistency level in Scylladb cluster. Scylladb has clustering (the same architect as CassandraDB cluster). It has several levels, 
but we check 3 famous levels: ONE, ALL, QUORUM [ref](https://opensource.docs.scylladb.com/stable/cql/consistency.html). Scylladb allows developers 
to choose the CL based on their logic, CL can be flexible with any commands from the developers.

Follow the instruction: https://github.com/scylladb/scylla-code-samples/blob/master/mms/docker-compose.yml

## Set up
In this example, I set up 3 nodes for the Scylla cluster. 

Execute: `docker compose up -d`

Check node status:

UN is up and normal. DN is down.
```
docker exec -it scylla-node1 nodetool status
Emulate Docker CLI using podman. Create /etc/containers/nodocker to quiet msg.
Datacenter: DC1
===============
Status=Up/Down
|/ State=Normal/Leaving/Joining/Moving
-- Address    Load      Tokens Owns Host ID                              Rack 
UN 10.89.7.11 386.82 KB 256    ?    5332d264-eebc-498d-b2c9-bd33c0c04334 Rack1
DN 10.89.7.12 398.85 KB 256    ?    adf35932-14f0-47ab-bf54-599ce8a322a7 Rack1
UN 10.89.7.13 419.66 KB 256    ?    6160373a-a1e7-4262-bff5-aa73008943b4 Rack1
```

Use cql:
```
docker exec -it scylla-node1 cqlsh
```

Create KEYSPACE:
```
cqlsh> CREATE KEYSPACE test_keyspace WITH replication = {'class': 'NetworkTopologyStrategy', 'DC1': 3};


cqlsh> DESCRIBE KEYSPACE test_keyspace
CREATE KEYSPACE test_keyspace WITH replication = {'class': 'org.apache.cassandra.locator.NetworkTopologyStrategy', 'DC1': '3'} AND durable_writes = true;
```

Use KEYSPACE:
```
cqlsh> USE test_keyspace;
```

Create table schema:
```
cqlsh:test_keyspace> CREATE TABLE test_table (
                     id UUID PRIMARY KEY,
                     value TEXT);
```

## Explain: How ScyllaDB works.
When a request (read/write) is sent to Scylla cluster, the node that receives this request is called Coordinator-node (any node can be Coordinator). 
If CL > ONE, the coordinator will get & check the consistency value from other nodes (depending on CL). 

## Consistency level is ONE
Write: Coordinator-node won't need to wait for ack from other nodes.

Read: Coordinator-node won't check the data from other nodes.

_**When developers use this CL => they choose A instead of C.**_

1. Use CL:
```
cqlsh:test_keyspace> CONSISTENCY ONE;
Consistency level set to ONE.
```
2. Insert data
```
cqlsh:test_keyspace> INSERT INTO test_table (id, value) VALUES (uuid(), 'Test Value');
```

3. Read data
```
select * from test_table;

 id                                   | value
--------------------------------------+------------
 d71ed770-8842-49ea-a067-a1c5383a2522 | Test Value

(1 rows)
```

## Consistency level is ALL
Write: Coordinator-node requires the ack from all other nodes (based on the keyspace's datacenter.)

Read: Coordinator-node checks the value from all other nodes (based on the keyspace's datacenter.)

_**When developers use this CL => they choose C instead of A.**_
1. Use CL
```
cqlsh:test_keyspace> CONSISTENCY ALL;
Consistency level set to ALL.
```

2. Network partition happens. Disconnect node2, node1 and node3 cannot communicate with node2.
```
docker network disconnect scylla-cluster_scylla-net scylla-node2
```

```
docker exec -it scylla-node1 nodetool status
Datacenter: DC1
===============
Status=Up/Down
|/ State=Normal/Leaving/Joining/Moving
-- Address    Load      Tokens Owns Host ID                              Rack 
UN 10.89.7.17 397.41 KB 256    ?    5c7b6034-500a-4875-a9ee-0fd16c88f60c Rack1
UN 10.89.7.19 407.90 KB 256    ?    d34805d3-887b-4ec7-91df-e4aee4ec1837 Rack1
DN 10.89.7.21 489.25 KB 256    ?    00e304af-3629-48a3-82b6-eaf943de1cac Rack1
```
```
docker exec -it scylla-node2 nodetool status
Datacenter: DC1
===============
Status=Up/Down
|/ State=Normal/Leaving/Joining/Moving
-- Address    Load      Tokens Owns Host ID                              Rack 
DN 10.89.7.17 397.41 KB 256    ?    5c7b6034-500a-4875-a9ee-0fd16c88f60c Rack1
DN 10.89.7.19 407.90 KB 256    ?    d34805d3-887b-4ec7-91df-e4aee4ec1837 Rack1
UN 10.89.7.21 489.25 KB 256    ?    00e304af-3629-48a3-82b6-eaf943de1cac Rack1
```

3. Write failed.
```
cqlsh:test_keyspace> INSERT INTO test_table (id, value) VALUES (uuid(), 'CL ALL FAIL');
NoHostAvailable: ('Unable to complete the operation against any hosts', {<Host: 10.89.7.17:9042 DC1>: Unavailable('Error from server: code=1000 [Unavailable exception] message="Cannot achieve consistency level for cl ALL. Requires 3, alive 2" info={\'consistency\': \'ALL\', \'required_replicas\': 3, \'alive_replicas\': 2}')})
```

4. Read failed.
```
cqlsh:test_keyspace> select * from test_table;
NoHostAvailable: ('Unable to complete the operation against any hosts', {<Host: 10.89.7.17:9042 DC1>: Unavailable('Error from server: code=1000 [Unavailable exception] message="Cannot achieve consistency level for cl ALL. Requires 3, alive 2" info={\'consistency\': \'ALL\', \'required_replicas\': 3, \'alive_replicas\': 2}')})
```

## Consistency level is QUORUM
The information of QUORUM: [ref](https://en.wikipedia.org/wiki/Quorum_(distributed_computing))

It allows some nodes to crash, network partition,... by the formula: `alive nodes >= N/2 + 1` N is the total nodes.

Write: Coordinator-node requires the ack from other nodes. If the ack >= `N/2+1` => the request success, if not => the request fail.

Read: Coordinator-node checks the data from other nodes. If the ack >= `N/2+1` => the request success, if not => the request fail.

_**ScyllaDB helps flexible by this CL. It provides C, A, and performance if developers use QUORUM**_

1. Use CL
```
cqlsh:test_keyspace> CONSISTENCY QUORUM;
Consistency level set to QUORUM.
```

2. Insert data
```
cqlsh:test_keyspace> INSERT INTO test_table (id, value) VALUES (uuid(), 'Quorum Value');
```

3. Read data
```
 cqlsh:test_keyspace> SELECT * FROM test_table;

 id                                   | value
--------------------------------------+--------------
 cb4de863-6f05-488a-a4d5-02acb27eaac4 | Quorum Value
 d71ed770-8842-49ea-a067-a1c5383a2522 |   Test Value

(2 rows)
```

4. Network partition happens. Disconnect node2 and node3.
5. Insert fail.
```
 cqlsh:test_keyspace>  INSERT INTO test_table (id, value) VALUES (uuid(), 'Quorum Value Fail');
NoHostAvailable: ('Unable to complete the operation against any hosts', {<Host: 10.89.7.17:9042 DC1>: Unavailable('Error from server: code=1000 [Unavailable exception] message="Cannot achieve consistency level for cl QUORUM. Requires 2, alive 1" info={\'consistency\': \'QUORUM\', \'required_replicas\': 2, \'alive_replicas\': 1}')})
```
6. Read fail.
```
cqlsh:test_keyspace> SELECT * FROM test_table;
NoHostAvailable: ('Unable to complete the operation against any hosts', {<Host: 10.89.7.17:9042 DC1>: Unavailable('Error from server: code=1000 [Unavailable exception] message="Cannot achieve consistency level for cl QUORUM. Requires 2, alive 1" info={\'consistency\': \'QUORUM\', \'required_replicas\': 2, \'alive_replicas\': 1}')})
```

## Testing with partition_key
In this example, we have 4 nodes in a datacenter. The token ring will be: [ref](https://cassandra.apache.org/doc/4.1/cassandra/architecture/dynamo.html#consistent-hashing-using-a-token-ring)
```
Node1: Token range 0–25%.
Node2: Token range 26–50%.
Node3: Token range 51–75%.
Node4: Token range 76–100%.
```

### Choose coordinator node: 

When a write request is sent to Scylladb, Scylladb will hash the primary-key (partition_key), and then it will store the data to the node based on node's token ring. 


### Choose replica nodes:
Scylladb use clockwise concept to choose replica nodes (the next node of the coordinator-node). Example:
```
id = 10 => partition_key = 27 => coordinator-node is node2
2 replica nodes are node3 and node4, respectively.
```

The question is: If I create a keyspace with 3 replica_factors in the datacenter has 4 nodes, node4 is down. What happens 
if i insert a new record with consistency-level = ALL?

=> The insert statement can be fail or success depend on its partition_key. 

Success: if the partition_key belongs to node1 => replica nodes: node2 and node3 => all node alive.
Fail: if the partition_key belongs to node2, node3 or node4. Example: partition_key belongs to node2 => replica nodes: 
node3, node4 => node3 ack, but node4 not ack (crashing) => Failed.

We can check the partition_key belongs to which nodes:

`docker exec -it scylla-node1 nodetool getendpoints <keyspace-name> <table-name> <primary_key-value>`
```
docker exec -it scylla-node1 nodetool getendpoints test_keyspace test_table '1765933e-d5e2-4e2a-b749-59aa0d0a3cc7'

10.89.7.64
10.89.7.62
10.89.7.63
```

## KEYSPACE with multiple datacenter
We can create a keyspace with multiple datacenter.
```
 CREATE KEYSPACE test_keyspace
    WITH replication = {
    'class': 'NetworkTopologyStrategy',
    'DC1': 3,
    'DC2': 3
    };
cqlsh> DESCRIBE test_keyspace;

CREATE KEYSPACE test_keyspace WITH replication = {'class': 'org.apache.cassandra.locator.NetworkTopologyStrategy', 'DC1': '3', 'DC2': '3'} AND durable_writes = true;
```

```
Benefits of Multi-Datacenter Keyspaces: 
	1.	Fault Tolerance: If an entire datacenter goes down, the cluster can still operate with the remaining datacenters.
	2.	Disaster Recovery: Ensures data is available in another datacenter in case of catastrophic failure.
	3.	Low Latency: Queries can be served locally within a datacenter by setting the consistency level to LOCAL_QUORUM.
	4.	Geographical Distribution: Place replicas closer to users in geographically distributed systems.
```

```
Consistency Levels in Multi-Datacenter Keyspaces: 
When working with multi-datacenter keyspaces, you can control where and how queries are executed using consistency levels:
	1.	LOCAL_QUORUM: (3/2 + 1) => 2 nodes need alive
	    Requires a quorum (majority) of replicas in the local datacenter to respond.
	    Ensures low-latency queries by avoiding cross-datacenter communication.
	2.	QUORUM: (6/2 + 1) => 4 nodes need alive
	    Requires a quorum of replicas across all datacenters to respond.
	    Guarantees stronger consistency but incurs higher latency due to cross-datacenter communication.
	3.	ALL: requires 6 nodes alive
        Requires all replicas in all datacenters to respond.
	    Provides the strongest consistency but is slower and less resilient to failures.
```

## Consistency in ScyllaDB
Scylladb is using [hinted-handoff](https://opensource.docs.scylladb.com/stable/architecture/anti-entropy/hinted-handoff.html) 
to synchronized data to down-nodes when they rejoin the cluster (short term: because the hint-storage-retention 
just 3 hours.)

Scylladb combines [anti-entropy read repair](https://opensource.docs.scylladb.com/stable/architecture/anti-entropy/read-repair.html) and [anti-entropy repair](https://opensource.docs.scylladb.com/stable/operating-scylla/procedures/maintenance/repair.html) to achieve consistency

```
Scenario: for hint-handoff

	•	Node A, Node B, and Node C form a cluster.
	•	Node B goes down, and Node A is the coordinator for a write intended for Node B.

Steps:

	1.	Write Operation:
		A write request arrives at Node A for key user123.
		Node A stores the data in its own replica and sends it to Node C.
		Node B is down, so Node A creates a hint for Node B and stores it locally.
	2.	Node B Recovers:
		Gossip protocol detects Node B is back online.
		Node A retrieves the hint for Node B from its hints directory.
	3.	Hint Delivery:
		Node A sends the stored hint to Node B in batches.
		Node B applies the hints and becomes consistent with the cluster.
	4.	Hint Deletion:
		Node A deletes the hint for Node B after receiving an acknowledgment.

```

We can monitor the hints by `promethues metrics`, 
1. Disconnect node3: `docker network disconnect scylla-cluster_scylla-net scylla-node3`
2. Checking node3 is down: 
```
docker exec -it scylla-node1 nodetool status                                                                      
Datacenter: DC1
===============
Status=Up/Down
|/ State=Normal/Leaving/Joining/Moving
-- Address     Load      Tokens Owns Host ID                              Rack 
UN 10.89.7.119 727.80 KB 256    ?    f7df5d27-1b80-4d95-b69f-c1fe0d369c3a Rack1
UN 10.89.7.121 814.32 KB 256    ?    0398ec03-c745-4bf2-9edd-915836389318 Rack1
DN 10.89.7.126 912.99 KB 256    ?    57c0eca5-edc9-4edd-88b3-eb0eb3088ba2 Rack1
```

3. Insert two new records with CL = ONE|QUORUM
4. Checking the metrics, execute container: `docker exec -it scylla-node1 bash`
```
curl http://127.0.0.1:9180/metrics | grep hints
...
# HELP scylla_hints_manager_written Number of successfully written hints.
# TYPE scylla_hints_manager_written counter
scylla_hints_manager_written{shard="0"} 2 // there are two missing record.
```

5. Restart node3 and join to the cluster
6. Checking the metrics in node1
```
curl http://127.0.0.1:9180/metrics | grep hints

...
# HELP scylla_hints_manager_sent_bytes_total The total size of the sent hints (in bytes)
# TYPE scylla_hints_manager_sent_bytes_total counter
scylla_hints_manager_sent_bytes_total{shard="0"} 470
# HELP scylla_hints_manager_sent_total Number of sent hints.
# TYPE scylla_hints_manager_sent_total counter
scylla_hints_manager_sent_total{shard="0"} 2 // node1 has synced two missing data to the down-node.
# HELP scylla_hints_manager_size_of_hints_in_progress Size of hinted mutations that are scheduled to be written.
# TYPE scylla_hints_manager_size_of_hints_in_progress gauge
scylla_hints_manager_size_of_hints_in_progress{shard="0"} 0.000000
# HELP scylla_hints_manager_written Number of successfully written hints.
# TYPE scylla_hints_manager_written counter
scylla_hints_manager_written{shard="0"} 2
100  239k    0  239k    0     0  7760k      0 --:--:-- --:--:-- --:--:-- 7972k
```

7. Checking the metrics in node3
```
curl http://127.0.0.1:9180/metrics | grep hints
...
# HELP scylla_hints_manager_written Number of successfully written hints.
# TYPE scylla_hints_manager_written counter
scylla_hints_manager_written{shard="0"} 0
# HELP scylla_storage_proxy_replica_received_hints_bytes_total total size of hints and MV hints received by this node
# TYPE scylla_storage_proxy_replica_received_hints_bytes_total counter
scylla_storage_proxy_replica_received_hints_bytes_total{scheduling_group_name="streaming",shard="0"} 470
# HELP scylla_storage_proxy_replica_received_hints_total number of hints and MV hints received by this node
# TYPE scylla_storage_proxy_replica_received_hints_total counter
scylla_storage_proxy_replica_received_hints_total{scheduling_group_name="streaming",shard="0"} 2 // node3 has received the synced.
100  224k    0  224k    0     0  18.8M      0 --:--:-- --:--:-- --:--:-- 19.9M
```

## Interact with Go

```
docker build -t my-go-app .

docker stop my-go-app-container

docker rm my-go-app-container

docker run --name my-go-app-container --network scylla-cluster_scylla-net -p 8080:8080 my-go-app
```

Result:
```
docker run --name my-go-app-container --network scylla-cluster_scylla-net -p 8080:8080 my-go-app
Emulate Docker CLI using podman. Create /etc/containers/nodocker to quiet msg.
2024-12-15T05:35:06.368Z        INFO    Displaying Results:
2024-12-15T05:35:06.375Z        INFO    Inserting Mike
2024-12-15T05:35:06.379Z        INFO    Displaying Results:
2024-12-15T05:35:06.384Z        INFO            Mike Tyson, 1515 Main St, http://www.facebook.com/mtyson
2024-12-15T05:35:06.384Z        INFO    Deleting Mike
2024-12-15T05:35:06.388Z        INFO    Displaying Results:

```
