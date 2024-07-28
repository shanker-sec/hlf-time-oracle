package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/beevik/ntp"
	"github.com/beevik/nts"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

type ntpOptsStruct struct {
	// List of NTP servers. Possible file record formats: "host|port" or "host".
	// Where "host" - IPv4 address or IPv6 address or domain name. Examples:
	// "[2001:6d0:ffd4::1]"
	// "[2001:6d0:ffd4::1]|123"
	// "82.142.168.18"
	// "82.142.168.18|123"
	// "0.pool.ntp.org"
	// "0.pool.ntp.org|123".
	file string

	// Timeout determines how long the client waits for a response from the
	// server before failing with a timeout error.
	timeout time.Duration

	// TTL specifies the maximum number of IP hops before the query datagram
	// is dropped by the network. See also: https://en.wikipedia.org/wiki/Time_to_live#IP_packets .
	TTL int

	// Version of the NTP protocol to use. Defaults to 4.
	Version int

	// LocalAddress contains the local IP address to use when creating a
	// connection to the remote NTP server. This may be useful when the local
	// system has more than one IP address. This address should not contain
	// a port number.
	LocalAddress string

	// Address of the remote NTP server.
	server string

	// Port indicates the port used to reach the remote NTP server.
	port int
}

type ntsOptsStruct struct {
	// List of NTS servers. Possible file record formats: "host|port" or "host".
	// Where "host" is domain name. Examples:
	// "time.cloudflare.com"
	// "time.cloudflare.com|4460".
	file string

	// Address of the remote NTS server.
	server string

	// Port indicates the port used to reach the remote NTS server.
	ntsPort int

	// SessionOptions contains options for customizing the behavior of an NTS
	// session.
	opt *nts.SessionOptions
}

type TimeOracleChaincode struct {
}

func (cc *TimeOracleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (cc *TimeOracleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, _ := stub.GetFunctionAndParameters()

	switch function {
	case "GetTimeNtp":
		return cc.GetTimeNtp()
	case "GetTimeNts":
		return cc.GetTimeNts()
	default:
		return shim.Error("Invalid function name.")
	}
}

// split takes string in format "server|port" and returns server, port and error.
func split(str *string) (string, int, error) {
	var (
		server = ""
		port   = 0
		err    error
	)

	fields := strings.Split(*str, "|")

	for i, data := range fields {
		switch i {
		case 0:
			server = data
		case 1:
			port, err = strconv.Atoi(data)

			if err != nil {
				return server, port, fmt.Errorf("bad port number: %v", err)
			}
		}
	}

	return server, port, nil
}

// checkFileSize checks if the file exists and its size. It takes file name. Returns error if:
// file doesn't exist;
// size of file == 0;
// size of file > maxFileSize.
func checkFileSize(name *string) error {
	fileInfo, err := os.Stat(*name)
	if err != nil {
		return fmt.Errorf("in checkFileSize(): %s ", err)
	}

	var maxFileSize int64 = 102400
	if (fileInfo.Size() == 0) || (fileInfo.Size() > maxFileSize) {
		return fmt.Errorf("in checkFileSize(), bad size of file %s: (must be > 0 or < %d bytes)", *name, maxFileSize)
	}

	return nil
}

// ntpQueryLoop sends a request sequentially to each NTP-server in the list until it receives the first response. It gets:
// a pointer to a file with the list of servers;
// a pointer to the data structure to form a query.
// Returns time (in format time.Time) and bool ('true' in case of a response, otherwise `false`).
func ntpQueryLoop(readFile *os.File, ntpOpts *ntpOptsStruct) (time.Time, bool) {
	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		var err error

		strPointer := fileScanner.Text()
		ntpOpts.server, ntpOpts.port, err = split(&strPointer)

		if err != nil {
			log.Printf("in the GetTimeNtp(): parsing error in file : %v %s ", ntpOpts.file, err)

			continue
		}

		options := ntp.QueryOptions{
			Timeout:      ntpOpts.timeout * time.Second,
			TTL:          ntpOpts.TTL,
			Port:         ntpOpts.port,
			Version:      ntpOpts.Version,
			LocalAddress: ntpOpts.LocalAddress,
		}

		response, err1 := ntp.QueryWithOptions(ntpOpts.server, options)
		if err1 != nil || (response == nil) { // not sure if `err1 != nil` always in case `response == nil`
			log.Printf("error in the GetTimeNtp(): %s ", err1)

			continue
		}

		if err2 := response.Validate(); err2 != nil {
			log.Printf("error in the GetTimeNtp(): %s ", err2)

			continue
		}

		return time.Now().Add(response.ClockOffset).UTC(), true
	}

	var nilTime time.Time

	return nilTime, false
}

