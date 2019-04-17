#!/usr/bin/env bash

echo "starting..."

# start mysql
/usr/local/bin/docker-entrypoint.sh mysqld --character-set-server=utf8mb4 &

/root/docker/wait_mysql.sh

# start go app
/root/bin/loader
/root/bin/hlc