package fritzbox

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const (
	devicePath = "/webservices/homeautoswitch.lua"
)

// DeviceService handles fritz!Box devices.
type DeviceService struct {
	c *Client
}

// List returns a list of all devices
func (s *DeviceService) List() ([]*Device, error) {
	u, err := commandURL("getdevicelistinfos", nil)
	if err != nil {
		return nil, err
	}
	req, err := s.c.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	var deviceList deviceList
	if _, err := s.c.Do(req, &deviceList); err != nil {
		return nil, err
	}
	return deviceList.Devices, nil
}

// Get returns a single device.
func (s *DeviceService) Get(ain string) (*Device, error) {
	list, err := s.List()
	if err != nil {
		return nil, err
	}
	for _, device := range list {
		if cleanAin(device.Identifier) == ain {
			return device, nil
		}
	}
	return nil, fmt.Errorf("device %q could not be found", ain)
}

// cleanAin removes spaces.
func cleanAin(str string) string {
	return strings.Replace(str, " ", "", -1)
}

func precheck(d *Device, lock bool) error {
	switch {
	case !d.IsConnected():
		return fmt.Errorf("device %q is not connected", d.Identifier)
	case d.IsLocked() && lock:
		return fmt.Errorf("device %q is locked; please unlock device via gui", d.Identifier)
	}
	return nil
}

// TurnOn turns a socket/thermostat on. grrr..
func (s *DeviceService) TurnOn(d *Device) (bool, error) {
	var u *url.URL
	var err error

	if err := precheck(d, true); err != nil {
		return false, err
	}

	params := map[string]string{
		"ain": cleanAin(d.Identifier),
	}

	switch {
	case d.IsSocket():
		u, err = commandURL("setswitchon", params)
	case d.IsThermostat():
		// According to https://avm.de/fileadmin/user_upload/Global/Service/Schnittstellen/AHA-HTTP-Interface.pdf, 254 means on.
		params["param"] = "254"
		u, err = commandURL("sethkrtsoll", params)
	}
	if err != nil {
		return false, err
	}

	req, err := s.c.NewRequest("GET", u.String(), nil)
	if err != nil {
		return false, err
	}
	if _, err = s.c.Do(req, nil); err != nil {
		return false, err
	}
	return true, nil
}

// TurnOff turns a socket/thermostat off.
func (s *DeviceService) TurnOff(d *Device) (bool, error) {
	var u *url.URL
	var err error

	if err := precheck(d, true); err != nil {
		return false, err
	}

	params := map[string]string{
		"ain": cleanAin(d.Identifier),
	}
	switch {
	case d.IsSocket():
		u, err = commandURL("setswitchoff", params)
	case d.IsThermostat():
		// According to https://avm.de/fileadmin/user_upload/Global/Service/Schnittstellen/AHA-HTTP-Interface.pdf, 253 means off.
		params["param"] = "253"
		u, err = commandURL("sethkrtsoll", params)
	}
	if err != nil {
		return false, err
	}

	req, err := s.c.NewRequest("GET", u.String(), nil)
	if err != nil {
		return false, err
	}
	var buf bytes.Buffer
	if _, err = s.c.Do(req, &buf); err != nil {
		return false, err
	}
	str := strings.TrimSpace(buf.String())
	if d.IsThermostat() {
		return str == "253", nil
	}
	return strconv.ParseBool(strings.TrimSpace(buf.String()))
}

// Toggle will turn a socket on, if it is off. Or it will turn a socket off,
// if it is on.
func (s *DeviceService) Toggle(d *Device) (bool, error) {
	if err := precheck(d, true); err != nil {
		return false, err
	}
	if !d.IsSocket() {
		return false, fmt.Errorf("device %q does not support toggling", d.Identifier)
	}
	params := map[string]string{
		"ain": cleanAin(d.Identifier),
	}
	u, err := commandURL("setswitchtoggle", params)
	if err != nil {
		return false, err
	}
	req, err := s.c.NewRequest("GET", u.String(), nil)
	if err != nil {
		return false, err
	}
	if _, err = s.c.Do(req, nil); err != nil {
		return false, err
	}
	return true, nil
}

// GetPower returns the power currently consumed.
func (s *DeviceService) GetPower(d *Device) (int64, error) {
	if err := precheck(d, false); err != nil {
		return 0, err
	}
	if !d.HasEnergy() {
		return 0, fmt.Errorf("device %q does not support getting power", d.Identifier)
	}
	params := map[string]string{
		"ain": cleanAin(d.Identifier),
	}
	u, err := commandURL("getswitchpower", params)
	if err != nil {
		return 0, err
	}
	req, err := s.c.NewRequest("GET", u.String(), nil)
	if err != nil {
		return 0, err
	}
	var buf bytes.Buffer
	if _, err = s.c.Do(req, &buf); err != nil {
		return 0, err
	}
	str := strings.TrimSpace(buf.String())
	power, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, err
	}
	return power, nil
}

// GetEnergy returns the energy since last reset :D.
func (s *DeviceService) GetEnergy(d *Device) (int64, error) {
	if err := precheck(d, false); err != nil {
		return 0, err
	}
	if !d.HasEnergy() {
		return 0, fmt.Errorf("device %q does not support power", d.Identifier)
	}
	params := map[string]string{
		"ain": cleanAin(d.Identifier),
	}
	u, err := commandURL("getswitchenergy", params)
	if err != nil {
		return 0, err
	}
	req, err := s.c.NewRequest("GET", u.String(), nil)
	if err != nil {
		return 0, err
	}
	var buf bytes.Buffer
	if _, err = s.c.Do(req, &buf); err != nil {
		return 0, err
	}
	str := strings.TrimSpace(buf.String())
	power, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, nil
	}
	return power, nil
}

