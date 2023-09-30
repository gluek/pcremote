package utils

import (
	"fmt"
	"strings"
	"github.com/abdfnx/gosh"
	"unicode"
)


func GetAudioDevices() string {
	err, out, errout := gosh.PowershellOutput("Get-AudioDevice -List")
	if err != nil {
		fmt.Println(errout)
	}
	return out
}

func SetDefaultAudioDevice(deviceID string) {
	var cmd string
	cmd = fmt.Sprintf("Set-AudioDevice -ID \"%s\"", deviceID)
	cmd = strings.Map(func(r rune) rune {
        if unicode.IsPrint(r) {
            return r
        }
        return -1
    }, cmd)
	
	gosh.PowershellCommand(cmd)
	
}

func FindID(name string) string {
	var id string = ""
	devices := GetAudioDevices()
	splitDevices := strings.Split(devices, "\n")
	for i:=0; i<len(splitDevices); i++ {
		if strings.Contains(splitDevices[i], name) {
			splitLine := strings.Split(splitDevices[i+1], " ") 
			lastElement := splitLine[len(splitLine)-1]
			id = lastElement
		}
	}
	return id
}
