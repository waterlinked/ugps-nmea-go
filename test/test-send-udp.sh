#!/bin/bash


function send {
    echo "==== $1"
    echo "$1" > /dev/udp/localhost/7777
}


function slumber {
	#sleep 1
	sleep 0.1
}

function once {
	send "\$bogous,,*47"
	slumber

	send "\$GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*47"
	slumber

	send "\$GPHDT,274.07,T*03"
	slumber

	send "\$GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*47"
	slumber

	send "\$GPHDT,274.07,T*03"
	#send "\$HCHDM,277.19,M*13"
	slumber

	#send "\$GPTHS,338.01,A*36"
	send "\$GPTHS,338.01,A*0E"
	slumber

	#send "\$GPTHS,338.01,A*36"
	send "\$HCHDM,276.71,M*1C"
	slumber

	send "\$GPHDT,274.07,T*03"
	slumber

}


#once
#exit 1

while [[ 1 ]]; do
	once
	slumber
done
