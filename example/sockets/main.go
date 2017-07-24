package main

import (
	"fmt"
	"log"
	"net/url"

	"github.com/philippfranke/go-fritzbox/fritzbox"
)

func main() {
	c := fritzbox.NewClient(nil)
	u, _ := url.Parse("")
	c.BaseURL = u
	if err := c.Auth("", ""); err != nil {
		log.Fatalf("auth: %+v", err)
	}

	devices, err := c.DeviceService.List()
	if err != nil {
		log.Fatalf("list: %+v", err)
	}

	for _, device := range devices {
		fmt.Printf("DeviceID: %s Name: %s\n", device.Identifier, device.Name)
		// Turn device on.
		on, err := c.DeviceService.TurnOn(device)
		fmt.Printf("\tTurn on: %t, error: %+v\n", on, err)

		// GetPower
		power, err := c.DeviceService.GetPower(device)
		if err != nil {
			fmt.Printf("\terror: %+v\n", err)
		} else {
			fmt.Printf("\tPower: %d mW\n", power)
		}
		// GetEnergy
		energy, err := c.DeviceService.GetEnergy(device)
		if err != nil {
			fmt.Printf("\terror: %+v\n", err)
		} else {
			fmt.Printf("\tPower: %d Wh\n", energy)
		}

		// GetTemperature
		temp, err := c.DeviceService.GetTemperature(device)
		if err != nil {
			fmt.Printf("\terror: %+v\n", err)
		} else {
			fmt.Printf("\tTemperature: %2.2f°\n", temp)
		}

		// GetSollTemperature
		sollTemp, err := c.DeviceService.GetSollTemperature(device)
		if err != nil {
			fmt.Printf("\terror: %+v\n", err)
		} else {
			fmt.Printf("\tTemperature: %2.2f°\n", sollTemp)
		}

		// SetSollTemperature
		if err := c.DeviceService.SetSollTemperature(device, 24); err != nil {
			fmt.Printf("\terror: %+v\n", err)
		} else {
			fmt.Printf("\tSet to 24°\n")
		}

		// GetSollTemperature
		sollTemp, err = c.DeviceService.GetSollTemperature(device)
		if err != nil {
			fmt.Printf("\terror: %+v\n", err)
		} else {
			fmt.Printf("\tTemperature: %2.2f°\n", sollTemp)
		}

		// Turn device off.
		off, err := c.DeviceService.TurnOff(device)
		fmt.Printf("\tTurn off: %t, error: %+v\n", off, err)

		toggle, err := c.DeviceService.Toggle(device)
		if err != nil {
			fmt.Printf("\terror: %+v\n", err)
		} else {
			fmt.Printf("\ttoggle: %t\n", toggle)
		}
	}
}
