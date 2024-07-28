
# hlf-time-oracle
`hlf-time-oracle` for Hyperledger Fabric version 2.5.x. Tested on version 2.5.9.

## Configuration
### NTP
Configurable `ntpOptsStruct` at `GetTimeNtp()` include:
* `File` : file with list of servers. Possible file record formats: "host|port" or "host". Where "host" - IPv4 address or IPv6 address or domain name. Examples: `[2001:6d0:ffd4::1]` or `[2001:6d0:ffd4::1]|123` or `82.142.168.18` or `82.142.168.18|123` or `0.pool.ntp.org` or `0.pool.ntp.org|123`
* `Timeout`: How long to wait before giving up on a response from the NTP server.
* `Version`: Which version of the NTP protocol to use (2, 3 or 4).
* `TTL`: The maximum number of IP hops before the request packet is discarded.
* `LocalAddress` : contains the local IP address to use when creating a  connection to the remote NTP server. This may be useful when the local system has more than one IP address. This address should not contain a port number.


### NTS

Configurable `ntpOptsStruct` at `GetTimeNts()` include:
* `File` : file with list of servers. Possible file record formats: "host|port" or "host". Where "host" is domain name. Examples: `time.cloudflare.com` or `time.cloudflare.com|4460`
* `ntsPort` : indicates the port used to reach the remote NTS server.
* `opt` : contains options for customizing the behavior of an NTS session.
Specify the file with servers in the `ntsOpts.file` variable. Possible file record formats: "host|port" or "host". Where "host" is domain name.

Configurable `ntpOptsStruct` include:
* `Timeout`: How long to wait before giving up on a response from the NTP server.
* `Version`: Which version of the NTP protocol to use (2, 3 or 4).
* `TTL`: The maximum number of IP hops before the request packet is discarded.
* `LocalAddress` : contains the local IP address to use when creating a  connection to the remote NTP server. This may be useful when the local system has more than one IP address. This address should not contain a port number.

### Creating a file with servers in container

Find container ID with chaincode:
```bash
 docker ps
```

Execute the command (in host system):
```bash
docker exec -u 0 -it container_ID /bin/bash
```
You will see: `root@xxxxxxxxxxxx`. Where `xxxxxxxxxxxx` - new container ID.

To create a file with servers in the container, you can use [this method](https://www.baeldung.com/linux/cat-writing-file).
Exit from container.

Execute the command (in host system):
```bash
docker commit new_container_ID REPOSITORY:TAG
```

Where `new_container_ID` - container ID from above. `REPOSITORY:TAG` - info from command `docker images`.
In my case, it was:
`docker commit 747018171b0b dev-peer0.org1.example.com-time_oracle_1-97e1b543395d763ed977a07a7c197579e3527e30b98325b1098e1dbb3484dea8-3c7a46c7225436ab088abcffd238a230591eda6c099e054c565820abcaefa67a:latest`

## Dependencies

If you use docker image [fabric-baseos](https://hub.docker.com/r/hyperledger/fabric-baseos) you need to modify this image by adding a root CA certificate:
```sh
apt update; apt install -y ca-certificates
```

Otherwise there will be an error when calling the function GetTimeNts():
`Failed to connect to NTS server key exchange failure: tls: failed to verify certificate: x509: certificate signed by unknown authority`

###  modification image fabric-baseos (for linux)

Execute the command:
```bash
docker run -u 0 -it hyperledger/fabric-baseos:latest /bin/bash
```

You will see:
`root@xxxxxxxxxxxx`. Where `xxxxxxxxxxxx` - container ID.

Execute in docker container:
```bash
apt update; apt install -y ca-certificates;exit
```

Execute the command (in host system):
```bash
docker commit container_ID hyperledger/fabric-baseos:latest
docker tag hyperledger/fabric-baseos:latest hyperledger/fabric-baseos:2.5
docker tag hyperledger/fabric-baseos:latest hyperledger/fabric-baseos:2.5.x
```
Where `container_ID` - container ID from above. And `2.5.x` - used version (for example: `2.5.5`)

## Unit tests

The detailed level of code coverage by unit tests is specified in the cover.html file (generated via `go tool cover`). 
Some unit tests make attempts to connect to servers. Including closed ports. Therefore, some places in the tests are paused to prevent blocking by network traffic controls. Pauses affect the overall duration of test execution. It took me 150 seconds to run the tests.
Some of the tests use IPv6. If there is no IPv6 on the system where the tests are run, the tests will fail.

## License
MIT
