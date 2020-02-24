package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/denisbrodbeck/machineid"
)

// Payload is what will be written to the output file (or eventually Stdout)
type Payload struct {
	OS                string
	machineID         string
	systemUUID        string
	serialNumber      string
	hardwareAddresses []string
}

const (
	windows = "windows"
	darwin  = "darwin"
	linux   = "linux"

	fileName = "output.txt"
)

func main() {
	_, err := os.Create(fileName)
	dealWithError(err)

	operatingSystem := runtime.GOOS
	machineID := getMachineID()
	macs := getMACAdresses()

	payload := &Payload{
		OS:                operatingSystem,
		machineID:         machineID,
		hardwareAddresses: macs,
	}

	switch operatingSystem {
	case windows:
		payload.serialNumber = windowsSerialNumber()
		payload.systemUUID = windowsUUID()
	case darwin:
		serialNumber, systemUUID := macInfo()
		payload.serialNumber = serialNumber
		payload.systemUUID = systemUUID
	case linux:
		systemUUID := getLinuxOSSystemUUID()
		serialNumber := getlinuxSerialNumber()
		payload.serialNumber = serialNumber
		payload.systemUUID = systemUUID
	default:
		panic("Could not detect Operating system")
	}
	writeOutput(payload)
}

func windowsUUID() string {
	out, err := exec.Command("wmic", "csproduct", "get", "UUID").Output()
	dealWithError(err)

	return string(out)
}

func windowsSerialNumber() string {
	out, err := exec.Command("wmic", "bios", "get", "serialnumber").Output()
	dealWithError(err)

	return string(out)
}

func macInfo() (string, string) {
	out, err := exec.Command("/usr/sbin/ioreg", "-l").Output()
	dealWithError(err)

	return findMacSerial(out), findMacHardwareUUID(out)
}

func findMacHardwareUUID(out []byte) string {
	re := regexp.MustCompile(`IOPlatformUUID\" = \"(.*)\"`)
	matchSlice := re.FindSubmatch(out)
	if len(matchSlice) < 2 {
		return ""
	}
	return string(matchSlice[1])
}

func findMacSerial(out []byte) string {
	for _, l := range strings.Split(string(out), "\n") {
		if strings.Contains(l, "IOPlatformSerialNumber") {
			s := strings.Split(l, " ")
			str := s[len(s)-1]
			return str
		}
	}
	panic("couldn't find IOS SerialNumber")
}

func getlinuxSerialNumber() string {
	out, err := exec.Command("/usr/sbin/dmidecode", "-s", "system-serial-number").Output()
	dealWithError(err)
	return string(out)
}

func getLinuxOSSystemUUID() string {
	out, err := exec.Command("/usr/sbin/dmidecode", "-s", "system-uuid").Output()
	dealWithError(err)
	return string(out)
}

func createOutputFile() (*os.File, error) {
	f, err := os.Create("output.txt")
	if err != nil {
		return nil, err
	}

	cwd, _ := os.Getwd()
	if err != nil {
		cwd = ""
	}

	fmt.Printf("%s/%s created\n", cwd, fileName)
	return f, nil
}

func writeOutput(p *Payload) {
	f, ferr := createOutputFile()
	dealWithError(ferr)

	numBytes, err := f.WriteString(fmt.Sprintf("SERIAL_NUMBER=%s\nSYSTEMUUID=%s\nMACHINEID=%s\nHARDWARE_ADDRESSES=%v", p.serialNumber, p.systemUUID, p.machineID, p.hardwareAddresses))
	if err != nil {
		fmt.Println(err)
		f.Close()
		return
	}

	fmt.Printf("wrote %d bytes", numBytes)

	err = f.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
}

func getMachineID() string {
	machineID, err := machineid.ID()
	if err != nil {
		machineID = ""
	}
	return machineID
}

func getMACAdresses() []string {
	ifas, err := net.Interfaces()
	dealWithError(err)

	var as []string
	for _, ifa := range ifas {
		a := ifa.HardwareAddr.String()
		if a != "" {
			as = append(as, a)
		}
	}
	return as
}

func dealWithError(err error) {
	if err != nil {
		panic(err)
	}
}
