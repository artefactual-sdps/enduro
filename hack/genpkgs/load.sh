#!/usr/bin/env sh

set -eux

__dir=$(cd "$(dirname "$0")" && pwd)

go run ${__dir}/main.go > ${__dir}/data.csv

mysql -h127.0.0.1 -uroot -proot123 -P3306 enduro \
	-e "SET GLOBAL local_infile=1;"

mysql -h127.0.0.1 -uroot -proot123 -P3306 enduro \
	-e "DELETE FROM sip;"

mysql -h127.0.0.1 -uroot -proot123 -P3306 --local-infile=1 enduro \
	-e "LOAD DATA LOCAL INFILE '${__dir}/data.csv' INTO TABLE sip FIELDS TERMINATED BY ','"

mysql -h127.0.0.1 -uroot -proot123 -P3306 enduro \
	-e "SET GLOBAL local_infile=0;"
