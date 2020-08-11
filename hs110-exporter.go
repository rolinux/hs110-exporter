package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	// fork from https://github.com/sausheong/hs1xxplug to keep it safe
	"github.com/rolinux/hs1xxplug"
)

var (
	hs110RelayState = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "hs110_relay_state",
		Help: "Plug On or Off state",
	},
		[]string{
			// target hostname or IP
			"target",
			// MAC address of the plug
			"mac",
			// plug alias
			"alias",
		},
	)

	hs110OnTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "hs110_on_time",
		Help: "The number of seconds from the last relay_state change to On",
	},
		[]string{
			// target hostname or IP
			"target",
			// MAC address of the plug
			"mac",
			// plug alias
			"alias",
		},
	)

	hs110VoltageMilliVolts = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "hs110_voltage_milivolts",
		Help: "The number of voltage millivolts passing through HS110 in the last minute",
	},
		[]string{
			// target hostname or IP
			"target",
			// MAC address of the plug
			"mac",
			// plug alias
			"alias",
		},
	)

	hs110CurrentMilliAmps = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "hs110_current_miliamps",
		Help: "The number of current milliamps passing through HS110 in the last minute",
	},
		[]string{
			// target hostname or IP
			"target",
			// MAC address of the plug
			"mac",
			// plug alias
			"alias",
		},
	)

	hs110PowerMilliWatts = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "hs110_power_milliwatts",
		Help: "The number of power milliwatts passing through HS110 in the last minute",
	},
		[]string{
			// target hostname or IP
			"target",
			// MAC address of the plug
			"mac",
			// plug alias
			"alias",
		},
	)

	// hope to be a counter
	hs110TotalWattHours = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "hs110_total_watthours",
		Help: "Total number of watt hours passing through HS110 from last check",
	},
		[]string{
			// target hostname or IP
			"target",
			// MAC address of the plug
			"mac",
			// plug alias
			"alias",
		},
	)
)

type hs110 struct {
	System struct {
		GetSysinfo struct {
			SwVer      string `json:"sw_ver"`
			HwVer      string `json:"hw_ver"`
			Type       string `json:"type"`
			Model      string `json:"model"`
			MAC        string `json:"mac"`
			DevName    string `json:"dev_name"`
			Alias      string `json:"alias"`
			RelayState int    `json:"relay_state"`
			OnTime     int    `json:"on_time"`
			ActiveMode string `json:"active_mode"`
			Feature    string `json:"feature"`
			Updating   int    `json:"updating"`
			IconHash   string `json:"icon_hash"`
			Rssi       int    `json:"rssi"`
			LedOff     int    `json:"led_off"`
			LongitudeI int    `json:"longitude_i"`
			LatitudeI  int    `json:"latitude_i"`
			HwID       string `json:"hwId"`
			FwID       string `json:"fwId"`
			DeviceID   string `json:"deviceId"`
			OemID      string `json:"oemId"`
			NextAction struct {
				Type int `json:"type"`
			} `json:"next_action"`
			NtcState int `json:"ntc_state"`
			ErrCode  int `json:"err_code"`
		} `json:"get_sysinfo"`
	} `json:"system"`
	Emeter struct {
		GetRealtime struct {
			VoltageMv int `json:"voltage_mv"`
			CurrentMa int `json:"current_ma"`
			PowerMw   int `json:"power_mw"`
			TotalWh   int `json:"total_wh"`
			ErrCode   int `json:"err_code"`
		} `json:"get_realtime"`
		GetVgainIgain struct {
			Vgain   int `json:"vgain"`
			Igain   int `json:"igain"`
			ErrCode int `json:"err_code"`
		} `json:"get_vgain_igain"`
	} `json:"emeter"`
}

func recordMetrics() {
	go func() {
		for {
			targetHS110 := os.Getenv("TARGET_HS110")
			// for TARGET-HS110 env variable use DNS or IP
			plug := hs1xxplug.Hs1xxPlug{IPAddress: targetHS110}
			results, err := plug.MeterInfo()
			if err != nil {
				log.Println("Target not responding - err:", err)
				// if target not responding sleep
				time.Sleep(1 * time.Minute)
				continue
			}

			var meterInfo hs110

			logSlice := []string{}

			err = json.Unmarshal([]byte(results), &meterInfo) // here!

			if err != nil {
				log.Println("Target not a HS110 - err:", err)
				// if target not responding sleep
				time.Sleep(1 * time.Minute)
				continue
			}

			// relay_state 1 or 0
			relayState := meterInfo.System.GetSysinfo.RelayState

			// On time (uptime in seconds)
			onTime := meterInfo.System.GetSysinfo.OnTime

			// MAC address
			mac := meterInfo.System.GetSysinfo.MAC

			// alias
			alias := meterInfo.System.GetSysinfo.Alias

			// voltage in milli volts
			voltageMv := meterInfo.Emeter.GetRealtime.VoltageMv

			// current in milli amps
			currentMa := meterInfo.Emeter.GetRealtime.CurrentMa

			// power in milli Watt
			powerMw := meterInfo.Emeter.GetRealtime.PowerMw

			// total watt hours
			totalWh := meterInfo.Emeter.GetRealtime.TotalWh

			// this should be set
			logSlice = append(logSlice, fmt.Sprintf("relay_state:%d", relayState))
			hs110RelayState.WithLabelValues(targetHS110, mac, alias).Set(float64(relayState))

			// this should be set
			logSlice = append(logSlice, fmt.Sprintf("on_time:%d", onTime))
			hs110OnTime.WithLabelValues(targetHS110, mac, alias).Set(float64(onTime))

			// a HS100 will return 0
			if totalWh != 0 {
				// totalWh can still be above 0 with current powerMw == 0
				logSlice = append(logSlice, fmt.Sprintf("watthours:%d", totalWh))
				hs110TotalWattHours.WithLabelValues(targetHS110, mac, alias).Set(float64(totalWh))

				if powerMw != 0 {
					logSlice = append(logSlice, fmt.Sprintf("millivolts:%d", voltageMv))
					hs110VoltageMilliVolts.WithLabelValues(targetHS110, mac, alias).Set(float64(voltageMv))

					logSlice = append(logSlice, fmt.Sprintf("milliamps:%d", currentMa))
					hs110CurrentMilliAmps.WithLabelValues(targetHS110, mac, alias).Set(float64(currentMa))

					logSlice = append(logSlice, fmt.Sprintf("milliwatts:%d", powerMw))
					hs110PowerMilliWatts.WithLabelValues(targetHS110, mac, alias).Set(float64(powerMw))
				}
				log.Printf("Target '%s' HS110 data: %s\n", targetHS110, strings.Join(logSlice, " "))
			} else {
				log.Printf("Target '%s' HS110 data: %s\n", targetHS110, strings.Join(logSlice, " "))
				log.Printf("Target '%s' not a HS110 - err: %v\n", targetHS110, err)
				// if target not responding sleep
				time.Sleep(1 * time.Minute)
			}
		}
	}()
}

func init() {
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(hs110RelayState)
	prometheus.MustRegister(hs110OnTime)
	prometheus.MustRegister(hs110VoltageMilliVolts)
	prometheus.MustRegister(hs110CurrentMilliAmps)
	prometheus.MustRegister(hs110PowerMilliWatts)
	prometheus.MustRegister(hs110TotalWattHours)
}

func main() {
	recordMetrics()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":9498", nil)
}
