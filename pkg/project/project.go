package project

var (
	bundleVersion = "0.0.1"
	description   = "The dns-network-policy-operator does something."
	gitSHA        = "n/a"
	name          = "dns-network-policy-operator"
	source        = "https://github.com/giantswarm/dns-network-policy-operator"
	version       = "n/a"
)

func BundleVersion() string {
	return bundleVersion
}

func Description() string {
	return description
}

func GitSHA() string {
	return gitSHA
}

func Name() string {
	return name
}

func Source() string {
	return source
}

func Version() string {
	return version
}
