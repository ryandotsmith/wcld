# wcld

wc -l (daemon)

Wcld is a process that will listen on TCP $PORT for incomming syslog data.
Wcld will parse the *crnl* sperated data looking for key=value substrings.
When a key=value substring is found, wcld will write the keys and values
to an hstore column in a postgres database.

## Usage

Once your applications are draining their logs into a wcld process, you can
begin reporting our your log data.

On a typical web process, the Heroku router will emmit the folling log message:

```
2012-02-16T06:06:16+00:00 heroku[router]: PUT shushu.herokuapp.com/resources/328408/billable_events/41143162 dyno=web.3 queue=0 wait=0ms service=89ms status=201 bytes=235
```
Notice how the log message contains the service time. This
represents the time it took our web process to respond to the request. We can
quickly group our app's average response time grouped by hour:

```bash
$ heroku pg:psql
```

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

## Setup

Download the binary from github and run it on Heroku using the null-buildpack.

### Deploy to Heroku

```bash
$ mkdir wcld
$ cd wcld
$ mkdir bin
$ curl -L -o wcld.tar.gz "https://github.com/downloads/ryandotsmith/wcld/wcld.tar.gz"
$ tar -xvzf wcld.tar.gz -C bin/
$ rm *.tar.gz
$ echo "wcld: bin/wcld" >> Procfile
$ git init
$ git add .
$ git commit -am "init"
$ heroku create -s cedar --buildpack=git://github.com/ryandotsmith/null-buildpack.git
$ git push heroku master
$ heroku addons:add heroku-postgresql:ika --version=9.1
$ heroku pg:psql
- create table log_data (id bigserial, time timestamptz, data hstore);
$ heroku scale wcld=1
$ heroku routes:attach `heroku routes:create` wcld.1
```

### Build

```bash
$ cd $GOROOT
$ hg update weekly.2012-02-07
$ cd src; ./all.sh
$ cd $GOPATH/src
$ git clone git://github.com/ryandotsmith/wcld.git
$ go install wcld
```

### Test

```bash
$ go test wcld
```
