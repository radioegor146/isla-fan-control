package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"time"
)

const GooxiTimeout = 5 * time.Second

func getContextWithTimeout(duration time.Duration) context.Context {
	nextContext, _ := context.WithTimeout(context.Background(), duration)
	return nextContext
}

func getDifference(a, b uint32) uint32 {
	if a >= b {
		return a - b
	}
	return b - a
}

func calculateFanValue(config FanConfig, nvml *NVML, uuidMapping map[string]int, lastTemperatureAtSet uint32) (*byte, uint32) {
	if len(config.GPUs) == 0 || len(config.Curve) == 0 {
		return nil, lastTemperatureAtSet
	}
	var temperature uint32
	temperature = 0
	for _, uuid := range config.GPUs {
		id, has := uuidMapping[uuid]
		if !has {
			slog.Warn("failed to find device by uuid", "uuid", uuid)
		}
		temperature = max(temperature, nvml.GetDeviceTemperature(id))
	}
	if getDifference(temperature, lastTemperatureAtSet) <= config.Hysteresis {
		return nil, lastTemperatureAtSet
	}

	curve := config.Curve

	if len(curve) == 1 {
		for _, v := range curve {
			return &v, temperature
		}
	}

	ts := make([]uint32, 0, len(curve))
	for t := range curve {
		ts = append(ts, t)
	}
	sort.Slice(ts, func(i, j int) bool { return ts[i] < ts[j] })

	tMin, tMax := ts[0], ts[len(ts)-1]
	vMin, vMax := curve[tMin], curve[tMax]

	if temperature <= tMin {
		return &vMin, temperature
	}
	if temperature >= tMax {
		return &vMax, temperature
	}

	for i := 0; i < len(ts)-1; i++ {
		t0, t1 := ts[i], ts[i+1]
		if temperature == t0 {
			value := curve[t0]
			return &value, temperature
		}
		if temperature > t0 && temperature <= t1 {
			v0 := curve[t0]
			v1 := curve[t1]

			if t1 == t0 {
				return &v1, temperature
			}

			ratio := float64(temperature-t0) / float64(t1-t0)
			out := float64(v0) + ratio*(float64(v1)-float64(v0))

			if out <= 0 {
				out = 0
			}
			if out >= 100 {
				out = 100
			}
			value := byte(out + 0.5)
			return &value, temperature
		}
	}

	return &vMax, temperature
}

func main() {
	if len(os.Args) < 2 {
		_, _ = fmt.Fprintln(os.Stderr, "usage: isla-fan-control <config path>")
		os.Exit(1)
	}

	configPath := os.Args[1]
	config, err := ParseConfig(configPath)
	if err != nil {
		panic(err)
	}

	nvml := NewNVML()
	gooxi, err := NewLocalGooxi()
	if err != nil {
		panic(err)
	}

	err = gooxi.SetFanMode(getContextWithTimeout(GooxiTimeout), false)
	if err != nil {
		panic(err)
	}

	currentFanValues := make(map[byte]byte)
	for fanId, fanConfig := range config.Fans {
		currentFanValues[fanId] = fanConfig.DefaultSpeed
		err = gooxi.SetFanSpeed(getContextWithTimeout(GooxiTimeout), fanId, fanConfig.DefaultSpeed)
	}

	lastTemperaturesAtSet := make(map[byte]uint32)

	gpuUuidMapping := make(map[string]int)

	for i := 0; i < nvml.GetDeviceCount(); i++ {
		uuid := nvml.GetDeviceUUID(i)
		gpuUuidMapping[uuid] = i
	}

	for {
		for fanId, fanConfig := range config.Fans {
			newFanValue, temperatureAtSet := calculateFanValue(fanConfig, nvml, gpuUuidMapping,
				lastTemperaturesAtSet[fanId])
			if newFanValue != nil {
				slog.Info("updating fan speed", "fanId", fanId, "value", *newFanValue, "temperature", temperatureAtSet)
				lastTemperaturesAtSet[fanId] = temperatureAtSet
				if *newFanValue != currentFanValues[fanId] {
					err = gooxi.SetFanSpeed(getContextWithTimeout(GooxiTimeout), fanId, *newFanValue)
					if err != nil {
						panic(err)
					}
					currentFanValues[fanId] = *newFanValue
				}
			}
		}
		time.Sleep(config.UpdateInterval)
	}
}
