package jlib

type PackageMetaInfo = GetPackagesResponse

// JavaPackage represents an installed Java package
type JavaPackage struct {
	*PackageMetaInfo
	JavaDir      string // Path to the java installation directory
	JavaExecPath string // Path to the java executable
}
