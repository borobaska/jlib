package jlib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/mitchellh/mapstructure"
)

const DISCO_API_V3_BASE_URL = "https://api.foojay.io/disco/v3.0"

type DiscoResponseWrapper[T any] struct {
	Result T `json:"result"`
}

type DiscoApiEndpoint struct {
	Uri string `json:"uri"`
}

func extractOptions[T any](options []*T) *T {
	if len(options) > 0 && options[0] != nil {
		return options[0]
	}
	return nil
}

func structToMap[T any](s T) (data map[string]interface{}, err error) {
	return data, mapstructure.Decode(s, &data)
}

func getAndParseResponseWithQuery[TResponse any](base string, query map[string]interface{}, path ...string) (TResponse, error) {
	u, err := url.JoinPath(base, path...)
	if err != nil {
		return *new(TResponse), err
	}
	q := url.Values{}

	for k, v := range query {
		// Make it possible to pass arrays as query parameters
		switch v := v.(type) {
		case []string:
			q.Add(k, strings.Join(v, ","))
		case []int:
			s := make([]string, len(v))
			for i, n := range v {
				s[i] = fmt.Sprintf("%v", n)
			}
			q.Add(k, strings.Join(s, ","))
		default:
			q.Set(k, fmt.Sprintf("%v", v))
		}
	}

	if len(q) > 0 {
		u += "?" + q.Encode()
	}

	resp, err := http.Get(u)
	if err != nil {
		return *new(TResponse), err
	}
	defer resp.Body.Close()

	var wrapper DiscoResponseWrapper[TResponse]
	err = json.NewDecoder(resp.Body).Decode(&wrapper)
	if err != nil {
		return *new(TResponse), err
	}

	return wrapper.Result, nil
}

func getAndParseResponse[TResponse any](base string, path ...string) (TResponse, error) {
	return getAndParseResponseWithQuery[TResponse](base, map[string]interface{}{}, path...)
}

func GetDiscoApiEndpoints() ([]DiscoApiEndpoint, error) {
	r, err := getAndParseResponse[[]DiscoApiEndpoint](DISCO_API_V3_BASE_URL)
	return r, err
}

type DistributionsOptions struct {
	IncludeVersions  bool     `mapstructure:"include_versions,omitempty"`
	IncludeSynonyms  bool     `mapstructure:"include_synonyms,omitempty"`
	DiscoveryScopeId []string `mapstructure:"discovery_scope_id,omitempty"`
}

type DistributionsResponse struct {
	Name           string   `json:"name"`
	ApiParameter   string   `json:"api_parameter"`
	Maintained     bool     `json:"maintained"`
	Available      bool     `json:"available"`
	BuildOfOpenJDK bool     `json:"build_of_openjdk"`
	BuildOfGraalVM bool     `json:"build_of_graalvm"`
	OfficialUri    string   `json:"official_uri"`
	Synonyms       []string `json:"synonyms"`
	Versions       []string `json:"versions"`
}

// Returns a list of all supported distributions
func GetDistributions(options ...*DistributionsOptions) ([]DistributionsResponse, error) {
	query, err := structToMap(extractOptions(options))
	if err != nil {
		return nil, err
	}

	r, err := getAndParseResponseWithQuery[[]DistributionsResponse](DISCO_API_V3_BASE_URL, query, "distributions")
	return r, err
}

type DistributionsForGivenVersionOptions struct {
	DiscoveryScopeId []string `mapstructure:"discovery_scope_id,omitempty"`
	Match            string   `mapstructure:"match,omitempty"`
	IncludeVersions  bool     `mapstructure:"include_versions,omitempty"`
	IncludeSynonyms  bool     `mapstructure:"include_synonyms,omitempty"`
}

// Returns a list of all distributions that support the given Java version
func GetDistributionsForGivenVersion(version string, options ...*DistributionsForGivenVersionOptions) ([]DistributionsResponse, error) {
	query, err := structToMap(extractOptions(options))
	if err != nil {
		return nil, err
	}

	r, err := getAndParseResponseWithQuery[[]DistributionsResponse](DISCO_API_V3_BASE_URL, query, "distributions", "versions", version)
	return r, err
}

