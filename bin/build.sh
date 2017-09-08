#!/bin/bash
# $1 = arm or default
# $2 = host to push binary
# $3 = all, bin, res

LOC=$(dirname "$0")
PROJECT_ROOT="/home/dkg/Projects/"
GOROOT="/usr/local/go"
export GOPATH="$PROJECT_ROOT/golang"

# help
if [ "$1" == "help" ]; then
	echo "build.sh"
	exit
fi

# Compile binary.
if [ $1 == "arm" ]; then
	echo "Compiling for ARM"
	GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=1 CC=arm-linux-gnueabihf-gcc \
	CGO_LDFLAGS="-L/home/dkg/arm-lib -Wl,-rpath,/home/dkg/arm-lib" \
 	$GOROOT/bin/go build -o $LOC/main $LOC/main.go 
 else
	echo "Compiling on local machine"
	go build main.go 
fi

# Push binary to remote if previous step completed.
if ! [ -z "$2" ]; then
 	if [ $3 == "all" ]; then
  	echo "Pushing binary to machine $2"
		rsync -avz -e "ssh -o StrictHostKeyChecking=no" --progress main $2:~/walle
  	echo "Pushing resources to machine $2"
		rsync -avz -e "ssh -o StrictHostKeyChecking=no" --progress ../resources $2:~/walle/resources
	elif [ $3 == "res" ]; then
  	echo "Pushing resources to machine $2"
		rsync -avz -e "ssh -o StrictHostKeyChecking=no" --progress ../resources $2:~/walle/resources
	elif [ $3 == "bin" ]; then
  	echo "Pushing binary to machine $2"
		rsync -avz -e "ssh -o StrictHostKeyChecking=no" --progress main $2:~/walle
	fi
fi
exit $?
