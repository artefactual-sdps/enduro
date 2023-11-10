#!/usr/bin/env sh

set -eux

__dir=$(cd "$(dirname "$0")" && pwd)

go run ${__dir}/main.go > ${__dir}/data.csv

mysql -h127.0.0.1 -uroot -proot123 -P7450 enduro \
	-e "DELETE FROM package;"

mysql -h127.0.0.1 -uroot -proot123 -P7450 enduro \
	-e "LOAD DATA LOCAL INFILE '${__dir}/data.csv' INTO TABLE package FIELDS TERMINATED BY ','"
