#!/usr/bin/env bash

echo "starting..."

# start mysql
/usr/local/bin/docker-entrypoint.sh mysqld &

/root/docker/wait_mysql.sh

# start go app
/root/hlc