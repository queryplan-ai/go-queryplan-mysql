# Performance Tests

This is a set of tests that are run to ensure that any performance differences between the native go mysql driver and the go-mysql-queryplan drivers are expected and within acceptable range.

To run these tests. run the following from the root of this repo:

```
make performance
```

Some tips to get reliable results:

1. Your mysql server should be local. Consider docker: 

```
docker run \
    --name qpperformance \
    -e MYSQL_USER=qpperformance \
    -e MYSQL_PASSWORD=qpperformance \
    -e MYSQL_DATABASE=qpperformance \
    -e MYSQL_ROOT_PASSWORD=qpperformance \
    -d \
    -p 33306:3306 \
    mysql:latest 

```


2. Export your connection string: 

```
export QUERYPLAN_DB_URI="qpperformance:qpperformance@tcp(localhost:33306)/qpperformance"
```



