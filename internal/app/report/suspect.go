package report

type SuspectResourceType int

const (
	IP_RANGE SuspectResourceType = iota
	DOMAIN_NAME
	EMAIL_ADDRESS
)
