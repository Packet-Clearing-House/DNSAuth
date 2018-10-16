package dnsdist

type RCode uint32

const (
	DNS_RCODE_NOERROR  RCode = 0
	DNS_RCODE_FORMERR        = 1
	DNS_RCODE_SERVFAIL       = 2
	DNS_RCODE_NXDOMAIN       = 3
	DNS_RCODE_NOTIMP         = 4
	DNS_RCODE_REFUSED        = 5
)

const DNS_RCODE_OTHER = "other"

var RCodeStringMap = map[RCode]string{
	DNS_RCODE_NOERROR:  "noerror",
	DNS_RCODE_FORMERR:  "formerr",
	DNS_RCODE_SERVFAIL: "servfail",
	DNS_RCODE_NXDOMAIN: "nxdomain",
	DNS_RCODE_NOTIMP:   "notimp",
	DNS_RCODE_REFUSED:  "refused",
}

func (r RCode) String() string {
	if v, ok := RCodeStringMap[r]; ok {
		return v
	}
	return DNS_RCODE_OTHER
}

type QType uint32

const (
	DNS_QTYPE_A      QType = 1
	DNS_QTYPE_NS           = 2
	DNS_QTYPE_CNAME        = 5
	DNS_QTYPE_SOA          = 6
	DNS_QTYPE_PTR          = 12
	DNS_QTYPE_MX           = 15
	DNS_QTYPE_TXT          = 16
	DNS_QTYPE_AAAA         = 28
	DNS_QTYPE_SRV          = 33
	DNS_QTYPE_NAPTR        = 35
	DNS_QTYPE_DS           = 43
	DNS_QTYPE_NSEC         = 47
	DNS_QTYPE_DNSKEY       = 48
	DNS_QTYPE_NSEC3        = 50
	DNS_QTYPE_SPF          = 99
)

const DNS_QTYPE_OTHER = "OTHER"

var QTypeStringMap = map[QType]string{
	DNS_QTYPE_A:      "A",
	DNS_QTYPE_NS:     "NS",
	DNS_QTYPE_CNAME:  "CNAME",
	DNS_QTYPE_SOA:    "SOA",
	DNS_QTYPE_PTR:    "PTR",
	DNS_QTYPE_MX:     "MX",
	DNS_QTYPE_TXT:    "TXT",
	DNS_QTYPE_AAAA:   "AAAA",
	DNS_QTYPE_SRV:    "SRV",
	DNS_QTYPE_NAPTR:  "NAPTR",
	DNS_QTYPE_DS:     "DS",
	DNS_QTYPE_NSEC:   "NSEC",
	DNS_QTYPE_DNSKEY: "DNSKEY",
	DNS_QTYPE_NSEC3:  "NSEC3",
	DNS_QTYPE_SPF:    "SPF",
}

func (q QType) String() string {
	if v, ok := QTypeStringMap[q]; ok {
		return v
	}
	return DNS_QTYPE_OTHER
}
