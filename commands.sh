ip link set br0 down

ip link delete br0 type bridge

ip link add name br0 type bridge

ip addr add 192.168.0.1/24 dev br0

ip link set name up

ip tuntap add dev tap0 mode tap

ip link set tap0 master br0

ip link set tap0 up

sysctl -w net.ipv4.ip_forward=1


# iptables -F FORWARD
# iptables -t nat -F POSTROUTING
# iptables -t nat -A POSTROUTING -s 192.168.0.0/24 -o eth0 -j MASQUERADE
# iptables -A FORWARD -i eth0 -o br0 -s 192.168.0.0/16 -j ACCEPT
# iptables -A FORWARD -i eth0 -o br0 -s 10.0.0.0/8 -j ACCEPT
# iptables -A FORWARD -i eth0 -o br0 -s 172.16.0.0/12 -j ACCEPT

# iptables -A FORWARD -i eth0 -o br0 -m state --state RELATED,ESTABLISHED -j ACCEPT

# iptables -A FORWARD -i br0 -o eth0 -d 10.0.0.0/8 -j DROP
# iptables -A FORWARD -i br0 -o eth0 -d 172.16.0.0/12 -j DROP
# iptables -A FORWARD -i br0 -o eth0 -d 192.168.0.0/16 -j DROP

# iptables -A FORWARD -i br0 -o eth0 -j ACCEPT