type GetDistributionOptions struct {
	LatestPerUpdate  bool     `mapstructure:"latest_per_update,omitempty"`
	DiscoveryScopeId []string `mapstructure:"discovery_scope_id,omitempty"`
	Match            string   `mapstructure:"match,omitempty"`
	IncludeVersions  bool     `mapstructure:"include_versions,omitempty"`
	IncludeSynonyms  bool     `mapstructure:"include_synonyms,omitempty"`
	IncludeEA        bool     `mapstructure:"include_ea,omitempty"`
}

// Returns detailled information about a given distribution
func GetDistribution(distribution string, options ...*GetDistributionOptions) ([]DistributionsResponse, error) {
	query, err := structToMap(extractOptions(options))
	if err != nil {
		return nil, err
	}

	r, err := getAndParseResponseWithQuery[[]DistributionsResponse](DISCO_API_V3_BASE_URL, query, "distributions", distribution)
	return r, err
}

// Redirects to either the direct download link or the download site of the requested package defined by it's id
func GetPackageRedirect(id string) (string, error) {
	u, err := url.JoinPath(DISCO_API_V3_BASE_URL, "ids", id, "redirect")
	if err != nil {
		return "", err
	}

	resp, err := http.Get(u)

	return resp.Request.URL.String(), err
}

func GetFilename(id string) (string, error) {
	finalURL, err := GetPackageRedirect(id)
	if err != nil {
		return "", err
	}
	return path.Base(finalURL), nil
}

// DownloadJavaByID downloads Java archive by its ID to dest directory and returns the filename
func DownloadJavaByID(id string, dst string) (*os.File, error) {
	javaUrl, err := GetPackageRedirect(id)
	if err != nil {
		return nil, err
	}
	return DownloadFile(javaUrl, dst)
}

type GetAllMajorVersionsOptions struct {
	EA               bool     `mapstructure:"ea,omitempty"`
	GA               bool     `mapstructure:"ga,omitempty"`
	Maintained       bool     `mapstructure:"maintained,omitempty"`
	IncludeBuild     bool     `mapstructure:"include_build,omitempty"`
	DiscoveryScopeId []string `mapstructure:"discovery_scope_id,omitempty"`
	Match            string   `mapstructure:"match,omitempty"`
	IncludeVersions  bool     `mapstructure:"include_versions,omitempty"`
}

type GetAllMajorVersionsResponse struct {
	MajorVersion    int      `json:"major_version"`
	TermOfSupport   string   `json:"term_of_support"`
	Maintained      bool     `json:"maintained"`
	EarlyAccessOnly bool     `json:"early_access_only"`
	ReleaseStatus   string   `json:"release_status"`
	Versions        []string `json:"versions"`
}

// Return a list of major versions defined by the given parameters
func GetAllMajorVersions(options ...*GetAllMajorVersionsOptions) ([]GetAllMajorVersionsResponse, error) {
	query, err := structToMap(extractOptions(options))
	if err != nil {
		return nil, err
	}

	r, err := getAndParseResponseWithQuery[[]GetAllMajorVersionsResponse](DISCO_API_V3_BASE_URL, query, "major_versions")
	return r, err
}

type BuildAndVersionOptions struct {
	IncludeBuild    bool `json:"include_build"`
	IncludeVersions bool `json:"include_versions"`
}

// Returns the specified major version including early access builds
func GetSpecificMajorVersionIncludingEA(version int, options ...*BuildAndVersionOptions) ([]GetAllMajorVersionsResponse, error) {
	query, err := structToMap(extractOptions(options))
	if err != nil {
		return nil, err
	}

	return getAndParseResponseWithQuery[[]GetAllMajorVersionsResponse](DISCO_API_V3_BASE_URL, query, "major_versions", fmt.Sprintf("%v", version), "ea")
}

// Returns the specified major version excluding early access builds
func GetSpecificMajorVersion(version int, options ...*BuildAndVersionOptions) ([]GetAllMajorVersionsResponse, error) {
	opt := extractOptions(options)
	query, err := structToMap(opt)
	if err != nil {
		return nil, err
	}

	return getAndParseResponseWithQuery[[]GetAllMajorVersionsResponse](DISCO_API_V3_BASE_URL, query, "major_versions", fmt.Sprintf("%v", version), "ga")
}

