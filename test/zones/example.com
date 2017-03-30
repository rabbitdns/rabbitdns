$ORIGIN example.com.
$TTL 3600
@ IN SOA     z.example.com. root.example.com. 2 3600 900 1814400 900
  IN NS ns1.example.com.
  IN NS ns2.example.com.
ns1 IN A 192.168.0.1
    IN AAAA 2001:db8::1
ns1 IN A 192.168.0.2
    IN AAAA 2001:db8::2
www IN A 192.168.0.2
www2 IN CNAME www
www3 IN CNAME www2
www4 IN CNAME www3
www5 IN CNAME www4
www6 IN CNAME www5
www7 IN CNAME www6
www8 IN CNAME www7
www9 IN CNAME www8
www10 IN CNAME www9
www11 IN CNAME www10
www12 IN CNAME www11
www13 IN CNAME www12
www14 IN CNAME www13
www15 IN CNAME www14
www16 IN CNAME www15
www17 IN CNAME www16
*.apple IN A 192.168.255.1
sub IN NS ns1.sub
ns1.sub IN A 192.168.0.3
