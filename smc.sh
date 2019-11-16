#!/bin/bash

CMD=$1

# usage
USAGE="Usage:
    $0 check
    $0 load
    $0 rxe load
    $0 rxe add <net_dev>
    $0 rxe del <rxe_dev>"

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

# run rxe commands
function run_rxe {
	rxe_params=/sys/module/rdma_rxe/parameters
	cmd=$1
	dev=$2

	case "$cmd" in
		"load")
			echo "Loading rdma_rxe kernel module"
			modprobe rdma_rxe
			;;
		"add")
			echo "Adding $dev to rxe devices."
			echo "$dev" > $rxe_params/add
			;;
		"del")
			echo "Removing $dev from rxe devices."
			echo "$dev" > $rxe_params/remove
			;;
		*)
			echo "$USAGE"
			;;
	esac
}

# run commands with other command line arguments
case "$CMD" in
	"check")
		check_kernel
		;;
	"load")
		load_module
		;;
	"rxe")
		run_rxe "$2" "$3"
		;;
	*)
		echo "$USAGE"
		;;
esac
