package values

import (
	"github.com/loft-sh/vcluster/pkg/helm"
	"github.com/loft-sh/vcluster/pkg/log"
	"k8s.io/apimachinery/pkg/version"
)

func GetDefaultReleaseValues(serverVersion *version.Info, chartOptions *helm.ChartOptions, log log.Logger) (string, error) {
	if chartOptions.ChartName == helm.K3SChart {
		return getDefaultK3SReleaseValues(serverVersion, chartOptions, log)
	} else if chartOptions.ChartName == helm.K0SChart {
		return getDefaultK0SReleaseValues(serverVersion, chartOptions, log)
	} else if chartOptions.ChartName == helm.K8SChart {
		return getDefaultK8SReleaseValues(serverVersion, chartOptions, log)
	}

	return "", nil
}
