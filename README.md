# Gateway
Service to receive all requests and redirect logic to other pieces from system.

# Getting Started

First step to run project, is to run docker command to build database. This database will need config from [SQL file from setup](./_setup/schema.sql) folder.

```shell
docker run --name db_name -d -ti -p 3306:3306 -e MYSQL_ROOT_PASSWORD=password -v "path\to\db:/var/lib/mysql" mysql:8.0 --sql-mode="TRADITIONAL" --bind-address="0.0.0.0" --default_time_zone="-03:00"
```