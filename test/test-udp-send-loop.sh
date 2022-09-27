#!/bin/bash


while [ 1 ]; do

while read p; do
  echo "$p" > /dev/udp/localhost/7777
  echo "$p"
  sleep 0.5
done <test/loop1.nmea

done