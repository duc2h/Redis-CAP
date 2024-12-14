# Testing redis-sentinel with CAP theorem

We will check the trade-off between C-A.

Setup: redis-sentinel 1 master, 2 slavers. Run: `docker compose up -d` 

Check the status ` docker exec -it <pod-name> (redis-master, redis-replica1, redis-replica2) redis-cli INFO replication`

## AP scenario
When the network partition happens, we will choose availability. The default setting of redis-sentinel is high-availability => we don't need to change the config.


1. Set value `docker exec -it redis-master redis-cli SET mykey "value1"` => OK
2. Get value from slaver-2 `docker exec -it redis-replica2 redis-cli GET mykey` => "value1"
3. Network partition happens: `docker network disconnect redis-cap_redis-network (remove redis-cap if using docker) redis-replica2` => success
4. Get value from slaver-2 `docker exec -it redis-replica2 redis-cli GET mykey` => "value1"
5. Update value `docker exec -it redis-master redis-cli SET mykey "value2"` => OK
6. Get value from slaver-2 `docker exec -it redis-replica2 redis-cli GET mykey` => "value1" slaver-2 is not up-to-date with other instances. => _**availability but not consistent.**_
7. Get value from slaver-1 `docker exec -it redis-replica1 redis-cli GET mykey` => "value2"
8. Solve network partition: `docker network connect redis-cap_redis-network redis-replica2` => success
9. Get value from slaver-2 `docker exec -it redis-replica2 redis-cli GET mykey` => "value2"


## CA scenario
When the network partition happens, we will choose consistency. We need to change the config.

Get the current config. As you can see, redis-sentinel allows inconsistency data between instances `min-replicas-to-write = 0`. We need to change `min-replicas-to-write = 2` to require all replicas to ack 
when the data has been synced successfully. If not the master will return fail.
```
docker exec -it redis-master redis-cli
127.0.0.1:6379> CONFIG GET min-replicas-to-write
1) "min-replicas-to-write"
2) "0"
127.0.0.1:6379> CONFIG GET min-replicas-max-lag
1) "min-replicas-max-lag"
2) "10"
```

1. Using redis-cli `docker exec -it redis-master redis-cli`
2. Change min-replicas-to-write: `CONFIG SET min-replicas-to-write 2` => OK
3. Network partition happens: `docker network disconnect redis-cap_redis-network redis-replica1` => success
4. Go to another terminal, set value `docker exec -it redis-master redis-cli SET mykey "value1"` => (error) NOREPLICAS Not enough good replicas to write. => _**Redis requires at least 2 replicas ack (require consistency)**_
5. Solve network partition: `docker network connect redis-cap_redis-network redis-replica1` => success
6. Set value: `docker exec -it redis-master redis-cli SET mykey "value1"` => OK
7. Get value: `docker exec -it redis-replica2 redis-cli GET mykey` => "value1"


