package audio

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/moutend/go-wca/pkg/wca"
)

// https://learn.microsoft.com/en-us/windows/win32/coreaudio/

type iIPolicyConfigVista struct {
	ole.IUnknown
}

type iPolicyConfigVistaVtbl struct {
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

func getAllDevices() ([]audioDevice, error) {
	deviceList := []audioDevice{}
	var mmde *wca.IMMDeviceEnumerator
	var mmdc *wca.IMMDeviceCollection
	var mmd *wca.IMMDevice
	var props *wca.IPropertyStore

	if err = ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
		return []audioDevice{}, err
	}
	defer ole.CoUninitialize()

	if err = wca.CoCreateInstance(wca.CLSID_MMDeviceEnumerator, 0, wca.CLSCTX_ALL, wca.IID_IMMDeviceEnumerator, &mmde); err != nil {
		return []audioDevice{}, err
	}
	defer mmde.Release()

	if err = mmde.EnumAudioEndpoints(wca.ERender, wca.DEVICE_STATE_ACTIVE, &mmdc); err != nil {
		return []audioDevice{}, err
	}
	defer mmdc.Release()

	var deviceCount uint32
	if err = mmdc.GetCount(&deviceCount); err != nil {
		return []audioDevice{}, err
	}
	var i uint32
	for i = 0; i < deviceCount; i++ {
		if err = mmdc.Item(i, &mmd); err != nil {
			return []audioDevice{}, err
		}

		var deviceID string
		if err = mmd.GetId(&deviceID); err != nil {
			return []audioDevice{}, err
		}

		if err = mmd.OpenPropertyStore(wca.STGM_READ, &props); err != nil {
			return []audioDevice{}, err
		}

		var varName wca.PROPVARIANT
		if err = props.GetValue(&wca.PKEY_Device_FriendlyName, &varName); err != nil {
			return []audioDevice{}, err
		}
		deviceName := varName.String()

		deviceList = append(deviceList, audioDevice{id: deviceID, friendly_name: deviceName})
	}

	return deviceList, nil
}

func (device *audioDevice) setDefaultEndpointByID() error {
	GUID_IPolicyConfigVista := ole.NewGUID("{568b9108-44bf-40b4-9006-86afe5b5a620}")
	GUID_CPolicyConfigVistaClient := ole.NewGUID("{294935CE-F637-4E7C-A41B-AB255460B862}")
	var policyConfig *iIPolicyConfigVista

	if err = ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
		return err
	}
	defer ole.CoUninitialize()

	if err = wca.CoCreateInstance(GUID_CPolicyConfigVistaClient, 0, wca.CLSCTX_ALL, GUID_IPolicyConfigVista, &policyConfig); err != nil {
		return err
	}
	defer policyConfig.Release()

	if err = policyConfig.SetDefaultEndpoint(device.id, wca.EConsole); err != nil {
		return err
	}
	return nil
}

func getDeviceIDByName(deviceName string) (deviceID string, err error) {
	deviceList, err := getAllDevices()
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

func (v *iIPolicyConfigVista) VTable() *iPolicyConfigVistaVtbl {
	return (*iPolicyConfigVistaVtbl)(unsafe.Pointer(v.RawVTable))
}

func (v *iIPolicyConfigVista) SetDefaultEndpoint(deviceID string, eRole wca.ERole) (err error) {
	err = pcvSetDefaultEndpoint(v, deviceID, eRole)
	return
}

func pcvSetDefaultEndpoint(pcv *iIPolicyConfigVista, deviceID string, eRole wca.ERole) (err error) {
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