// ntsQueryLoop sends a request sequentially to each NTS-server in the list until it receives the first response. It gets:
// a pointer to a file with the list of servers;
// a pointer to the data structure to form a query.
// Returns time (in format time.Time) and bool ('true' in case of a response, otherwise `false`).
func ntsQueryLoop(readFile *os.File, ntsOpts *ntsOptsStruct, ntpOpts *ntpOptsStruct) (time.Time, bool) {
	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	var err error

	var host string

	for fileScanner.Scan() {
		strPointer := fileScanner.Text()
		ntsOpts.server, ntsOpts.ntsPort, err = split(&strPointer)

		if err != nil {
			log.Printf("in the GetTimeNts(): parsing error in file : %v %s ", ntsOpts.file, err)

			continue
		}

		switch ntsOpts.ntsPort {
		case 0:
			host = ntsOpts.server

		default:
			host = strings.Join([]string{ntsOpts.server, strconv.Itoa(ntsOpts.ntsPort)}, ":")
		}

		session, err := nts.NewSessionWithOptions(host, ntsOpts.opt)

		if err != nil {
			log.Printf("error in the GetTimeNts(): %s", err)

			continue
		}

		options := ntp.QueryOptions{ // no need to specify a NTP port (using NTP Port from NTS response, see: https://www.rfc-editor.org/rfc/rfc8915#section-4.1.8).
			Timeout:      ntpOpts.timeout * time.Second,
			TTL:          ntpOpts.TTL,
			Version:      ntpOpts.Version,
			LocalAddress: ntpOpts.LocalAddress,
		}

		response, err2 := session.QueryWithOptions(&options)
		if (err2 != nil) || (response == nil) { // not sure if `err2 != nil` always in case `response == nil`
			log.Printf("error in the GetTimeNts(): %s", err2)

			continue
		}

		if err3 := response.Validate(); err3 != nil {
			log.Printf("error in the GetTimeNts(): %s", err3)

			continue
		}

		return time.Now().Add(response.ClockOffset).UTC(), true
	}

	var nilTime time.Time

	return nilTime, false
}

// GetTimeNts returns the timestamp from one of NTS server in format: yyyy-mm-dd hh:mm:ss.nnnnnnnnn +0000 UTC.
// For example: "2024-07-09 15:37:13.879908993 +0000 UTC"
// In case of failure to connect to any of the servers:
// the following is logged: "Reach end of file";
// returns an error with the text "Failed to get response from NTS servers, see log file".
// The log also stores information about the reasons for the unsuccessful receipt of data from the NTS server
// If a man-in-the-middle attack is attempted, the incident will be logged.
func (cc *TimeOracleChaincode) GetTimeNts() pb.Response {
	var ntsOpts = ntsOptsStruct{
		file:    "nts.txt",
		server:  "",
		ntsPort: 4460,
		opt:     &nts.SessionOptions{},
	}

	var ntpOpts = ntpOptsStruct{
		timeout:      5,
		TTL:          128,
		Version:      4,
		LocalAddress: "",
	}

	if err := checkFileSize(&ntsOpts.file); err != nil {
		log.Println(err)

		return shim.Error("error in servers list, see log")
	}

	readFile, _ := os.Open(ntsOpts.file) // error has already been checked in the checkFileSize().
	defer readFile.Close()

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	if accurateTime, result := ntsQueryLoop(readFile, &ntsOpts, &ntpOpts); result {
		return shim.Success([]byte(fmt.Sprint(accurateTime)))
	}

	log.Printf("Reach end of file : %s", ntsOpts.file)

	return shim.Error("Failed to get response from NTS servers, see log file")
}

// GetTimeNtp returns the timestamp from one of NTP server in format: yyyy-mm-dd hh:mm:ss.nnnnnnnnn +0000 UTC.
// For example: "2024-07-09 15:37:13.879908993 +0000 UTC"
// In case of failure to connect to any of the servers:
// the following is logged: "Reach end of file";
// returns an error with the text "Failed to get response from NTP servers, see log file".
// The log also stores information about the reasons for the unsuccessful receipt of data from the NTP server.
func (cc *TimeOracleChaincode) GetTimeNtp() pb.Response {
	var ntpOpts = ntpOptsStruct{
		file:         "ntp.txt",
		timeout:      5,
		TTL:          128,
		Version:      4,
		LocalAddress: "",
		server:       "",
		port:         123,
	}

	if err := checkFileSize(&ntpOpts.file); err != nil {
		log.Println("File error :", err)

		return shim.Error("error in servers list, see log file")
	}

	readFile, _ := os.Open(ntpOpts.file) // error has already been checked in the checkFileSize().
	defer readFile.Close()

	if accurateTime, result := ntpQueryLoop(readFile, &ntpOpts); result {
		return shim.Success([]byte(fmt.Sprint(accurateTime)))
	}

	log.Printf("Reach end of file : %s", ntpOpts.file)

	return shim.Error("Failed to get response from NTP servers, see log file")
}

func main() {
	err := shim.Start(new(TimeOracleChaincode))
	if err != nil {
		fmt.Printf("Error starting TimeOracleChaincode: %s", err)
	}
}
