package jlib

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStructToMap(t *testing.T) {
	type testStruct struct {
		A string
		B int
	}
	s := testStruct{A: "test", B: 1}
	result, err := structToMap(&s)
	assert.NoError(t, err)
	assert.Equal(t, "test", result["A"])
	assert.Equal(t, 1, result["B"])
}

func TestDownloadFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	dst := path.Join(t.TempDir(), "test")
	err := os.MkdirAll(dst, os.ModePerm)
	assert.NoError(t, err)
	_, err = DownloadFile("http://212.183.159.230/5MB.zip", dst)
	assert.NoError(t, err)
	assert.FileExists(t, path.Join(dst, "5MB.zip"))
}

func TestGetDiscoApiEndpoints(t *testing.T) {
	result, err := GetDiscoApiEndpoints()
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestGetDistributions(t *testing.T) {
	result, err := GetDistributions()
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestGetDistributionsForGivenVersion(t *testing.T) {
	result, err := GetDistributionsForGivenVersion("8")
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestGetDistribution(t *testing.T) {
	result, err := GetDistribution("zulu")
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestDownloadJavaByID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	dst := path.Join(t.TempDir(), "test")
	err := os.MkdirAll(dst, os.ModePerm)
	assert.NoError(t, err)
	file, err := DownloadJavaByID("e210b8304ddd4b4e8d0a79282f4472fb", dst)
	assert.NoError(t, err)
	assert.FileExists(t, path.Join(dst, file.Name()))
}

func TestGetPackageRedirect(t *testing.T) {
	result, err := GetPackageRedirect("e210b8304ddd4b4e8d0a79282f4472fb")
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestGetMajorVersions(t *testing.T) {
	result, err := GetAllMajorVersions(&GetAllMajorVersionsOptions{
		Maintained: true, // without this option disco API returns wrong JSON that cannot be deserialized
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestGetMajorVersion(t *testing.T) {
	result, err := GetMajorVersion(8)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestGetMajorVersionsNew(t *testing.T) {
	result, err := GetMajorVersionsNew(&GetMajorVersionsNewOptions{
		Maintained: true, // without this option disco API returns wrong JSON that cannot be deserialized
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestGetPackages(t *testing.T) {
	result, err := GetPackages(&GetPackagesOptions{
		JDKVersion:      8,
		Distribution:    []string{"zulu"},
		OperatingStatus: []string{"linux"},
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestGetAllPackages(t *testing.T) {
	result, err := GetAllPackages(&GetAllPackagesOptions{
		IncludeEA:    false,
		Downloadable: true,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestGetAllPackagesGraalVM(t *testing.T) {
	result, err := GetAllPackagesGraalVM(&GetAllPackagesOptions{
		IncludeEA:    false,
		Downloadable: true,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestGetAllPackagesOpenJDK(t *testing.T) {
	result, err := GetAllPackagesOpenJDK(&GetAllPackagesOptions{
		IncludeEA:    false,
		Downloadable: true,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestGetJDKPackages(t *testing.T) {
	result, err := GetJDKPackages(&GetPackagesOptions{
		JDKVersion:      8,
		Architecture:    []string{"amd64", "x86_64"},
		OperatingSystem: []string{"linux"},
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestJREPackages(t *testing.T) {
	result, err := GetJREPackages(&GetPackagesOptions{
		JDKVersion:      8,
		Distribution:    []string{"zulu"},
		OperatingSystem: []string{"linux", "windows"},
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestGetPackage(t *testing.T) {
	result, err := GetPackage("e210b8304ddd4b4e8d0a79282f4472fb")
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestGetParameters(t *testing.T) {
	result, err := GetParameters()
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestGetRemainingDaysToNextRelease(t *testing.T) {
	result, err := GetRemainingDaysToNextRelease()
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.DateOfNextRelease)
}

func TestGetRemainingDaysToNextUpdate(t *testing.T) {
	result, err := GetRemainingDaysToNextUpdate()
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.DateOfNextUpdate)
}

func TestGetSupportedArchitectures(t *testing.T) {
	result, err := GetSupportedArchitectures()
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestGetSupportedArchiveTypes(t *testing.T) {
	result, err := GetSupportedArchiveTypes()
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestGetSupportedFeatures(t *testing.T) {
	result, err := GetSupportedFeatures()
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestGetSupportedFPUs(t *testing.T) {
	result, err := GetSupportedFPUs()
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestGetSupportedLatestParameters(t *testing.T) {
	result, err := GetSupportedLatestParameters()
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestGetSupportedLibCTypes(t *testing.T) {
	result, err := GetSupportedLibCTypes()
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestGetSupportedOperatingSystems(t *testing.T) {
	result, err := GetSupportedOperatingSystems()
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestGetSupportedPackageTypes(t *testing.T) {
	result, err := GetSupportedPackageTypes()
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestGetSupportedReleaseTypes(t *testing.T) {
	result, err := GetSupportedReleaseStatus()
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestGetSupportedTermsOfSupport(t *testing.T) {
	result, err := GetSupportedTermsOfSupport()
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}
