package main

import (
	"log"
	"os"
	"testing"
	"time"

	pb "github.com/hyperledger/fabric-protos-go/peer"
)

func Test_split(t *testing.T) {
	hostOnly := "host.com"
	hostAndPort := "host.com|123"
	BadPort := "host.com|123k"

	type args struct {
		str *string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		port    int
		wantErr bool
	}{
		{
			name: "Test Host Only String",
			args: args{
				str: &hostOnly,
			},
			want:    "host.com",
			wantErr: false,
		},
		{
			name: "Test Host And Port String",
			args: args{
				str: &hostAndPort,
			},
			want:    "host.com",
			port:    123,
			wantErr: false,
		},
		{
			name: "Negative Port",
			args: args{
				str: &BadPort,
			},
			want:    "host.com",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := split(tt.args.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("split() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("split() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.port {
				t.Errorf("split() got1 = %v, want %v", got1, tt.port)
			}
		})
	}
}

func Test_checkFileSize(t *testing.T) {
	notExist := "test.txt"
	zeroFileSize := "testFile.txt"

	f, err := os.Create(zeroFileSize)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	type args struct {
		name *string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Negative file doesn't exist",
			args: args{
				name: &notExist,
			},
			wantErr: true,
		},
		{
			name: "Negative zero file size",
			args: args{
				name: &zeroFileSize,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkFileSize(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("checkFileSize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	os.Remove(zeroFileSize)
}

func Test_ntpQueryLoop_1(t *testing.T) {

	ntpFile := "testFile.txt"
	var opts = ntpOptsStruct{
		file:    ntpFile,
		timeout: 5,
		TTL:     128,
		Version: 4,
	}

	type args struct {
		ntpOpts *ntpOptsStruct
	}
	tests := []struct {
		name      string
		args      args
		isSuccess bool
	}{
		{
			name: "Positive NTP via IPv4 without port number",
			args: args{
				ntpOpts: &opts,
			},
			isSuccess: true,
		},

		{
			name: "Negative NTP via IPv6 with bad port number", // Run only if you have IPv6 address
			args: args{
				ntpOpts: &opts,
			},
			isSuccess: false,
		},
		{
			name: "Positive NTP via IPv6 with port number", // Run only if you have IPv6 address
			args: args{
				ntpOpts: &opts,
			},
			isSuccess: true,
		},

		{
			name: "Positive NTP via IPv6 without port number", // Run only if you have IPv6 address
			args: args{
				ntpOpts: &opts,
			},
			isSuccess: true,
		},
		{
			name: "Positive last DNS name is good",
			args: args{
				ntpOpts: &opts,
			},
			isSuccess: true,
		},
		{
			name: "Negative bad NTP Version",
			args: args{
				ntpOpts: &opts,
			},
			isSuccess: false,
		},
	}

	for _, tt := range tests {
		switch tt.name {
		case "Positive NTP via IPv4 without port number":
			if err := os.WriteFile(ntpFile, []byte("82.142.168.18"), 0666); err != nil { // 0.pool.ntp.org
				log.Fatal(err)
			}

		case "Negative NTP via IPv6 with bad port number":
			if err := os.WriteFile(ntpFile, []byte("[2001:6d0:ffd4::1]|122"), 0666); err != nil { // ntp.ix.ru
				log.Fatal(err)
			}

		case "Positive NTP via IPv6 with port number":
			if err := os.WriteFile(ntpFile, []byte("[2001:6d0:ffd4::1]|123"), 0666); err != nil { // ntp.ix.ru
				log.Fatal(err)
			}

		case "Positive NTP via IPv6 without port number":
			if err := os.WriteFile(ntpFile, []byte("[2001:6d0:ffd4::1]"), 0666); err != nil { // ntp.ix.ru
				log.Fatal(err)
			}

		case "Positive last DNS name is good":
			if err := os.WriteFile(ntpFile, []byte("aaaaaaa.ru\nbbbbbbbbbbbb.ru\n0.beevik-ntp.pool.ntp.org"), 0666); err != nil {
				log.Fatal(err)
			}
		case "Negative bad NTP Version":
			if err := os.WriteFile(ntpFile, []byte("0.beevik-ntp.pool.ntp.org"), 0666); err != nil {
				log.Fatal(err)
			}

			var badVerOpt = ntpOptsStruct{
				file:    ntpFile,
				timeout: 5,
				TTL:     128,
				Version: 1,
			}

			opts = badVerOpt
		}

		readFile, _ := os.Open(ntpFile)
		defer readFile.Close()

		t.Run(tt.name, func(t *testing.T) {
			time.Sleep(2 * time.Second) // pause between connections to protect against blocking
			gotTime, got1 := ntpQueryLoop(readFile, tt.args.ntpOpts)
			if got1 != tt.isSuccess {
				t.Errorf("ntpQueryLoop() got1 = %v, want %v", got1, tt.isSuccess)
			}

			if tt.isSuccess == true {
				log.Printf("Time from NTP: %s \n", gotTime)
			}

		})
	}

	if err := os.Remove(ntpFile); err != nil {
		log.Fatal(err)
	}
}

func Test_ntpQueryLoop_2(t *testing.T) {

	ntpFile := "testFile.txt"
	var opts = ntpOptsStruct{
		file:    ntpFile,
		timeout: 5,
		TTL:     128,
		Version: 4,
	}

	type args struct {
		ntpOpts *ntpOptsStruct
	}
	tests := []struct {
		name      string
		args      args
		isSuccess bool
	}{
		{
			name: "Negative bad DNS name",
			args: args{
				ntpOpts: &opts,
			},
			isSuccess: false,
		},
		{
			name: "Negative domain name with bad port number (> 65535)",
			args: args{
				ntpOpts: &opts,
			},
			isSuccess: false,
		},

		{
			name: "Negative NTP via IPv6 with invalid port number",
			args: args{
				ntpOpts: &opts,
			},
			isSuccess: false,
		},
	}

	for _, tt := range tests {
		switch tt.name {
		case "Negative NTP via IPv6 with invalid port number":
			if err := os.WriteFile(ntpFile, []byte("[2001:6d0:ffd4::1]|122k"), 0666); err != nil { // ntp.ix.ru
				log.Fatal(err)
			}
		case "Negative bad DNS name":
			if err := os.WriteFile(ntpFile, []byte("ferfefefefefe.ru"), 0666); err != nil {
				log.Fatal(err)
			}
		case "Negative domain name with bad port number (> 65535)":
			if err := os.WriteFile(ntpFile, []byte("0.beevik-ntp.pool.ntp.org|65536"), 0666); err != nil {
				log.Fatal(err)
			}

			var badVerOpt = ntpOptsStruct{
				file:    ntpFile,
				timeout: 5,
				TTL:     128,
				Version: 1,
			}

			opts = badVerOpt
		}

		readFile, _ := os.Open(ntpFile)
		defer readFile.Close()

		t.Run(tt.name, func(t *testing.T) {
			gotTime, got1 := ntpQueryLoop(readFile, tt.args.ntpOpts)
			if got1 != tt.isSuccess {
				t.Errorf("ntpQueryLoop() got1 = %v, want %v", got1, tt.isSuccess)
			}

			if tt.isSuccess == true {
				log.Printf("Time from NTP: %s \n", gotTime)
			}

		})
	}

	if err := os.Remove(ntpFile); err != nil {
		log.Fatal(err)
	}
}

func Test_ntpQueryLoop_3(t *testing.T) {

	ntpFile := "testFile.txt"
	var opts = ntpOptsStruct{
		file:    ntpFile,
		timeout: 5,
		TTL:     128,
		Version: 4,
	}

	type args struct {
		ntpOpts *ntpOptsStruct
	}
	tests := []struct {
		name      string
		args      args
		isSuccess bool
	}{
		{
			name: "Positive last DNS name is good",
			args: args{
				ntpOpts: &opts,
			},
			isSuccess: true,
		},
	}

	for _, tt := range tests {
		switch tt.name {
		case "Positive last DNS name is good":
			if err := os.WriteFile(ntpFile, []byte("aaaaaaa.ru\nbbbbbbbbbbbb.ru\n0.beevik-ntp.pool.ntp.org"), 0666); err != nil {
				log.Fatal(err)
			}
		}

		readFile, _ := os.Open(ntpFile)
		defer readFile.Close()

		t.Run(tt.name, func(t *testing.T) {
			gotTime, got1 := ntpQueryLoop(readFile, tt.args.ntpOpts)
			if got1 != tt.isSuccess {
				t.Errorf("ntpQueryLoop() got1 = %v, want %v", got1, tt.isSuccess)
			}

			if tt.isSuccess == true {
				log.Printf("Time from NTP: %s \n", gotTime)
			}

		})
	}

	if err := os.Remove(ntpFile); err != nil {
		log.Fatal(err)
	}
}

func TestTimeOracleChaincode_GetTimeNtp(t *testing.T) {
	var ntpOpts = ntpOptsStruct{
		file:    "ntp.txt",
		timeout: 5,
		TTL:     128,
		Version: 4,
	}

	tests := []struct {
		name   string
		cc     *TimeOracleChaincode
		want   pb.Response
		status int32
	}{
		{
			name:   "Negative IPv4 and bad port number (> 65535)",
			status: 500,
		},
		{
			name:   "Negative domain name and bad port number (> 65535)",
			status: 500,
		},

		{
			name:   "Negative IPv4 and bad port number",
			status: 500,
		},
		{
			name:   "Negative IPv4 and invalid port number",
			status: 500,
		},

		{
			name:   "Negative zero file size",
			status: 500,
		},
	}
	for _, tt := range tests {
		switch tt.name {
		case "Negative zero file size":
			if err := os.WriteFile(ntpOpts.file, []byte(""), 0666); err != nil { // 0.pool.ntp.org
				log.Fatal(err)
			}
		case "Negative IPv4 and bad port number (> 65535)":
			if err := os.WriteFile(ntpOpts.file, []byte("82.142.168.18|65536"), 0666); err != nil { // 0.pool.ntp.org
				log.Fatal(err)
			}

		case "Negative domain name and bad port number (> 65535)":
			if err := os.WriteFile(ntpOpts.file, []byte("0.pool.ntp.org|65536"), 0666); err != nil { // 0.pool.ntp.org
				log.Fatal(err)
			}

		case "Negative IPv4 and bad port number":
			if err := os.WriteFile(ntpOpts.file, []byte("82.142.168.18|122"), 0666); err != nil { // 0.pool.ntp.org
				log.Fatal(err)
			}

		case "Negative IPv4 and invalid port number":
			if err := os.WriteFile(ntpOpts.file, []byte("82.142.168.18|122k"), 0666); err != nil { // 0.pool.ntp.org
				log.Fatal(err)
			}
		}
		readFile, _ := os.Open(ntpOpts.file)
		defer readFile.Close()

		t.Run(tt.name, func(t *testing.T) {
			cc := &TimeOracleChaincode{}
			time.Sleep(3 * time.Second) // pause between connections to protect against blocking
			got := cc.GetTimeNtp()
			if got.Status != tt.status {
				t.Errorf("TimeOracleChaincode.GetTimeNtp() = %v, want %v", got.Status, tt.status)
			}
		})
	}

	if err := os.Remove(ntpOpts.file); err != nil {
		log.Fatal(err)
	}
}

// Run only if you have IPv6 address
func TestTimeOracleChaincode_GetTimeNtp_IPv6(t *testing.T) {
	var ntpOpts = ntpOptsStruct{
		file:    "ntp.txt",
		timeout: 5,
		TTL:     128,
		Version: 4,
	}

	tests := []struct {
		name   string
		cc     *TimeOracleChaincode
		want   pb.Response
		status int32
	}{
		{
			name:   "Negative IPv6 and bad port number (> 65535)",
			status: 500,
		},

		{
			name:   "Negative NTP via IPv6 with bad port number",
			status: 500,
		},

		{
			name:   "Negative NTP via IPv6 with invalid port number",
			status: 500,
		},

		{
			name:   "Negative bad IPv6 address",
			status: 500,
		},
		{
			name:   "Positive NTP via IPv6 without port number",
			status: 200,
		},
		{
			name:   "Positive NTP via IPv6 with port number",
			status: 200,
		},
	}
	for _, tt := range tests {
		switch tt.name {
		case "Negative IPv6 and bad port number (> 65535)":
			if err := os.WriteFile(ntpOpts.file, []byte("[2001:6d0:ffd4::1]|65536"), 0666); err != nil { // ntp.ix.ru
				log.Fatal(err)
			}

		case "Negative NTP via IPv6 with bad port number":
			if err := os.WriteFile(ntpOpts.file, []byte("[2001:6d0:ffd4::1]|122"), 0666); err != nil { // ntp.ix.ru
				log.Fatal(err)
			}

		case "Negative NTP via IPv6 with invalid port number":
			if err := os.WriteFile(ntpOpts.file, []byte("[2001:6d0:ffd4::1]|122k"), 0666); err != nil { // ntp.ix.ru
				log.Fatal(err)
			}

		case "Negative bad IPv6 address":
			if err := os.WriteFile(ntpOpts.file, []byte("[20z1:6d0:ffd4::18]"), 0666); err != nil {
				log.Fatal(err)
			}

		case "Positive NTP via IPv6 without port number":
			if err := os.WriteFile(ntpOpts.file, []byte("[2001:6d0:ffd4::1]"), 0666); err != nil { // ntp.ix.ru
				log.Fatal(err)
			}

		case "Positive NTP via IPv6 with port number":
			if err := os.WriteFile(ntpOpts.file, []byte("[2001:6d0:ffd4::1]|123"), 0666); err != nil { // ntp.ix.ru
				log.Fatal(err)
			}
		}
		readFile, _ := os.Open(ntpOpts.file)
		defer readFile.Close()

		t.Run(tt.name, func(t *testing.T) {
			cc := &TimeOracleChaincode{}
			time.Sleep(3 * time.Second) // pause between connections to protect against blocking
			got := cc.GetTimeNtp()
			if got.Status != tt.status {
				t.Errorf("TimeOracleChaincode.GetTimeNtp() = %v, want %v", got.Status, tt.status)
			}
		})
	}

	if err := os.Remove(ntpOpts.file); err != nil {
		log.Fatal(err)
	}
}

func TestTimeOracleChaincode_GetTimeNtp_2(t *testing.T) {
	var ntpOpts = ntpOptsStruct{
		file:    "ntp.txt",
		timeout: 5,
		TTL:     128,
		Version: 4,
	}

	tests := []struct {
		name string
		cc   *TimeOracleChaincode
		//want   pb.Response
		status int32
	}{
		{
			name:   "Positive NTP via IPv4 without port number",
			status: 200,
		},
		{
			name:   "Positive NTP via IPv4 with port number",
			status: 200,
		},
		{
			name:   "Positive NTP via domain name with port number",
			status: 200,
		},

		{
			name:   "Positive NTP via domain name without port number",
			status: 200,
		},
	}
	for _, tt := range tests {
		switch tt.name {
		case "Positive NTP via IPv4 without port number":
			if err := os.WriteFile(ntpOpts.file, []byte("82.142.168.18"), 0666); err != nil { // 0.pool.ntp.org
				log.Fatal(err)
			}

		case "Positive NTP via IPv4 with port number":
			if err := os.WriteFile(ntpOpts.file, []byte("194.190.168.1|123"), 0666); err != nil { // 0.pool.ntp.org
				log.Fatal(err)
			}

		case "Positive NTP via domain name with port number":
			if err := os.WriteFile(ntpOpts.file, []byte("0.pool.ntp.org|123"), 0666); err != nil {
				log.Fatal(err)
			}

		case "Positive NTP via domain name without port number":
			if err := os.WriteFile(ntpOpts.file, []byte("ntp.ix.ru"), 0666); err != nil {
				log.Fatal(err)
			}
		}

		readFile, _ := os.Open(ntpOpts.file)
		defer readFile.Close()

		t.Run(tt.name, func(t *testing.T) {
			cc := &TimeOracleChaincode{}
			time.Sleep(2 * time.Second) // pause between connections to protect against blocking
			got := cc.GetTimeNtp()
			if got.Status != tt.status {
				t.Errorf("TimeOracleChaincode.GetTimeNtp() = %v, want %v", got.Status, tt.status)
			}

			if got.Status == 200 {
				log.Printf("TimeOracleChaincode.GetTimeNtp() = %s \n", got.Payload)
			}
		})
	}
	if err := os.Remove(ntpOpts.file); err != nil {
		log.Fatal(err)
	}
}

func TestTimeOracleChaincode_GetTimeNtp_3(t *testing.T) {
	var ntpOpts = ntpOptsStruct{
		file:    "ntp.txt",
		timeout: 5,
		TTL:     128,
		Version: 4,
	}

	tests := []struct {
		name string
		cc   *TimeOracleChaincode
		//want   pb.Response
		status int32
	}{
		{
			name:   "Positive last DNS name is good",
			status: 200,
		},
	}
	for _, tt := range tests {
		switch tt.name {

		case "Positive last DNS name is good":
			if err := os.WriteFile(ntpOpts.file, []byte("aaaaaaa.com\n1.pool.ntp.org"), 0666); err != nil {
				log.Fatal(err)
			}
		}

		readFile, _ := os.Open(ntpOpts.file)
		defer readFile.Close()

		t.Run(tt.name, func(t *testing.T) {
			cc := &TimeOracleChaincode{}
			time.Sleep(2 * time.Second) // pause between connections to protect against blocking
			got := cc.GetTimeNtp()
			if got.Status != tt.status {
				t.Errorf("TimeOracleChaincode.GetTimeNtp() = %v, want %v", got.Status, tt.status)
			}

			if got.Status == 200 {
				log.Printf("TimeOracleChaincode.GetTimeNtp() = %s \n", got.Payload)
			}
		})
	}

	if err := os.Remove(ntpOpts.file); err != nil {
		log.Fatal(err)
	}
}

func TestTimeOracleChaincode_GetTimeNts_1(t *testing.T) {
	var ntsOpts = ntsOptsStruct{
		file:    "nts.txt",
		ntsPort: 4460,
	}

	tests := []struct {
		name   string
		cc     *TimeOracleChaincode
		status int32
	}{
		{
			name:   "Positive NTS via domain name without port number",
			status: 200,
		},
		{
			name:   "Positive NTS via domain name with port number",
			status: 200,
		},
		{
			name:   "Positive last DNS name is good",
			status: 200,
		},
		{
			name:   "Negative NTS via IPv6",
			status: 500,
		},
		{
			name:   "Negative NTS via IPv4",
			status: 500,
		},
		{
			name:   "Negative NTS domain name with bad port number",
			status: 500,
		},
	}
	for _, tt := range tests {
		switch tt.name {
		case "Negative NTS via IPv4":
			if err := os.WriteFile(ntsOpts.file, []byte("162.159.200.1"), 0666); err != nil { // time.cloudflare.com
				log.Fatal(err)
			}

		case "Negative NTS via IPv6":
			if err := os.WriteFile(ntsOpts.file, []byte("[2606:4700:f1::1]"), 0666); err != nil { // time.cloudflare.com
				log.Fatal(err)
			}

		case "Positive NTS via domain name with port number":
			if err := os.WriteFile(ntsOpts.file, []byte("time.cloudflare.com|4460|"), 0666); err != nil {
				log.Fatal(err)
			}

		case "Negative NTS domain name with bad port number":
			if err := os.WriteFile(ntsOpts.file, []byte("time.cloudflare.com|4461|"), 0666); err != nil {
				log.Fatal(err)
			}

		case "Positive NTS via domain name without port number":
			if err := os.WriteFile(ntsOpts.file, []byte("time.cloudflare.com"), 0666); err != nil {
				log.Fatal(err)
			}

		case "Positive last DNS name is good":
			if err := os.WriteFile(ntsOpts.file, []byte("aaaaaaa.com\ntime.cloudflare.com"), 0666); err != nil {
				log.Fatal(err)
			}
		}

		readFile, _ := os.Open(ntsOpts.file)
		defer readFile.Close()

		t.Run(tt.name, func(t *testing.T) {
			cc := &TimeOracleChaincode{}
			time.Sleep(2 * time.Second) // pause between connections to protect against blocking
			got := cc.GetTimeNts()
			if got.Status != tt.status {
				t.Errorf("TimeOracleChaincode.GetTimeNts() = %v, want %v", got.Status, tt.status)
			}

			if got.Status == 200 {
				log.Printf("TimeOracleChaincode.GetTimeNts() =  %s \n", got.Payload)
			}
		})
	}
	if err := os.Remove(ntsOpts.file); err != nil {
		log.Fatal(err)
	}
}

// without connect
func TestTimeOracleChaincode_GetTimeNts_2(t *testing.T) {
	var ntsOpts = ntsOptsStruct{
		file:    "nts.txt",
		ntsPort: 4460,
	}

	tests := []struct {
		name   string
		cc     *TimeOracleChaincode
		status int32
	}{
		{
			name:   "Negative NTS domain name with invalid port number",
			status: 500,
		},

		{
			name:   "Negative NTS zero file size",
			status: 500,
		},

		{
			name:   "Negative NTS domain name with port number > 65535",
			status: 500,
		},
	}
	for _, tt := range tests {
		switch tt.name {

		case "Negative NTS zero file size":
			if err := os.WriteFile(ntsOpts.file, []byte(""), 0666); err != nil {
				log.Fatal(err)
			}

		case "Negative NTS domain name with invalid port number":
			if err := os.WriteFile(ntsOpts.file, []byte("time.cloudflare.com|4460P|"), 0666); err != nil {
				log.Fatal(err)
			}

		case "Negative NTS domain name with port number > 65535":
			if err := os.WriteFile(ntsOpts.file, []byte("time.cloudflare.com|65536"), 0666); err != nil {
				log.Fatal(err)
			}
		}

		readFile, _ := os.Open(ntsOpts.file)
		defer readFile.Close()

		t.Run(tt.name, func(t *testing.T) {
			cc := &TimeOracleChaincode{}
			got := cc.GetTimeNts()
			if got.Status != tt.status {
				t.Errorf("TimeOracleChaincode.GetTimeNts() = %v, want %v", got.Status, tt.status)
			}
		})
	}

	if err := os.Remove(ntsOpts.file); err != nil {
		log.Fatal(err)
	}
}
