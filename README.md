# OpenWRT Monitor

This little tool polls openwrt routers for TX and RX data and visualizes it.

It relies on three environment variables:
 ROUTER_USER - the router's user, must have access to do some RPC calls.
 ROUTER_PASSWORD - the router's password for the HTTP interface.
 ROUTER_URL - Something like http://192.168.0.1/

Since the tool doesn't know how big the pipe is it will dynamically adjust the scale. Run speedtest and it will likely
be calibrated.

This tool is neither important nor particularly useful. It lacks tests and probably isn't very robust.


## Screenshot

![Image](webassets/Screenshot%202022-05-25%20at%2016.20.18.png)
