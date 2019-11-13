# Handshake

This page contains information gathered from
[RFC7609](https://tools.ietf.org/html/rfc7609) and the [SMC implementation in
the Linux kernel](https://github.com/torvalds/linux/tree/master/net/smc) about
the TCP and CLC handshakes during the SMC connection setup.

## TCP Handshake

When establishing a TCP connection, SMC capable peers indicate their
willingness to use SMC for this connection with an experimental TCP option in
the SYN and SYN/ACK packets.

SMC's Experimental TCP Option:
* Kind: 254 (experimental option)
* Length: 6 bytes
* ExID: 0xE2D4C3D9 ("SMCR" in EBCDIC/IBM-1047 encoding)

If both peers sent this option during the connection establishment, they
perform a CLC handshake over the established TCP connection.

## CLC Handshake

During the CLC handshake, SMC peers exchange parameters that are required to
establish a SMC-R connection over a RoCE IB device or a SMC-D connection over
an ISM device.

The regular CLC handshake is a three-way handshake that consists of the
following messages:
* Proposal
* Accept
* Confirm

In case of errors during the handshake, an additional message is used to abort
the SMC connection setup:
* Decline

See [RFC7609](https://tools.ietf.org/html/rfc7609) and the following sections
for more info on these messages. Also, you can use the tool
[smc-clc](https://github.com/hwipl/smc-clc) to capture and show CLC handshakes
including all these messages on a network interface.

### Proposal

With a CLC Proposal message, the client signals to the server the willingness
to use SMC-R, SMC-D or both for this connection. So, the message contains
information that helps the server to determine if a SMC connection is possible.

CLC proposal message:
* Eyecatcher: "SMCR" or "SMCD" in EBCDIC/IBM-1047 encoding
* Type: 1 (Proposal)
* Length: variable
* Version: 1
* Path: SMC-R or SMC-D or both
* Client's Peer ID
* Client's preferred SMC-R GID
* Client's preferred RoCE MAC address
* IP Area Offset
* Client's SMC-D GID (optional, in "Area for future growth", indicated by IP
  Area Offset)
* Client's IPv4 Prefix
* IPv6 Prefix Count
* Client's IPv6 Prefix(es) (optional, depending on IPv6 Prefix Count)
* Trailer: "SMCR" or "SMCD" in EBCDIC/IBM-1047 encoding

SMC-R IPv4 example with hex dump
([html with highlighting](examples/lo-proposal-smcr-ipv4.html)):

```
127.0.0.1:60294 -> 127.0.0.1:50000: Proposal: Eyecatcher: SMC-R,
Type: 1 (Proposal), Length: 52, Version: 1, Flag: 0, Path: SMC-R,
Peer ID: 45472@98:03:9b:ab:cd:ef, SMC-R GID: fe80::9a03:9bff:feab:cdef,
RoCE MAC: 98:03:9b:ab:cd:ef, IP Area Offset: 0, SMC-D GID: 0,
IPv4 Prefix: 127.0.0.0/8, IPv6 Prefix Count: 0, Trailer: SMC-R
00000000  e2 d4 c3 d9 01 00 34 10  b1 a0 98 03 9b ab cd ef  |......4.........|
00000010  fe 80 00 00 00 00 00 00  9a 03 9b ff fe ab cd ef  |................|
00000020  98 03 9b ab cd ef 00 00  7f 00 00 00 08 00 00 00  |................|
00000030  e2 d4 c3 d9                                       |....|
```

SMC-R IPv6 example with hex dump
([html with highlighting](examples/lo-proposal-smcr-ipv6.html)):

```
::1:33186 -> ::1:50000: Proposal: Eyecatcher: SMC-R, Type: 1 (Proposal),
Length: 69, Version: 1, Flag: 0, Path: SMC-R, Peer ID: 14660@98:03:9b:ab:cd:ef,
SMC-R GID: fe80::9a03:9bff:feab:cdef, RoCE MAC: 98:03:9b:ab:cd:ef,
IP Area Offset: 0, SMC-D GID: 0, IPv4 Prefix: 0.0.0.0/0, IPv6 Prefix Count: 1,
IPv6 Prefix: ::1/128, Trailer: SMC-R
00000000  e2 d4 c3 d9 01 00 45 10  39 44 98 03 9b ab cd ef  |......E.9D......|
00000010  fe 80 00 00 00 00 00 00  9a 03 9b ff fe ab cd ef  |................|
00000020  98 03 9b ab cd ef 00 00  00 00 00 00 00 00 00 01  |................|
00000030  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 01  |................|
00000040  80 e2 d4 c3 d9                                    |.....|
```

### Accept

With a CLC Accept message, the server accepts the client's proposal and, if
multiple SMC variants were given by the client, selects either SMC-R or SMC-D.
Also, the server provides additional information for the SMC connection setup.
The format of the Accept message depends on the selected SMC variant.

SMC-R CLC Accept Message:
* Eyecatcher: "SMCR" in EBCDIC/IBM-1047 encoding
* Type: 2 (Proposal)
* Length: 68
* Version: 1
* First Contact Flag
* Path: SMC-R
* Server's Peer ID
* Server's SMC-R GID
* Server's RoCE MAC address
* Server's QP number
* Server's RMB RKey
* Server's RMB element index
* Server's RMB element alert token
* Server's RMB element buffer size
* Server's QP MTU size
* Server's RMB virtual address
* Server's initial packet sequence number
* Trailer: "SMCR" in EBCDIC/IBM-1047 encoding

SMC-D CLC Accept Message:
* Eyecatcher: "SMCR" or "SMCD" in EBCDIC/IBM-1047 encoding
* Type: 2 (Proposal)
* Length: 48
* Version: 1
* First Contact Flag
* Path: SMC-D
* Server's SMC-D GID
* Server's SMC-D Token
* Server's DMB element index
* Server's DMB element buffer size
* Link identifier
* Trailer: "SMCR" or "SMCD" in EBCDIC/IBM-1047 encoding


SMC-R IPv4 example with hex dump
([html with highlighting](examples/lo-accept-smcr-ipv4.html)):

```
127.0.0.1:50000 -> 127.0.0.1:60294: Accept: Eyecatcher: SMC-R,
Type: 2 (Accept), Length: 68, Version: 1, First Contact: 1, Path: SMC-R,
Peer ID: 45472@98:03:9b:ab:cd:ef, SMC-R GID: fe80::9a03:9bff:feab:cdef,
RoCE MAC: 98:03:9b:ab:cd:ef, QP Number: 228, RMB RKey: 5501, RMBE Index: 1,
RMBE Alert Token: 5, RMBE Size: 2 (65536), QP MTU: 3 (1024),
RMB Virtual Address: 0xf0a60000, Packet Sequence Number: 7534078,
Trailer: SMC-R
00000000  e2 d4 c3 d9 02 00 44 18  b1 a0 98 03 9b ab cd ef  |......D.........|
00000010  fe 80 00 00 00 00 00 00  9a 03 9b ff fe ab cd ef  |................|
00000020  98 03 9b ab cd ef 00 00  e4 00 00 15 7d 01 00 00  |............}...|
00000030  00 05 23 00 00 00 00 00  f0 a6 00 00 00 72 f5 fe  |..#..........r..|
00000040  e2 d4 c3 d9                                       |....|
```

### Confirm

SMC-R IPv4 example:

```
127.0.0.1:60294 -> 127.0.0.1:50000: Confirm: Eyecatcher: SMC-R,
Type: 3 (Confirm), Length: 68, Version: 1, Flag: 0, Path: SMC-R,
Peer ID: 45472@98:03:9b:ab:cd:ef, SMC-R GID: fe80::9a03:9bff:feab:cdef,
RoCE MAC: 98:03:9b:ab:cd:ef, QP Number: 229, RMB RKey: 6271, RMBE Index: 1,
RMBE Alert Token: 6, RMBE Size: 2 (65536), QP MTU: 3 (1024),
RMB Virtual Address: 0xf0a40000, Packet Sequence Number: 887204,
Trailer: SMC-R
00000000  e2 d4 c3 d9 03 00 44 10  b1 a0 98 03 9b ab cd ef  |......D.........|
00000010  fe 80 00 00 00 00 00 00  9a 03 9b ff fe ab cd ef  |................|
00000020  98 03 9b ab cd ef 00 00  e5 00 00 18 7f 01 00 00  |................|
00000030  00 06 23 00 00 00 00 00  f0 a4 00 00 00 0d 89 a4  |..#.............|
00000040  e2 d4 c3 d9                                       |....|
```

### Decline

SMC-R IPv4 example:

```
127.0.0.1:36140 -> 127.0.0.1:50000: Decline: Eyecatcher: SMC-R,
Type: 4 (Decline), Length: 28, Version: 1, Out of Sync: 0, Path: SMC-R,
Peer ID: 9509@25:25:25:25:25:00, Peer Diagnosis: no SMC device found (R or D),
Trailer: SMC-R
00000000  e2 d4 c3 d9 04 00 1c 10  25 25 25 25 25 25 25 00  |........%%%%%%%.|
00000010  03 03 00 00 00 00 00 00  e2 d4 c3 d9              |............|
```

### Example Packet Dumps with Highlighting

The folder [examples](examples/) contains example packet dumps of CLC handshake
messages with highlighting of the fields and bytes in the messages.  The
highlighting is realized as a mouseover effect.

Note: Open the dumps in a web browser to see the highlighting mouseover effect.

SMC-R CLC handshake messages over IPv4 on the loopback device:
* [CLC Proposal](examples/lo-proposal-smcr-ipv4.html)
* [CLC Accept](examples/lo-accept-smcr-ipv4.html)
* [CLC Confirm](examples/lo-confirm-smcr-ipv4.html)
* [CLC Decline](examples/lo-decline-smcr-ipv4.html)

SMC-R CLC handshake messages over IPv6 on the loopback device:
* [CLC Proposal](examples/lo-proposal-smcr-ipv6.html)
* [CLC Accept](examples/lo-accept-smcr-ipv6.html)
* [CLC Confirm](examples/lo-confirm-smcr-ipv6.html)
* [CLC Decline](examples/lo-decline-smcr-ipv6.html)
