package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/denisbrodbeck/machineid"
)

func main() {
	_, err := os.Create("output.txt")
	if err != nil {
		panic(err)
	}
	machineID, err := machineid.ID()
	if err != nil {
		machineID = ""
	}

	var output []byte

	switch runtime.GOOS {
	case "windows":
		output = getWindowsOSSerial()
	case "darwin":
		output = getMacOSSerial()
	case "linux":
		output = getLinuxOSSerial()
	default:
		panic("couldn't detect os")
	}
	writeOutput(output, machineID)
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

func getWindowsOSSerial() []byte {
	out, err := exec.Command("wmic", "bios", "get", "serialnumber").Output()
	if err != nil {
		log.Fatalf("Something went wrong trying to execute the serial number command: \n %q", err)
	}
	return out
}

func getMacOSSerial() []byte {
	out, err := exec.Command("/usr/sbin/ioreg", "-l").Output()
	if err != nil {
		log.Fatalf("Something went wrong trying to execute the serial number command: \n %q", err)
	}
	return findMacSerial(out)
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

func getLinuxOSSerial() []byte {
	out, err := exec.Command("/usr/sbin/dmidecode", "-s", "system-serial-number").Output()
	if err != nil {
		panic(err)
	}
	return out
}

func writeOutput(out []byte, machineID string) {
	f, err := os.Create("output.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	cwd, _ := os.Getwd()
	if err != nil {
		cwd = ""
	}
	fmt.Printf("File created: %s/output.txt\n", cwd)
	macs, err := getMACAdresses()
	if err != nil {
		panic(err)
	}
	numBytes, err := f.WriteString(fmt.Sprintf("SERIAL_NUMBER=%s\nMACHINEID=\"%s\"\nHARDWARE_ADDRESSES=%v", out, machineID, macs))
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
