package udpproxy

import wsserver "github.com/devdevaraj/firestarter/ws_server"

var udp_proxy_cache = map[string]*UDPProxy_t{}

var target_cache = wsserver.TARGET_CACHE

func GetUDPProxy(id string) (*UDPProxy_t, bool) {
	target, ok := target_cache[id]
	if !ok || target == "" {
		return nil, false
	}
	cached, ok := udp_proxy_cache[id]
	if ok {
		return cached, true
	}
	newProxy := NewUDPProxy(target)
	udp_proxy_cache[id] = newProxy
	return newProxy, true
}
