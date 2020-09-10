## Example

Run `N` nodes kafka cluster and `M` nodes consumer:

```
> docker-compose up --scale kafka0=3 --scale kafka-consumer=2 kafka-consumer
```

Kafka is configured to listen on docker internal interface, port `29029`, 
announcing a container's internal IP to the clients.

See `listeners` and `advertised.listeners` settings in `docker-compose.yml`.

Produce messages to kafka:

```
> docker container run \
    --network test-kafka_default \
    --rm -ti \
    wurstmeister/kafka:2.12-2.2.1 \
        "echo 'test message 1' | kafka-console-producer.sh --broker-list kafka0:29092 --topic go-topic0"
```

Or the same with kafkacat:

```
> docker container run --rm -ti --network test-kafka_default edenhill/kafkacat:1.6.0 -b kafka0:29092 -P -t go-topic-0
test message 1
[Press Ctrl-D]
```

Note, `--network` flag to run a container on the same bridge network as kafka cluster above.

Inspect kafka cluster:

```
> docker container run --rm -ti --network test-kafka_default edenhill/kafkacat:1.6.0 -b kafka0:29092 -L
```

Gracefully close consumers nodes:

```
> docker-compose kill -s TERM kafka-consumer
```

## Further Reading

https://rmoff.net/2018/08/02/kafka-listeners-explained/
https://github.com/wurstmeister/kafka-docker/wiki/Connectivity
