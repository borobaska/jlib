package jlib

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
)

type VersionManager struct {
	DataDir string // Path where JLib stores the data
}

func NewVersionManager(dataDir string) *VersionManager {
	return &VersionManager{DataDir: dataDir}
}

func NewDefaultVersionManager() (*VersionManager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return NewVersionManager(path.Join(home, ".jlib")), nil
}

type JavaInstallOptions = GetPackagesOptions

func IsInstalled(err error) bool {
	return err == ErrPackageAlreadyInstalled
}

func IsNotInstalled(err error) bool {
	return !IsInstalled(err)
}

var ErrPackageAlreadyInstalled = fmt.Errorf("package already installed")

func (vm *VersionManager) Install(options *JavaInstallOptions) (*JavaPackage, error) {
	options.ArchiveType = []string{"zip"}

	packages, err := GetPackages(options)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch packages: %w", err)
	}
	if len(packages) == 0 {
		return nil, fmt.Errorf("no packages found")
	}

	dirname, err := GetFilename(packages[0].ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get filename: %w", err)
	}

	dirname = strings.TrimSuffix(dirname, ".zip")

	metapath := path.Join(vm.DataDir, dirname, "meta.json")
	_, err = os.Stat(metapath)
	if err == nil {
		meta, err := readStructFromJSONFile[PackageMetaInfo](metapath)
		if err != nil {
			return nil, fmt.Errorf("failed to read package metadata: %w", err)
		}

		return &JavaPackage{
			PackageMetaInfo: meta,
			JavaDir:         path.Join(vm.DataDir, dirname),
			JavaExecPath:    path.Join(vm.DataDir, dirname, "bin", addExeIfWindows("java")),
		}, ErrPackageAlreadyInstalled
	}

	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to check if package is installed: %w", err)
	}

	tmp := os.TempDir()

	file, err := DownloadJavaByID(packages[0].ID, tmp)
	if err != nil {
		return nil, fmt.Errorf("failed to download package: %w", err)
	}

	if err := unzip(file.Name(), vm.DataDir); err != nil {
		return nil, fmt.Errorf("failed to unzip package: %w", err)
	}

	err = saveStructToJSONFile(packages[0], path.Join(vm.DataDir, dirname, "meta.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to save package metadata: %w", err)
	}

	return &JavaPackage{
		PackageMetaInfo: &packages[0],
		JavaDir:         path.Join(vm.DataDir, dirname),
		JavaExecPath:    path.Join(vm.DataDir, dirname, "bin", addExeIfWindows("java")),
	}, nil
}

func (vm *VersionManager) List() ([]*JavaPackage, error) {
	files, err := os.ReadDir(vm.DataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var javas []*JavaPackage
	for _, file := range files {
		if file.IsDir() {
			meta, err := readStructFromJSONFile[PackageMetaInfo](path.Join(vm.DataDir, file.Name(), "meta.json"))
			if err != nil {
				return nil, fmt.Errorf("failed to read package metadata: %w", err)
			}
			javas = append(javas, &JavaPackage{
				PackageMetaInfo: meta,
				JavaDir:         path.Join(vm.DataDir, file.Name()),
				JavaExecPath:    path.Join(vm.DataDir, file.Name(), "bin", addExeIfWindows("java")),
			})
		}
	}

	return javas, nil
}

func (vm *VersionManager) Remove(java *JavaPackage) error {
	if err := os.RemoveAll(java.JavaDir); err != nil {
		return fmt.Errorf("failed to remove directory: %w", err)
	}
	return nil
}

func (vm *VersionManager) GetJavaByID(id string) (*JavaPackage, error) {
	javas, err := vm.List()
	if err != nil {
		return nil, err
	}

	for _, java := range javas {
		if java.ID == id {
			return java, nil
		}
	}

	return nil, fmt.Errorf("package not found")
}

var ErrJavaNotFound = fmt.Errorf("java not found")

func (vm *VersionManager) Use(distribution string, jdkVersion int) (*JavaPackage, error) {
	javas, err := vm.List()
	if err != nil {
		return nil, err
	}

	for _, java := range javas {
		if java.Distribution == distribution && java.JDKVersion == jdkVersion {
			return java, nil
		}
	}
	return nil, ErrJavaNotFound
}

// Get the Java version or install it if it doesn't exist
func (vm *VersionManager) UseOrInstall(distribution string, jdkVersion int) (*JavaPackage, error) {
	java, err := vm.Use(distribution, jdkVersion)
	if errors.Is(err, ErrJavaNotFound) {
		return vm.Install(&JavaInstallOptions{
			Distribution:    []string{distribution},
			JDKVersion:      jdkVersion,
			OperatingSystem: GetOS(),
			Architecture:    GetArch(),
		})
	}
	return java, nil
}
