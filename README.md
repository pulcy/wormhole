# Wormhole: TCP Service Proxy

Wormhole is a TCP proxy for micro service.
It is configured using data in ETCD coming from [Registrator](https://github.com/gliderlabs/registrator).

Internally, workhole uses [haproxy](http://www.haproxy.org/) to do the actual proxying,
where the configuration for haproxy is created by wormhole.
