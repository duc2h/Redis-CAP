# Testing consistency level in Scylladb cluster.

We will test the consistency level in Scylladb cluster. Scylladb has clustering (the same architect as CassandraDB cluster). It has several levels, 
but we check 3 famous levels: ONE, ALL, QUORUM [ref](https://opensource.docs.scylladb.com/stable/cql/consistency.html). Scylladb allows developers 
to choose the CL based on their logic, CL can be flexible with any commands from the developers.

Follow the instruction: https://github.com/scylladb/scylla-code-samples/blob/master/mms/docker-compose.yml

## Set up
In this example, I set up 3 nodes for the Scylla cluster. 

Execute: `docker compose up -d`

Check node status:
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