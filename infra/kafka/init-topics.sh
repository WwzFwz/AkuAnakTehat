#!/bin/sh
set -eu
for topic in bnpb.hazard-events.v1 bnpb.hazard-events.v1.dlq; do
  /opt/kafka/bin/kafka-topics.sh --bootstrap-server kafka:9092 --create --if-not-exists \
    --topic "$topic" --partitions 1 --replication-factor 1 \
    --config retention.ms=604800000 --config cleanup.policy=delete --config max.message.bytes=5242880
  /opt/kafka/bin/kafka-configs.sh --bootstrap-server kafka:9092 --alter \
    --entity-type topics --entity-name "$topic" --add-config max.message.bytes=5242880
done