// Returns information about the requested major version
func GetMajorVersion(version int, options ...*BuildAndVersionOptions) ([]GetAllMajorVersionsResponse, error) {
	query, err := structToMap(extractOptions(options))
	if err != nil {
		return nil, err
	}

	return getAndParseResponseWithQuery[[]GetAllMajorVersionsResponse](DISCO_API_V3_BASE_URL, query, "major_versions", fmt.Sprintf("%v", version))
}

type GetMajorVersionsNewOptions struct {
	ReleaseStatus         string `mapstructure:"release_status,omitempty"`
	ReleaseStatusVersions string `mapstructure:"release_status_versions,omitempty"`
	Maintained            bool   `mapstructure:"maintained,omitempty"`
	IncludeBuild          bool   `mapstructure:"include_build,omitempty"`
	IncludeVersions       bool   `mapstructure:"include_versions,omitempty"`
	LTSOnly               bool   `mapstructure:"lts_only,omitempty"`
}

type GetMajorVersionsNewResponse struct {
	MajorVersion    int    `json:"major_version"`
	TermOfSupport   string `json:"term_of_support"`
	Maintained      bool   `json:"maintained"`
	EarlyAccessOnly bool   `json:"early_access_only"`
	ReleaseStatus   string `json:"release_status"`
}

// Return a list of major versions defined by the given parameters
func GetMajorVersionsNew(options ...*GetMajorVersionsNewOptions) ([]GetMajorVersionsNewResponse, error) {
	query, err := structToMap(extractOptions(options))
	if err != nil {
		return nil, err
	}

	return getAndParseResponseWithQuery[[]GetMajorVersionsNewResponse](DISCO_API_V3_BASE_URL, query, "major_versions")
}

type GetPackagesResponseFeature = GetSupportedArchiveTypesResponse

type GetPackagesOptions struct {
	Version               string   `mapstructure:"version,omitempty"`
	VersionByDefinition   string   `mapstructure:"version_by_definition,omitempty"`
	JDKVersion            int      `mapstructure:"jdk_version,omitempty"`
	Distro                []string `mapstructure:"distro,omitempty"`
	Distribution          []string `mapstructure:"distribution,omitempty"`
	Architecture          []string `mapstructure:"architecture,omitempty"`
	ArchiveType           []string `mapstructure:"archive_type,omitempty"`
	OperatingSystem       []string `mapstructure:"operating_system,omitempty"`
	PackageType           string   `mapstructure:"package_type,omitempty"`
	OperatingStatus       []string `mapstructure:"operating_status,omitempty"`
	LibcType              []string `mapstructure:"libc_type,omitempty"`
	LibCType              []string `mapstructure:"lib_c_type,omitempty"`
	ReleaseStatus         []string `mapstructure:"release_status,omitempty"`
	TermOfSupport         []string `mapstructure:"term_of_support,omitempty"`
	Bitness               int      `mapstructure:"bitness,omitempty"`
	FPU                   []string `mapstructure:"fpu,omitempty"`
	JavaFXBundled         bool     `mapstructure:"javafx_bundled,omitempty"`
	WithJavaFXAvailable   bool     `mapstructure:"with_javafx_available,omitempty"`
	DirectlyDownloadable  bool     `mapstructure:"directly_downloadable,omitempty"`
	Latest                string   `mapstructure:"latest,omitempty"`
	Feature               []string `mapstructure:"feature,omitempty"`
	SignatureAvailable    bool     `mapstructure:"signature_available,omitempty"`
	FreeToUseInProduction bool     `mapstructure:"free_to_use_in_production,omitempty"`
	TCKTested             string   `mapstructure:"tck_tested,omitempty"`
	AqavitCertified       string   `mapstructure:"aqavit_certified,omitempty"`
	DiscoveryScopeId      []string `mapstructure:"discovery_scope_id,omitempty"`
	Match                 string   `mapstructure:"match,omitempty"`
}

