# smc

Documentation, notes, and scripts for Shared Memory Communication (SMC) in
Linux.

## Kernel Support

Make sure your kernel is configured with `CONFIG_SMC`. If your kernel provides
the `/proc/config.gz` file, you can use it to check if your kernel has been
compiled with SMC support with:

```console
$ zgrep "CONFIG_SMC" /proc/config.gz
```

or you can use the [smc.sh](smc.sh) script in this folder with:

```console
$ ./smc.sh check
```

If SMC support is configured as a module, you need to load the `smc` kernel
module as root with:

```console
# modprobe smc
```

or with the `smc.sh` script:

```console
# ./smc.sh load
```

## Devices

SMC requires, depending on the SMC variant, a different set of devices:

SMC-R uses three devices: a *handshake device*, a *RoCE IB device*, and a *RoCE
net device*. The handshake device is a network interface (e.g., `eth0`) used
for the initial TCP connection setup and CLC handshake (see below) between two
SMC peers. For example, a SMC server socket application listens on this device.
The RoCE IB device is a RDMA over Converged Ethernet (RoCE) Infiniband (IB)
device used to transfer the SMC-R traffic after the SMC connection has been
established. The RoCE IB device transfers the traffic over an Ethernet device.
This Ethernet device is the RoCE net device. The handshake device and RoCE net
device can be the same device.

SMC-D uses two devices: a *handshake device* and an *ISM device*. The handshake
device is the same as in the case of SMC-R. The ISM device is an Internal
Shared Memory (ISM) device that is used to transfer the SMC-D traffic after the
SMC connection has been established. ISM devices are only available on the s390
architecture.

## Device Configuration

In order to use SMC, the devices that are used for a SMC connection need to be
configured.

Configure the handshake device (SMC-R and SMC-D):
* make sure the network interface is up
* if you use VLAN, configure the right VLAN on the network interface
* Configure your IPv4 and/or IPv6 address(es) on the network interface

Configure the RoCE net device (SMC-R only):
* make sure the network interface is up
* if you use VLAN, configure the right VLAN on the network interface
* configure an IPv6 address on the network interface

After the previous steps, the RoCE IB device should also be ready to use. ISM
devices do not need extra configuration.

Example with handshake device `eth0` and RoCE net device `eth1`, without VLAN,
with IP address `192.168.1.23` on the handshake device, and with IPv6 stateless
address autoconfiguration:

```console
# # configure handshake device eth0:
# ip link set eth0 up
# ip address add 192.168.1.23/24 dev eth0
# # configure RoCE net device eth1:
# ip link set eth1 up
```

## Pnetid Configuration

SMC uses a so-called Physical Network ID (pnetid) to map a handshake device to
a RoCE IB device or to an ISM device. Thus, you need to make sure that your
handshake device and your RoCE IB device or your ISM device have the same
pnetid. You can use the tool [pnetctl](https://github.com/hwipl/pnetctl) to
read the currently configured pnetids and to configure the pnetids of your
devices.

## SMC Socket Programming

See the folder [socket](socket/) for information on SMC socket programming.

## CLC Handshake

See the folder [handshake](handshake/) for information on SMC's CLC handshake.
