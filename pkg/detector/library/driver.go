package library

import (
	"golang.org/x/xerrors"

	ftypes "github.com/aquasecurity/fanal/types"
	ecosystem "github.com/aquasecurity/trivy-db/pkg/vulnsrc/ghsa"
	"github.com/aquasecurity/trivy-db/pkg/vulnsrc/vulnerability"
	"github.com/aquasecurity/trivy/pkg/detector/library/cargo"
	"github.com/aquasecurity/trivy/pkg/detector/library/comparer"
	"github.com/aquasecurity/trivy/pkg/detector/library/composer"
	"github.com/aquasecurity/trivy/pkg/detector/library/ghsa"
	"github.com/aquasecurity/trivy/pkg/detector/library/maven"
	"github.com/aquasecurity/trivy/pkg/detector/library/npm"
	"github.com/aquasecurity/trivy/pkg/types"
)

type advisory interface {
	DetectVulnerabilities(string, string) ([]types.DetectedVulnerability, error)
}

// NewDriver returns a driver according to the library type
func NewDriver(libType string) (Driver, error) {
	var driver Driver
	switch libType {
	case ftypes.Cargo:
		driver = newCargoDriver()
	case ftypes.Composer:
		driver = newComposerDriver()
	case ftypes.Npm, ftypes.Yarn, ftypes.NodePkg, ftypes.JavaScript:
		driver = newNpmDriver()
	case ftypes.NuGet:
		driver = newNugetDriver()
	case ftypes.Jar:
		driver = newMavenDriver()
	case ftypes.GoBinary, ftypes.GoMod:
		driver = Driver{
			ecosystem:  vulnerability.Go,
			advisories: []advisory{NewAdvisory(vulnerability.Go, comparer.GenericComparer{})},
		}
	default:
		return Driver{}, xerrors.Errorf("unsupported type %s", libType)
	}
	return driver, nil
}

// Driver implements the advisory
type Driver struct {
	ecosystem  string
	advisories []advisory
}

// Aggregate aggregates drivers
func Aggregate(ecosystem string, advisories ...advisory) Driver {
	return Driver{ecosystem: ecosystem, advisories: advisories}
}

// Detect scans and returns vulnerabilities
func (d *Driver) Detect(pkgName string, pkgVer string) ([]types.DetectedVulnerability, error) {
	var detectedVulnerabilities []types.DetectedVulnerability
	uniqVulnIDMap := make(map[string]struct{})
	for _, adv := range d.advisories {
		vulns, err := adv.DetectVulnerabilities(pkgName, pkgVer)
		if err != nil {
			return nil, xerrors.Errorf("failed to detect vulnerabilities: %w", err)
		}
		for _, vuln := range vulns {
			if _, ok := uniqVulnIDMap[vuln.VulnerabilityID]; ok {
				continue
			}
			uniqVulnIDMap[vuln.VulnerabilityID] = struct{}{}
			detectedVulnerabilities = append(detectedVulnerabilities, vuln)
		}
	}

	return detectedVulnerabilities, nil
}

// Type returns the driver ecosystem
func (d *Driver) Type() string {
	return d.ecosystem
}

func newComposerDriver() Driver {
	c := comparer.GenericComparer{}
	return Aggregate(vulnerability.Composer, NewAdvisory(vulnerability.Composer, c), composer.NewAdvisory(), ghsa.NewAdvisory(ecosystem.Composer, c))
}

func newCargoDriver() Driver {
	return Aggregate(vulnerability.Cargo, NewAdvisory(vulnerability.Cargo, comparer.GenericComparer{}), cargo.NewAdvisory())
}

func newNpmDriver() Driver {
	c := npm.Comparer{}
	return Aggregate(vulnerability.Npm, NewAdvisory(vulnerability.Npm, c), npm.NewAdvisory(), ghsa.NewAdvisory(ecosystem.Npm, c))
}


func newNugetDriver() Driver {
	c := comparer.GenericComparer{}
	return Aggregate(vulnerability.NuGet, NewAdvisory(vulnerability.NuGet, c), ghsa.NewAdvisory(ecosystem.Nuget, c))
}

func newMavenDriver() Driver {
	c := maven.Comparer{}
	return Aggregate(vulnerability.Maven, NewAdvisory(vulnerability.Maven, c), ghsa.NewAdvisory(ecosystem.Maven, c))
}
