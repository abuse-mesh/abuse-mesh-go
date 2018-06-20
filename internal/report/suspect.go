package report

type SuspectResourceType int

const (
	IPRange SuspectResourceType = iota
	DomainName
	EmailAddress
)
