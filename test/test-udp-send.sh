#!/bin/bash


function send {
    echo "==== $1"
    echo "$1" > /dev/udp/localhost/7777
}


function once {
	send "\$bogous,,*47"

	send "\$GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*47"
	sleep 2

	send "\$GPHDT,274.07,T*03"
	sleep 1

	sleep 2
	send "\$GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*47"

	sleep 1
	send "\$GPHDT,274.07,T*03"

	sleep 1
	#send "\$GPTHS,338.01,A*36"
	send "\$GPTHS,338.01,A*0E"

}


#once
#exit 1

while [[ 1 ]]; do
	once
	sleep 1
done

