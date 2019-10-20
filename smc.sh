#!/bin/bash

CMD=$1

# usage
USAGE="Usage:
    $0 check
    $0 load"

# check if smc support is in the kernel
function check_kernel {
	# check /proc/config.gz
	echo -n "Kernel support for SMC: "
	if zgrep "^CONFIG_SMC=y" /proc/config.gz > /dev/null; then
		echo "yes"
	elif zgrep "^CONFIG_SMC=m" /proc/config.gz > /dev/null; then
		echo "module"
	else
		echo "no"
	fi
}

# load the smc kernel module
function load_module {
	echo -n "Loading SMC kernel module: "
	if modprobe smc; then
		echo "ok."
	else
		echo "failed."
	fi
}

# run commands with other command line arguments
case "$CMD" in
	"check")
		check_kernel
		;;
	"load")
		load_module
		;;
	*)
		echo "$USAGE"
		;;
esac
