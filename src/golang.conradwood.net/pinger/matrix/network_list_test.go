package matrix

import (
	"testing"
)

func TestLookupPrivate(t *testing.T) {
	test_lookup_private(t, "192.168.59.1/24", "ASN_LOCAL_192.168.59.0/24")
	test_lookup_private(t, "192.168.59.1/32", "ASN_LOCAL_192.168.59.0/24")
	test_lookup_private(t, "192.168.59.1", "ASN_LOCAL_192.168.59.0/24")
	test_lookup_private(t, "192.168.59.1/16", "ASN_LOCAL_192.168.0.0/16")
}
func test_lookup_private(t *testing.T, ip, exp string) {
	ni := lookup_private_net_info(ip)
	if ni.asn != exp {
		t.Errorf("for ip \"%s\" expected \"%s\", but got \"%s\"\n", ip, exp, ni.asn)
	}
}
