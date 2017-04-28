#!/bin/sh
set -e 
set -x

if [ ! -f "gaffer.db" ]; then
  cat > /tmp/init.sql << __EOF__ 
CREATE TABLE users (id integer not null primary key, token text);
CREATE TABLE clusters (id text not null primary key, user int, data text);
__EOF__
sqlite3 gaffer.db < /tmp/init.sql
fi

exec $@
