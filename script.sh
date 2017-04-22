#!/bin/sh

echo "This is a really important script"
echo "This script counts all the way to 10!" 

echo "$COOL"

for i in {1..10} ; do 
  echo "$i, $(date)";
  sleep 1
done

echo "I have finished!"

exit 1
