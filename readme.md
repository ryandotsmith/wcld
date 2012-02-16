# wcld

wc -l (daemon)

Wcld is a process that will listen on TCP $PORT for incomming syslog data.
Wcld will parse the *crnl* sperated data looking for key=value substrings.
When a key=value substring is found, wcld will write the keys and values
to an hstore column in a postgres database.


## Deploy to Heroku

Setup wcld:

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
$ heroku addons:add heroku-postgresql:ika version=9.1
$ heroku pg:psql
# create table log_data (id bigserial, time timestamptz, data hstore);
$ heroku scale wcld=1
$ heroku routes:attach `heroku routes:create` wcld.1
```

## Build Instructions

```bash
$ cd $GOROOT
$ hg update weekly.2012-02-07
$ cd src; ./all.sh
$ cd $GOPATH/src
$ git clone git://github.com/ryandotsmith/wcld.git
$ go install wcld
```

## Testing Instructions

```bash
$ go test wcld
```
