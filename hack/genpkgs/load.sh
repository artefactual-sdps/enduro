#!/usr/bin/env sh

set -eux

__dir=$(cd "$(dirname "$0")" && pwd)

# Generate SIPs
go run ${__dir}/main.go sip > ${__dir}/data.csv

mysql -h127.0.0.1 -uroot -proot123 -P3306 enduro \
	-e "SET GLOBAL local_infile=1;"

mysql -h127.0.0.1 -uroot -proot123 -P3306 enduro \
	-e "DELETE FROM sip;"

mysql -h127.0.0.1 -uroot -proot123 -P3306 --local-infile=1 enduro \
	-e "LOAD DATA LOCAL INFILE '${__dir}/data.csv' INTO TABLE sip FIELDS TERMINATED BY ','"

# Generate AIPs
go run ${__dir}/main.go aip > ${__dir}/data.csv

mysql -h127.0.0.1 -uroot -proot123 -P3306 enduro_storage \
	-e "SET GLOBAL local_infile=1;"

mysql -h127.0.0.1 -uroot -proot123 -P3306 enduro_storage \
	-e "DELETE FROM aip;"

mysql -h127.0.0.1 -uroot -proot123 -P3306 --local-infile=1 enduro_storage \
	-e "LOAD DATA LOCAL INFILE '${__dir}/data.csv' INTO TABLE aip FIELDS TERMINATED BY ','"

mysql -h127.0.0.1 -uroot -proot123 -P3306 enduro_storage \
	-e "SET GLOBAL local_infile=0;"
