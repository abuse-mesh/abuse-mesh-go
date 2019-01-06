package config

//NodeConfig is the structural representation of the configuration of the node
type NodeConfig struct {
	//UUID is the ID of this node in the network
	UUID string `mapstructure:"uuid" json:"uuid" validate:"required,uuid"`

	//ListenIP is the ip address on which the node will listen for AbuseMesh protocol connections
	ListenIP string `mapstructure:"listen-ip" json:"listen-ip" validate:"required,ip"`

	//ListenPort is the TCP port on which we will listen for AbuseMesh protocol connections
	ListenPort int `mapstructure:"listen-port" json:"listen-port" validate:"min=1,max=65535"`

	//If true the security features will be disabled
	//NOTE: for development, other nodes should not accept insecure connections
	Insecure bool `mapstructure:"insecure" json:"insecure"`

	//The x509 cert used to create the encrypted connection with this node
	TLSCertFile string `mapstructure:"tls-cert-file" json:"tls-cert-file"`

	//The x509 key used to create the encrypted connection with this node
	TLSKeyFile string `mapstructure:"tls-key-file" json:"tls-key-file"`

	//The Autonomous System Number for which this node claims to handle abuse
	ASN int32 `mapstructure:"asn" json:"asn"`

	ContactDetails ContactDetailsConfig `mapstructure:"contact-details" json:"contact-details"`

	PGPProvider string `mapstructure:"pgp-provider" json:"pgp-provider" validate:"oneof=file gpg"`

	PGPProviders PGPProvidersConfig `mapstructure:"pgp-providers" json:"pgp-providers"`
}

type PGPProvidersConfig struct {
	File FilePGPProvidersConfig `mapstructure:"file" json:"file"`
}

type FilePGPProvidersConfig struct {
	KeyRingFile   string `mapstructure:"key-ring-file" json:"key-ring-file"`
	KeyID         string `mapstructure:"key-id" json:"key-id" validate:"base64"`
	AskPassphrase bool   `mapstructure:"ask-passphrase" json:"ask-passphrase"`
	Passphrase    string `mapstructure:"passphrase" json:"passphrase"`
}

type ContactDetailsConfig struct {
	OrganizationName string                `mapstructure:"organization-name" json:"organization-name"`
	EmailAddress     string                `mapstructure:"email-address" json:"email-address" validate:"email"`
	PhoneNumber      string                `mapstructure:"phone-number" json:"phone-number"`
	PhysicalAddress  string                `mapstructure:"physical-address" json:"physical-address"`
	ContactPersons   []ContactPersonConfig `mapstructure:"contact-persons" json:"contact-persons"`
}

type ContactPersonConfig struct {
	FirstName    string `mapstructure:"first-name" json:"first-name"`
	MiddleName   string `mapstructure:"middle-name" json:"middle-name"`
	LastName     string `mapstructure:"last-name" json:"last-name"`
	JobTitle     string `mapstructure:"job-title" json:"job-title"`
	EmailAddress string `mapstructure:"email-address" json:"email-address" validate:"email"`
	PhoneNumber  string `mapstructure:"phone-number" json:"phone-number"`
}
