package create

import (
	"fmt"
	"strings"

	"github.com/loft-sh/vcluster/pkg/helm"
	"github.com/loft-sh/vcluster/pkg/util"
)

var AllowedDistros = []string{"k3s", "k0s", "k8s"}

// CreateOptions holds the create cmd options
type CreateOptions struct {
	ChartVersion string
	ChartName    string
	ChartRepo    string
	K3SImage     string
	Distro       string
	CIDR         string
	ExtraValues  []string

	KubernetesVersion string

	CreateNamespace    bool
	DisableIngressSync bool
	CreateClusterRole  bool
	Expose             bool
	Connect            bool
	Upgrade            bool

	ReleaseValues string
}

func (o *CreateOptions) ToChartOptions(namespace string) (*helm.ChartOptions, error) {
	if !util.Contains(o.Distro, AllowedDistros) {
		return nil, fmt.Errorf("unsupported distro %s, please select one of: %s", o.Distro, strings.Join(AllowedDistros, ", "))
	}

	if o.ChartName == "vcluster" && o.Distro != "k3s" {
		o.ChartName += "-" + o.Distro
	}

	return &helm.ChartOptions{
		ChartName:          o.ChartName,
		ChartRepo:          o.ChartRepo,
		ChartVersion:       o.ChartVersion,
		CIDR:               o.CIDR,
		CreateClusterRole:  o.CreateClusterRole,
		DisableIngressSync: o.DisableIngressSync,
		Expose:             o.Expose,
		K3SImage:           o.K3SImage,
		KubernetesVersion:  o.KubernetesVersion,
		Namespace:          namespace,
	}, nil
}
