#!/usr/bin/env bash

echo checking topics created

topics="payment-requested payment-verified payment-failed transaction-requested transaction-completed transaction-failed"

for topic in $topics; do
    # /opt/bitnami/kafka/bin/
    kafka-topics.sh --create --if-not-exists --bootstrap-server "$KAFKA_BROKER" --topic "$topic"
done;

echo 'done'