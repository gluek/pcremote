package utils

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/go-ole/go-ole"
	"github.com/moutend/go-wca/pkg/wca"
)

// https://learn.microsoft.com/en-us/windows/win32/coreaudio/

var err error

type AudioDevice struct {
	id            string
	friendly_name string
}

type IPolicyConfigVista struct {
	ole.IUnknown
}

type IPolicyConfigVistaVtbl struct {
	ole.IUnknownVtbl
	GetMixFormat          uintptr
	GetDeviceFormat       uintptr
	SetDeviceFormat       uintptr
	GetProcessingPeriod   uintptr
	SetProcessingPeriod   uintptr
	GetShareMode          uintptr
	SetShareMode          uintptr
	GetPropertyValue      uintptr
	SetPropertyValue      uintptr
	SetDefaultEndpoint    uintptr
	SetEndpointVisibility uintptr
}

func AudioMessageRouter(msg mqtt.Message) error {
	switch msg.Topic() {
	case "computer/sound/device/speaker":
		speakerID, err := GetDeviceIDByName("Lautsprecher")
		if err != nil {
			return err
		}
		SetDefaultEndpointByID(speakerID)
	case "computer/sound/device/soundbar":
		soundbarID, err := GetDeviceIDByName("Cinebar")
		if err != nil {
			return err
		}
		SetDefaultEndpointByID(soundbarID)
	}
	return nil
}

func GetAllDevices() ([]AudioDevice, error) {
	deviceList := []AudioDevice{}
	var mmde *wca.IMMDeviceEnumerator
	var mmdc *wca.IMMDeviceCollection
	var mmd *wca.IMMDevice
	var props *wca.IPropertyStore

	if err = ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
		return []AudioDevice{}, err
	}
	defer ole.CoUninitialize()

	if err = wca.CoCreateInstance(wca.CLSID_MMDeviceEnumerator, 0, wca.CLSCTX_ALL, wca.IID_IMMDeviceEnumerator, &mmde); err != nil {
		return []AudioDevice{}, err
	}
	defer mmde.Release()

	if err = mmde.EnumAudioEndpoints(wca.ERender, wca.DEVICE_STATE_ACTIVE, &mmdc); err != nil {
		return []AudioDevice{}, err
	}
	defer mmdc.Release()

	var deviceCount uint32
	if err = mmdc.GetCount(&deviceCount); err != nil {
		return []AudioDevice{}, err
	}
	var i uint32
	for i = 0; i < deviceCount; i++ {
		if err = mmdc.Item(i, &mmd); err != nil {
			return []AudioDevice{}, err
		}

		var deviceID string
		if err = mmd.GetId(&deviceID); err != nil {
			return []AudioDevice{}, err
		}
		//fmt.Printf("DeviceID: %s\n", deviceID)

		if err = mmd.OpenPropertyStore(wca.STGM_READ, &props); err != nil {
			return []AudioDevice{}, err
		}

		var varName wca.PROPVARIANT
		if err = props.GetValue(&wca.PKEY_Device_FriendlyName, &varName); err != nil {
			return []AudioDevice{}, err
		}
		deviceName := varName.String()
		//fmt.Printf("Name: %s\n", deviceName)

		deviceList = append(deviceList, AudioDevice{id: deviceID, friendly_name: deviceName})
	}

	return deviceList, nil
}

func SetDefaultEndpointByID(deviceID string) error {
	GUID_IPolicyConfigVista := ole.NewGUID("{568b9108-44bf-40b4-9006-86afe5b5a620}")
	GUID_CPolicyConfigVistaClient := ole.NewGUID("{294935CE-F637-4E7C-A41B-AB255460B862}")
	var policyConfig *IPolicyConfigVista

	if err = ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
		return err
	}
	defer ole.CoUninitialize()

	if err = wca.CoCreateInstance(GUID_CPolicyConfigVistaClient, 0, wca.CLSCTX_ALL, GUID_IPolicyConfigVista, &policyConfig); err != nil {
		return err
	}
	defer policyConfig.Release()

	if err = policyConfig.SetDefaultEndpoint(deviceID, wca.EConsole); err != nil {
		return err
	}
	return nil
}

func GetDeviceIDByName(deviceName string) (deviceID string, err error) {
	deviceList, err := GetAllDevices()
	if err != nil {
		return "", err
	}
	for _, v := range deviceList {
		if strings.Contains(v.friendly_name, deviceName) {
			return v.id, nil
		}
	}
	return "", fmt.Errorf("device with name %s not found", deviceName)
}

func (v *IPolicyConfigVista) VTable() *IPolicyConfigVistaVtbl {
	return (*IPolicyConfigVistaVtbl)(unsafe.Pointer(v.RawVTable))
}

func (v *IPolicyConfigVista) SetDefaultEndpoint(deviceID string, eRole wca.ERole) (err error) {
	err = pcvSetDefaultEndpoint(v, deviceID, eRole)
	return
}

func pcvSetDefaultEndpoint(pcv *IPolicyConfigVista, deviceID string, eRole wca.ERole) (err error) {
	var ptr *uint16
	if ptr, err = syscall.UTF16PtrFromString(deviceID); err != nil {
		return
	}
	hr, _, _ := syscall.Syscall(
		pcv.VTable().SetDefaultEndpoint,
		3,
		uintptr(unsafe.Pointer(pcv)),
		uintptr(unsafe.Pointer(ptr)),
		uintptr(uint32(eRole)))
	if hr != 0 {
		err = ole.NewError(hr)
	}
	return
}
