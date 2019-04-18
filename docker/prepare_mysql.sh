#!/usr/bin/env bash

/usr/local/bin/docker-entrypoint.sh mysqld --character-set-server=utf8mb4 &

# wait until MySQL is really available
maxcounter=360

counter=1
while ! mysql --protocol TCP -u"$MYSQL_USER" -p"$MYSQL_PASSWORD" -e "show databases;" > /dev/null 2>&1; do
    sleep 1
    counter=`expr $counter + 1`
    if [ $counter -gt $maxcounter ]; then
        >&2 echo "We have been waiting for MySQL too long already; failing."
        exit 1
    fi;
done