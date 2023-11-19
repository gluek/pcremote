# PCRemote
This project is a small traybar tool which reacts to mqtt messages to execute pc routines.

## Features
- Change default audio devices

## Usage

Example config.json:
```json
{
    "broker": {
        "ip": "127.0.0.1",
        "port": "1883"
    },
    "audiodevices": {
        "device1": {
            "friendly_name": "Speaker",
            "topic": "pc/sound/device/speaker"
        },
        "device2": {
            "friendly_name": "Soundbar",
            "topic": "pc/sound/device/soundbar"
        }
    }
}
```

### Audio
Any message send to the registered `topic` of the audio device triggers a routine which will change the default audio device to `friendly_name`, which needs to be a substring of the actual device name.