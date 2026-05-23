#!/bin/bash
set -e

ADMIN_ENDPOINT="http://${GARAGE_RPC_HOST}:3903"
RPC_ENDPOINT="${GARAGE_RPC_HOST}:3901"

echo "Получаем Node ID..."

STATUS=$(curl -sf -H "Authorization: Bearer $GARAGE_ADMIN_TOKEN" \
                "$ADMIN_ENDPOINT/v2/GetClusterStatus")

NODE_ID=$(echo "$STATUS" | jq -r '.nodes[0].id')
if [ "$NODE_ID" = "null" ] || [ -z "$NODE_ID" ]; then
    echo "Не удалось получить node_id из ответа: $STATUS"
    exit 1
fi

ROLE=$(echo "$STATUS" | jq -r '.nodes[0].role')
if [ "$ROLE" != "null" ]; then
    echo "Узел уже имеет роль (или layout был применён ранее)."
    exit 0
fi

echo "Инициализируем layout для ноды $NODE_ID..."

RPC_HOST="${NODE_ID}@${RPC_ENDPOINT}"
/garage --rpc-host ${RPC_HOST} --rpc-secret ${GARAGE_RPC_SECRET} layout assign -z dc1 -c 1G $NODE_ID
/garage --rpc-host ${RPC_HOST} --rpc-secret ${GARAGE_RPC_SECRET} layout apply --version 1

echo "Инициализация завершена."