// GetTemperature returns the device's current temperature.
func (s *DeviceService) GetTemperature(d *Device) (float64, error) {
	if err := precheck(d, false); err != nil {
		return 0, err
	}
	if !d.HasTemperature() {
		return 0, fmt.Errorf("device %q does not support getting temperature", d.Identifier)
	}

	params := map[string]string{
		"ain": cleanAin(d.Identifier),
	}

	u, err := commandURL("gettemperature", params)
	req, err := s.c.NewRequest("GET", u.String(), nil)
	if err != nil {
		return 0, err
	}
	var buf bytes.Buffer
	if _, err = s.c.Do(req, &buf); err != nil {
		return 0, err
	}
	str := strings.TrimSpace(buf.String())
	temp, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, err
	}
	// temp is 200 = 20°
	return float64(temp) / 10.0, nil
}

// GetSollTemperature returns the thermostat's desired temperature.
func (s *DeviceService) GetSollTemperature(d *Device) (float64, error) {
	if !d.IsThermostat() {
		return 0, fmt.Errorf("device %q does not support getting soll temperature", d.Identifier)
	}
	if err := precheck(d, false); err != nil {
		return 0, err
	}
	params := map[string]string{
		"ain": cleanAin(d.Identifier),
	}
	u, err := commandURL("gethkrtsoll", params)
	if err != nil {
		return 0, err
	}
	req, err := s.c.NewRequest("GET", u.String(), nil)
	if err != nil {
		return 0, err
	}
	var buf bytes.Buffer
	if _, err = s.c.Do(req, &buf); err != nil {
		return 0, err
	}
	temp, err := strconv.ParseInt(strings.TrimSpace(buf.String()), 10, 64)
	if err != nil {
		return 0, err
	}
	fmt.Println(temp)
	if temp == 253 || temp == 254 {
		return 0, fmt.Errorf("device %q is off", d.Identifier)
	}
	// Temperature is between 16 and 58. 16 represents 8° and 17 is 8,5°.
	return float64(temp) / 2, nil
}

// SetSollTemperature sets the thermostat's  desired temperature.
func (s *DeviceService) SetSollTemperature(d *Device, temp float64) error {
	if err := precheck(d, true); err != nil {
		return err
	}
	if !d.IsThermostat() {
		return fmt.Errorf("Device %q does not support setting soll temperature", d.Identifier)
	}
	if temp < 8 || temp > 28 {
		return fmt.Errorf("temperature needs to be between 8 and 28")
	}
	params := map[string]string{
		"ain": cleanAin(d.Identifier),
	}
	// Temperature is between 16 and 58. 16 represents 8° and 17 is 8,5°.
	params["param"] = fmt.Sprintf("%2.0f", temp*2)
	u, err := commandURL("sethkrtsoll", params)
	if err != nil {
		return err
	}
	req, err := s.c.NewRequest("GET", u.String(), nil)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if _, err = s.c.Do(req, &buf); err != nil {
		return err
	}
	if strings.TrimSpace(buf.String()) != params["param"] {
		return fmt.Errorf("new temperature does not match desired temperature: %s", buf.String())
	}
	return nil
}

func commandURL(cmd string, cmds map[string]string) (*url.URL, error) {
	u, err := url.Parse(devicePath)
	if err != nil {
		return nil, err
	}
	params := u.Query()
	params.Add("switchcmd", cmd)
	for k, v := range cmds {
		params.Add(k, v)
	}
	u.RawQuery = params.Encode()
	return u, nil
}

// DeviceList represents a list of devices returned by the fritz!Box.
type deviceList struct {
	XMLName xml.Name  `xml:"devicelist"`
	Version string    `xml:"version,attr"`
	Devices []*Device `xml:"device"`
}

// Device represents a device returned by the fritz!Box.
type Device struct {
	XMLName         xml.Name `xml:"device"`
	Identifier      string   `xml:"identifier,attr"`
	Connected       bool     `xml:"present"`
	FunctionBitMask uint32   `xml:"functionbitmask,attr"`
	Firmware        string   `xml:"fwversion,attr"`
	Manufacturer    string   `xml:"manufacturer,attr"`
	Name            string   `xml:"productname,attr"`
	Lock            bool     `xml:"switch>lock"`
}

// IsConnected reports whether a device is connected.
func (d *Device) IsConnected() bool {
	return d.Connected
}

// IsLocked reports whether a device is locked.
func (d *Device) IsLocked() bool {
	return d.Lock
}

// IsThermostat reports whether a device is a thermostat.
func (d *Device) IsThermostat() bool {
	return d.FunctionBitMask&(1<<6) != 0
}

// IsSocket reports whether a device is a socket.
func (d *Device) IsSocket() bool {
	return d.FunctionBitMask&(1<<9) != 0
}

// IsAlarm reports whether a device allow alerts.
func (d *Device) IsAlarm() bool {
	return d.FunctionBitMask&(1<<4) != 0
}

// HasTemperature reports whether a device monitors its temperature.
func (d *Device) HasTemperature() bool {
	return d.FunctionBitMask&(1<<8) != 0
}

// HasEnergy reports whether a device monitors its energy.
func (d *Device) HasEnergy() bool {
	return d.FunctionBitMask&(1<<7) != 0

}

// IsDECTRepeater reports whether a device is a DECTRepeater.
func (d *Device) IsDECTRepeater() bool {
	return d.FunctionBitMask&(1<<10) != 0
}
