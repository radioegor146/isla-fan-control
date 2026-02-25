package main

import (
	"errors"

	"github.com/NVIDIA/go-nvml/pkg/nvml"

	"log/slog"
)

type NVML struct {
	devices []nvml.Device
}

func NewNVML() *NVML {
	ret := nvml.Init()
	if !errors.Is(ret, nvml.SUCCESS) {
		slog.Warn("NVML initialization failure, all temperatures will return 100°C", "error", ret)
		return &NVML{
			devices: []nvml.Device{},
		}
	}
	count, ret := nvml.DeviceGetCount()
	if !errors.Is(ret, nvml.SUCCESS) {
		slog.Warn("failed to get device count, all temperatures will return 100°C", "error", ret)
		return &NVML{
			devices: []nvml.Device{},
		}
	}
	devices := make([]nvml.Device, 0)
	for i := 0; i < count; i++ {
		device, ret := nvml.DeviceGetHandleByIndex(i)
		if !errors.Is(ret, nvml.SUCCESS) {
			slog.Warn("failed to get device, all temperatures will return 100°C", "deviceId", i, "error", ret)
		}
		slog.Info("NVML device found", "device", device)
		devices = append(devices, device)
	}
	slog.Info("NVML initialization succeeded", "deviceCount", len(devices))
	return &NVML{
		devices: devices,
	}
}

func (n *NVML) GetDeviceTemperature(deviceId int) uint32 {
	if len(n.devices) < deviceId {
		slog.Warn("device not found, returning 100°C", "deviceId", deviceId)
		return 100
	}
	device := n.devices[deviceId]
	temperature, ret := device.GetTemperature(nvml.TEMPERATURE_GPU)
	if !errors.Is(ret, nvml.SUCCESS) {
		slog.Warn("failed to get device temperature, returning 100°C", "deviceId", deviceId)
		return 100
	}
	return temperature
}

func (n *NVML) GetDeviceUUID(deviceId int) string {
	if len(n.devices) < deviceId {
		slog.Warn("device not found, returning 'unknown'", "deviceId", deviceId)
		return "unknown"
	}
	device := n.devices[deviceId]
	uuid, result := device.GetUUID()
	if !errors.Is(result, nvml.SUCCESS) {
		slog.Warn("failed to get device uuid, returning 'unknown'", "deviceId", deviceId)
		return "unknown"
	}
	return uuid
}

func (n *NVML) GetDeviceCount() int {
	return len(n.devices)
}