type GetPackagesResponse struct {
	ID                   string `json:"id"`
	ArchiveType          string `json:"archive_type"`
	Distribution         string `json:"distribution"`
	MajorVersion         int    `json:"major_version"`
	JavaVersion          string `json:"java_version"`
	DistributionVersion  string `json:"distribution_version"`
	JDKVersion           int    `json:"jdk_version"`
	LatestBuildAvailable bool   `json:"latest_build_available"`
	ReleaseStatus        string `json:"release_status"`
	TermOfSupport        string `json:"term_of_support"`
	OperatingSystem      string `json:"operating_system"`
	LibCType             string `json:"lib_c_type"`
	Architecture         string `json:"architecture"`
	FPU                  string `json:"fpu"`
	PackageType          string `json:"package_type"`
	JavaFXBundled        bool   `json:"javafx_bundled"`
	DirectlyDownloadable bool   `json:"directly_downloadable"`
	Filename             string `json:"filename"`
	Links                struct {
		PkgInfoURI          string `json:"pkg_info_uri"`
		PkgDownloadRedirect string `json:"pkg_download_redirect"`
	} `json:"links"`
	FreeUseInProduction bool                         `json:"free_use_in_production"`
	TCKTested           string                       `json:"tck_tested"`
	TCKCertURI          string                       `json:"tck_cert_uri"`
	AqavitCertified     string                       `json:"aqavit_certified"`
	AqavitCertURI       string                       `json:"aqavit_cert_uri"`
	Size                int                          `json:"size"`
	Feature             []GetPackagesResponseFeature `json:"feature"`
}

// Returns a list of packages defined by the given parameters.
// The version parameter not only supports different formats for version numbers (e.g. 11.9.0.1, 1.8.0_262, 15, 16-ea) but also ranges (e.g. 15.0.1..<16). The ranges are defined as follows: VersionNumber1...VersionNumber2 => includes VersionNumber1 and VersionNumber2 VersionNumber1.. includes VersionNumber1 and excludes VersionNumber2 VersionNumber1>..VersionNUmber2 => excludes VersionNumber1 and includes VersionNumber2 VersionNumber1>. excludes VersionNumber1 and VersionNumber2
func GetPackages(options ...*GetPackagesOptions) ([]GetPackagesResponse, error) {
	query, err := structToMap(extractOptions(options))
	if err != nil {
		return nil, err
	}

	return getAndParseResponseWithQuery[[]GetPackagesResponse](DISCO_API_V3_BASE_URL, query, "packages")
}

type GetAllPackagesOptions struct {
	Downloadable bool `mapstructure:"downloadable,omitempty"`
	IncludeEA    bool `mapstructure:"include_ea,omitempty"`
}

// Returns all packages defined the downloadable and include_ea parameter
func GetAllPackages(options ...*GetAllPackagesOptions) ([]GetPackagesResponse, error) {
	query, err := structToMap(extractOptions(options))
	if err != nil {
		return nil, err
	}

	return getAndParseResponseWithQuery[[]GetPackagesResponse](DISCO_API_V3_BASE_URL, query, "packages")
}

// Returns all packages that are builds of GraalVM
func GetAllPackagesGraalVM(options ...*GetAllPackagesOptions) ([]GetPackagesResponse, error) {
	query, err := structToMap(extractOptions(options))
	if err != nil {
		return nil, err
	}

	return getAndParseResponseWithQuery[[]GetPackagesResponse](DISCO_API_V3_BASE_URL, query, "packages", "all_builds_of_graalvm")
}

type AllPackagesOpenJDKOptions struct {
	GetAllPackagesOptions
	Minimized bool `mapstructure:"minimized"`
}

// Returns all packages that are builds of OpenJDK
func GetAllPackagesOpenJDK(options ...*GetAllPackagesOptions) ([]GetPackagesResponse, error) {
	query, err := structToMap(extractOptions(options))
	if err != nil {
		return nil, err
	}

	return getAndParseResponseWithQuery[[]GetPackagesResponse](DISCO_API_V3_BASE_URL, query, "packages", "all_builds_of_openjdk")
}

// Returns a list of packages that are of package_type JDK defined by the given parameters. The version parameter not only supports different formats for version numbers (e.g. 11.9.0.1, 1.8.0_262, 15, 16-ea) but also ranges (e.g. 15.0.1..<16). The ranges are defined as follows: VersionNumber1...VersionNumber2 => includes VersionNumber1 and VersionNumber2 VersionNumber1.. includes VersionNumber1 and excludes VersionNumber2 VersionNumber1>..VersionNUmber2 => excludes VersionNumber1 and includes VersionNumber2 VersionNumber1>. excludes VersionNumber1 and VersionNumber2
func GetJDKPackages(options ...*GetPackagesOptions) ([]GetPackagesResponse, error) {
	query, err := structToMap(extractOptions(options))
	if err != nil {
		return nil, err
	}

	return getAndParseResponseWithQuery[[]GetPackagesResponse](DISCO_API_V3_BASE_URL, query, "packages", "jdks")
}

