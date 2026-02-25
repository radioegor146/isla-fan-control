# isla-fan-control

Simple Go service to control fans on Gooxi motherboard using temperature from NVML

Sample config (mine):
```yaml
update_interval: 10s
fans:
  1: # cpu fan
    gpus: []
    curve: {}
    default_speed: 100
    hysteresis: 0
  2: &gpu-fan # gpu fan 1
    gpus:
      - GPU-6e0f078b-1c3d-a0c8-c394-d2058633c52b # left 1
      - GPU-1af51d39-c4d6-4137-5d03-8c7ed6578b90 # right 1
    curve:
      20: 10
      30: 20
      40: 20
      45: 30
      50: 40
      55: 50
      60: 60
      65: 80
      70: 100
    default_speed: 10
    hysteresis: 2
  6: # gpu fan 2
    <<: *gpu-fan
    gpus:
      - GPU-d36b4fdd-78a1-c22b-d97b-7469027f7a5e # left 2
      - GPU-e688fa0e-32d1-400b-3a4a-e19f7eee5360 # right 2
  4: # gpu fan 3
    <<: *gpu-fan
    gpus:
      - GPU-81ff54d5-4fcc-3b62-6918-cd42f3be662f # left 3
      - GPU-47fd88c7-da07-0f42-a1f7-69de341860c4 # right 3
  7: # gpu fan 3
    <<: *gpu-fan
    gpus:
      - GPU-d7802c83-c228-94da-eead-917699bff114 # left 4
      - GPU-e1e51e97-371d-77c5-5199-d72ac98713e4 # right 4
```