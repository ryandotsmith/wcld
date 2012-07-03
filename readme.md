# wcld

wc -l (daemon)

Wcld is a process that will listen on TCP $PORT for incoming log data.
Wcld will parse the *crnl* separated data looking for key=value substrings.
When a key=value substring is found, wcld will write the keys and values
to an hstore column in a PostgreSQL database.

## Usage

### Durability

Wcld can be configured for maximum throughput by relaxing it's durability constraints.
There is a buffer for the database writing mechanism that can be configured. A buffer size
of 1 will force wcld to commit each log line that is consumed. A buffer of size 1000 will
allow wcld to write 1000 log lines to the database before commiting the transaction. Of course
if the program crashes before the transaction is commited then the data will be lost.

The durability can be configured by using the `-d` flag. The default is 1.

```bash
$ bin/wcld -d=1000
```

### Queries

Once your applications are draining their logs into a wcld process, you can
begin reporting our your log data.

On a typical web process, the Heroku router will emit the following log message:

```
2012-02-16T06:06:16+00:00 heroku[router]: PUT shushu.herokuapp.com/resources/328408/billable_events/41143162 dyno=web.3 queue=0 wait=0ms service=89ms status=201 bytes=235
```
Notice how the log message contains the service time. This
represents the time it took our web process to respond to the request. We can
quickly group our app's average response time grouped by hour:

```bash
$ heroku pg:psql
```

#### Avg

```sql
SELECT
  date_trunc('hour', time) AS time_group,
  avg((data -> 'service')::interval)
FROM
  log_data
WHERE
  data ? 'service'
  GROUP BY time_group
  ORDER BY time_group
;
```

```
       time_group       |       avg
------------------------+-----------------
 2012-02-13 20:00:00+00 | 00:00:00.074848
 2012-02-13 21:00:00+00 | 00:00:00.076898
 2012-02-13 22:00:00+00 | 00:00:00.073627
 2012-02-13 23:00:00+00 | 00:00:00.075232
 2012-02-14 00:00:00+00 | 00:00:00.075852
 2012-02-14 01:00:00+00 | 00:00:00.073475
 2012-02-14 02:00:00+00 | 00:00:00.072609
 2012-02-14 03:00:00+00 | 00:00:00.073081
```

#### Percentile

```sql
SELECT
  perctile,
  avg(elapsed_time::interval)
FROM (
  SELECT
    data -> 'elapsed_time' as elapsed_time,
    ntile(100) over (order by (data -> 'elapsed_time')) as perctile
  FROM
    log_data
  WHERE
    data -> 'action' = 'find_prev_rec'
    and
    time > now() - '9 minutes'::interval
    and
    expired = false
) x
WHERE
  perctile = 95
GROUP BY perctile
;
```

```
 perctile |       avg
----------+-----------------
       95 | 00:00:00.008944
(1 row)
```

### Indexing

One possible indexing strategy:

```sql
ALTER TABLE log_data ADD COLUMN expired boolean default false;
UPDATE log_data SET expired = 't' where time <= now() - '3 days'::interval;
CREATE INDEX recent_events on log_data (time) where expired = false;
-- use crom to REINDEX each day ??
```


## Deploy to Heroku

* Create app with [Go buildpack](https://gist.github.com/4984b5d9fe9244776197)
* Attach database to app
* Attach route to app
* Point emitter app's at new wcld app

### Create App

```bash
$ git clone git://github.com/ryandotsmith/wcld.git
$ cd wcld
$ heroku create -s cedar --buildpack=https://github.com/kr/heroku-buildpack-go
$ echo "wcld/wcld" >.godir
$ echo "wcld: bin/wcld -f=\"kv\"" > Procfile
$ git add . ; git commit -am "init"
$ git push heroku master
```

### Attach Database

```bash
$ heroku addons:add heroku-postgresql:ika
$ heroku pg:wait
$ heroku pg:promote HEROKU_POSTGRESQL_<COLOR>
$ heroku pg:psql
psql- create extension hstore;
psql- create table events (time timestamptz, data hstore);
psql- create index index_events_by_time on events (time);
```
### Attach Route

```bash
$ heroku routes:create
$ heroku routes:attach tcp://... wcld
```

### Start WCLD Process

```bash
$ heroku scale wcld=2 #can use multiple processes
```

### Use it to drain an emitter app:

```bash
$ heroku drains:add syslog://... -a other-app
```

### Build

```bash
$ cd $GOROOT
$ hg update weekly
$ cd src; ./all.bash
$ cd $GOPATH/src
$ git clone git://github.com/ryandotsmith/wcld.git
$ cd wcld
$ go build .
```

### Test

```bash
$ cd $GOPATH/src/wcld
$ go test .
```