// Returns a list of packages that are of package_type JRE defined by the given parameters. The version parameter not only supports different formats for version numbers (e.g. 11.9.0.1, 1.8.0_262, 15, 16-ea) but also ranges (e.g. 15.0.1..<16). The ranges are defined as follows: VersionNumber1...VersionNumber2 => includes VersionNumber1 and VersionNumber2 VersionNumber1.. includes VersionNumber1 and excludes VersionNumber2 VersionNumber1>..VersionNUmber2 => excludes VersionNumber1 and includes VersionNumber2 VersionNumber1>. excludes VersionNumber1 and VersionNumber2
func GetJREPackages(options ...*GetPackagesOptions) ([]GetPackagesResponse, error) {
	query, err := structToMap(extractOptions(options))
	if err != nil {
		return nil, err
	}

	return getAndParseResponseWithQuery[[]GetPackagesResponse](DISCO_API_V3_BASE_URL, query, "packages", "jres")
}

// Returns information about a package defined by the given package id
func GetPackage(id string) (GetPackagesResponse, error) {
	res, err := getAndParseResponse[[]GetPackagesResponse](DISCO_API_V3_BASE_URL, "packages", id)
	if err != nil {
		return GetPackagesResponse{}, err
	}
	if len(res) == 0 {
		return GetPackagesResponse{}, fmt.Errorf("no package found for id %v", id)
	}
	return res[0], err
}

type ParametersV3 struct {
	Packages struct {
		Architecture          string `json:"architecture"`
		ArchiveType           string `json:"archive_type"`
		Bitness               string `json:"bitness"`
		FPU                   string `json:"fpu"`
		DirectlyDownloadable  string `json:"directly_downloadable"`
		Distro                string `json:"distro"`
		Feature               string `json:"feature"`
		JavaFXBundled         string `json:"javafx_bundled"`
		WithJavaFXIfAvailable string `json:"with_javafx_if_available"`
		Latest                string `json:"latest"`
		LibCType              string `json:"lib_c_type"`
		MajorVersion          string `json:"major_version"`
		OperatingSystem       string `json:"operating_system"`
		PackageType           string `json:"package_type"`
		ReleaseStatus         string `json:"release_status"`
		TermOfSupport         string `json:"term_of_support"`
		FreeUseInProduction   string `json:"free_use_in_production"`
		ChecksumType          string `json:"checksum_type"`
		Version               string `json:"version"`
	} `json:"packages"`

	MajorVersions struct {
		EA         string `json:"ea"`
		Maintained string `json:"maintained"`
	} `json:"major_versions"`

	Distributions struct {
		DiscoveryScopeId string `json:"discovery_scope_id"`
	} `json:"distributions"`

	Ids struct {
		Token string `json:"token"`
	} `json:"ids"`
}

func GetParameters() (*ParametersV3, error) {
	p, err := getAndParseResponse[[]ParametersV3](DISCO_API_V3_BASE_URL, "parameters")
	if err != nil {
		return nil, err
	}
	if len(p) == 0 {
		return nil, fmt.Errorf("no parameters found")
	}
	return &p[0], err
}

type RemainingDaysToNextReleaseResponse struct {
	DaysToNextRelease int    `json:"days_to_next_release"`
	DateOfNextRelease string `json:"date_of_next_release"`
}

// Returns the remaining days to next feature release (e.g. 21 GA) based on the current release cadence
func GetRemainingDaysToNextRelease() (*RemainingDaysToNextReleaseResponse, error) {
	r, err := getAndParseResponse[[]RemainingDaysToNextReleaseResponse](DISCO_API_V3_BASE_URL, "remaining_days", "release")
	if err != nil {
		return nil, err
	}
	if len(r) == 0 {
		return nil, fmt.Errorf("no data found")
	}
	return &r[0], err
}

