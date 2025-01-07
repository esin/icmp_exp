## ICMP Exponent

### Info
The application allows you to respond to ICMP requests exponentially, constantly increasing the timeout between requests by 2 times until it reaches 1048576 msecs (17 minutes).

### Building
Just `go mod tidy && go build .`

### Manual for launching
Run commands from root:

1. Disable icmp reply from kernel:
   `echo 1 | tee /proc/sys/net/ipv4/icmp_echo_ignore_all`
2. Allow using raw sockets:
   `echo 1 | tee /proc/sys/net/ipv4/ping_group_range`
3. Allow rules in iptables:
   ```
   iptables -I OUTPUT -p icmp --icmp-type echo-reply -j ACCEPT
   iptables -I OUTPUT -p icmp --icmp-type echo-reply -j ACCEPT
   ```
4. Allow run binary from regular user:
   `setcap cap_net_raw+ep icmp_exp`

#### Systemd part
5. Put binary to `/usr/local/bin/`
6. Create `/etc/systemd/system/icmp_exp.service`
7. Put everything from [icmp_exp.service](icmp_exp.service)
8. Change User/Group
9. Reload:
   `systemctl daemon-reload`
10. Run and enable autostart:
   `systemctl start icmp_exp; systemctl enable icmp_exp`
