# TP-Link HS110 Wi-Fi Smart Plug with Energy Monitoring Prometheus exporter

There are existing hs110-exporters written in other programming languages but due to [Sau Sheong Chang hsxxx Go library](https://github.com/sausheong/hs1xxplug) I decided to write my own exporter that is supposed to use less resources (CPU, memory and disk).

Please [feedback](https://github.com/rolinux/hs110-exporter/issues) if the metrics are not of the right type or you found any issues we can fix.

The exporter has been running for several months against 3 plugs (1 container per plug) without any obvious issues.

I run my hs110-exporter(s) using a command like:

```
$ sudo docker run -dit --restart always -e TARGET_HS110=192.168.252.57 \
 -p 9498:9498 --name hs110-exporter-57 ghcr.io/rolinux/hs110-exporter:latest
```

Notes:

1. the `TARGET_HS110` environment variable is the target plug IP or hostname.
1. if you have multiple plugs you have to increase the host port number (for example `-p 9499:9498` and/or `-p 9500:9498`)

The plan is to provide on docker.io/rolinux/hs110-exporter version for AMD64, ARM64 and ARM but if you need another architecture that is supported by Go and docker buildx then I will try to add it.
