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

or you can use the `smc.sh` script in this folder with:

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