type GetRemainingDaysToNextUpdateReponse struct {
	DaysToNextUpdate int    `json:"days_to_next_update"`
	DateOfNextUpdate string `json:"date_of_next_update"`
}

func GetRemainingDaysToNextUpdate() (*GetRemainingDaysToNextUpdateReponse, error) {
	r, err := getAndParseResponse[[]GetRemainingDaysToNextUpdateReponse](DISCO_API_V3_BASE_URL, "remaining_days", "update")
	if err != nil {
		return nil, err
	}
	if len(r) == 0 {
		return nil, fmt.Errorf("no data found")
	}
	return &r[0], err
}

type GetSupportedArchitecturesResponse struct {
	Name      string `json:"name"`
	UiString  string `json:"ui_string"`
	ApiString string `json:"api_string"`
	Bitness   string `json:"bitness"`
}

func GetSupportedArchitectures() ([]GetSupportedArchitecturesResponse, error) {
	return getAndParseResponse[[]GetSupportedArchitecturesResponse](DISCO_API_V3_BASE_URL, "supported_architectures")
}

type GetSupportedArchiveTypesResponse struct {
	Name      string `json:"name"`
	UiString  string `json:"ui_string"`
	ApiString string `json:"api_string"`
}

func GetSupportedArchiveTypes() ([]GetSupportedArchiveTypesResponse, error) {
	return getAndParseResponse[[]GetSupportedArchiveTypesResponse](DISCO_API_V3_BASE_URL, "supported_archive_types")
}

type GetSupportedFeaturesResponse = GetSupportedArchiveTypesResponse

func GetSupportedFeatures() ([]GetSupportedFeaturesResponse, error) {
	return getAndParseResponse[[]GetSupportedFeaturesResponse](DISCO_API_V3_BASE_URL, "supported_features")
}

type GetSupportedFPUsResponse = GetSupportedArchiveTypesResponse

func GetSupportedFPUs() ([]GetSupportedFPUsResponse, error) {
	return getAndParseResponse[[]GetSupportedFPUsResponse](DISCO_API_V3_BASE_URL, "supported_fpus")
}

type GetSupportedLatestParametersResponse = GetSupportedArchiveTypesResponse

func GetSupportedLatestParameters() ([]GetSupportedLatestParametersResponse, error) {
	return getAndParseResponse[[]GetSupportedLatestParametersResponse](DISCO_API_V3_BASE_URL, "supported_latest_parameters")
}

type GetSupportedLibCTypesResponse = GetSupportedArchiveTypesResponse

func GetSupportedLibCTypes() ([]GetSupportedLibCTypesResponse, error) {
	return getAndParseResponse[[]GetSupportedLibCTypesResponse](DISCO_API_V3_BASE_URL, "supported_lib_c_types")
}

type GetSupportedOperatingSystemsResponse struct {
	Name      string `json:"name"`
	UiString  string `json:"ui_string"`
	ApiString string `json:"api_string"`
	LibCType  string `json:"lib_c_type"`
}

func GetSupportedOperatingSystems() ([]GetSupportedOperatingSystemsResponse, error) {
	return getAndParseResponse[[]GetSupportedOperatingSystemsResponse](DISCO_API_V3_BASE_URL, "supported_operating_systems")
}

type GetSupportedPackageTypesResponse = GetSupportedArchiveTypesResponse

func GetSupportedPackageTypes() ([]GetSupportedPackageTypesResponse, error) {
	return getAndParseResponse[[]GetSupportedPackageTypesResponse](DISCO_API_V3_BASE_URL, "supported_package_types")
}

type GetSupportedReleaseStatusResponse = GetSupportedArchiveTypesResponse

func GetSupportedReleaseStatus() ([]GetSupportedReleaseStatusResponse, error) {
	return getAndParseResponse[[]GetSupportedReleaseStatusResponse](DISCO_API_V3_BASE_URL, "supported_release_status")
}

type GetSupportedTermsOfSupportResponse = GetSupportedArchiveTypesResponse

func GetSupportedTermsOfSupport() ([]GetSupportedTermsOfSupportResponse, error) {
	return getAndParseResponse[[]GetSupportedTermsOfSupportResponse](DISCO_API_V3_BASE_URL, "supported_terms_of_support")
}
