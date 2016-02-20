#!/bin/bash

for i in mysql/*.sql; do
  cat $i | sed "s/\`/\"/g" | ./mysql2sqlite.sh > sqlite3/${i##*/}
done
