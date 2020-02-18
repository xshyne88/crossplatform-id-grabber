package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/denisbrodbeck/machineid"
)

const fileName = "output.txt"

type Payload struct {
	OS string
	machineID string
	systemUUID string
	serialNumber string
}

const (
	windows = "windows"
	darwin = "darwin"
	linux = "linux"
)

func main() {
	_, err := os.Create(fileName)
	dealWithError(err)

	operatingSystem := runtime.GOOS
	machineID = getMachineID()

	payload := &Payload{
		OS: operatingSystem
		machineID: machineID
	}

	switch operatingSystem {
	case windows:
		payload.SerialNumber = windowsSerialNumber()
		payload.systemUUID = ""
	case darwin:
		serialNumber, systemUUID = macInfo()
	case linux:
		output = linuxSerialNumber()
		systemUUID = string(getLinuxOSSystemUUID())
	default:
		panic("couldn't detect os")
	}

	writeOutput(payload)
}

func windowsSerialNumber() string {
	out, err := exec.Command("wmic", "bios", "get", "serialnumber").Output()
	dealWithError(err)

	return string(out)
}

func macInfo() []byte {
	out, err := exec.Command("/usr/sbin/ioreg", "-l").Output()
	findMacHardwareUUID(out)
	dealWithError(err)

	return findMacSerial(out)
}

func findMacHardwareUUID(out []byte) string, error {
	re := regexp.MustCompile(`IOPlatformUUID\" = \"(.*)\"`)
	matchSlice := re.FindSubmatch(out)
	return string(matchSlice[1])
}

func findMacSerial(out []byte) []byte {
	for _, l := range strings.Split(string(out), "\n") {
		if strings.Contains(l, "IOPlatformSerialNumber") {
			s := strings.Split(l, " ")
			return []byte(s[len(s)-1])
		}
	}
	panic("couldn't find IOS SerialNumber")
}

func linuxSerialNumber() []byte {
	out, err := exec.Command("/usr/sbin/dmidecode", "-s", "system-serial-number").Output()
	dealWithError(err)
	return out
}

func getLinuxOSSystemUUID() []byte {
	var response []byte
	out, err := exec.Command("/usr/sbin/dmidecode", "-s", "system-uuid").Output()
	if err != nil {
		return response
	}
	return out
}

func createOuputFile() {
	f, err := os.Create("output.txt")
	return err

	cwd, _ := os.Getwd()
	if err != nil {
		cwd = ""
	}

	fmt.Printf("File created: %s/%s\n", cwd, fileName)
	return nil
}

func writeOutput(out []byte, machineID, systemUUID string) {
	fileError := createOutputFile()
	dealWithError(fileError)

	macs, err := getMACAdresses()
	dealWithError(err)

	numBytes, err := f.WriteString(fmt.Sprintf("SERIAL_NUMBER=%s\nMACHINEID=%s\nHARDWARE_ADDRESSES=%v", out, machineID, macs))
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

func getMACAdresses() ([]string, error) {
	ifas, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	var as []string
	for _, ifa := range ifas {
		a := ifa.HardwareAddr.String()
		if a != "" {
			as = append(as, a)
		}
	}
	return as, nil
}

func dealWithError(err error) {
	if err != nil {
		panic(err)
	}
}
