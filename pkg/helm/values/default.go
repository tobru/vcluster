package values

import (
	"github.com/loft-sh/vcluster/pkg/helm"
	"github.com/loft-sh/vcluster/pkg/log"
	"k8s.io/client-go/kubernetes"
)

func GetDefaultReleaseValues(client kubernetes.Interface, chartOptions *helm.ChartOptions, log log.Logger) (string, error) {
	if chartOptions.ChartName == helm.K3SChart {
		return getDefaultK3SReleaseValues(client, chartOptions, log)
	} else if chartOptions.ChartName == helm.K0SChart {
		return getDefaultK0SReleaseValues(client, chartOptions, log)
	} else if chartOptions.ChartName == helm.K8SChart {
		return getDefaultK8SReleaseValues(client, chartOptions, log)
	}

	return "", nil
}
