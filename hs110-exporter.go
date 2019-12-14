package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	// fork from https://github.com/sausheong/hs1xxplug to keep it safe
	"github.com/rolinux/hs1xxplug"
)

var (
	hs110Miliwatts = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "hs110_miliwatts",
		Help: "The number of miliwatts passing through HS110 in the last minute",
	},
		[]string{
			// target hostname or IP
			"target",
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
			Mac        string `json:"mac"`
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
			targetHS110 := os.Getenv("TARGET-HS110")
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

			err = json.Unmarshal([]byte(results), &meterInfo) // here!

			if err != nil {
				log.Println("Target not a HS110 - err:", err)
				// if target not responding sleep
				time.Sleep(1 * time.Minute)
				continue
			}
			powerMw := meterInfo.Emeter.GetRealtime.PowerMw
			// a HS100 will return 0
			if powerMw != 0 {
				log.Printf("Target '%s' HS110 microwatts are '%d'\n", targetHS110, powerMw)
				hs110Miliwatts.WithLabelValues(targetHS110).Set(float64(powerMw))
			} else {
				log.Println("Target not a HS110 - err:", err)
				// if target not responding sleep
				time.Sleep(1 * time.Minute)
			}
		}
	}()
}

func init() {
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(hs110Miliwatts)
}

func main() {
	recordMetrics()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":9498", nil)
}
