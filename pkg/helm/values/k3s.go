package values

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/loft-sh/vcluster/pkg/helm"
	"github.com/loft-sh/vcluster/pkg/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"
)

var K3SVersionMap = map[string]string{
	"1.21": "rancher/k3s:v1.21.4-k3s1",
	"1.20": "rancher/k3s:v1.20.11-k3s2",
	"1.19": "rancher/k3s:v1.19.13-k3s1",
	"1.18": "rancher/k3s:v1.18.20-k3s1",
	"1.17": "rancher/k3s:v1.17.17-k3s1",
	"1.16": "rancher/k3s:v1.16.15-k3s1",
}

const noDeployValues = `  baseArgs:
    - server
    - --write-kubeconfig=/k3s-config/kube-config.yaml
    - --data-dir=/data
    - --no-deploy=traefik,servicelb,metrics-server,local-storage
    - --disable-network-policy
    - --disable-agent
    - --disable-scheduler
    - --disable-cloud-controller
    - --flannel-backend=none
    - --kube-controller-manager-arg=controllers=*,-nodeipam,-nodelifecycle,-persistentvolume-binder,-attachdetach,-persistentvolume-expander,-cloud-node-lifecycle`

var baseArgsMap = map[string]string{
	"1.17": noDeployValues,
	"1.16": noDeployValues,
}

var replaceRegEx = regexp.MustCompile("[^0-9]+")
var errorMessageFind = "provided IP is not in the valid range. The range of valid IPs is "

func getDefaultK3SReleaseValues(chartOptions *helm.ChartOptions, log log.Logger) (string, error) {
	var (
		image               = chartOptions.K3SImage
		serverVersionString string
		serverMinorInt      int
		err                 error
	)

	if image == "" {
		serverVersionString = GetKubernetesVersion(chartOptions.KubernetesVersion)
		serverMinorInt, err = GetKubernetesMinorVersion(chartOptions.KubernetesVersion)
		if err != nil {
			return "", err
		}

		var ok bool
		image, ok = K3SVersionMap[serverVersionString]
		if !ok {
			if serverMinorInt > 21 {
				log.Infof("officially unsupported host server version %s, will fallback to virtual cluster version v1.21", serverVersionString)
				image = K3SVersionMap["1.21"]
				serverVersionString = "1.21"
			} else {
				log.Infof("officially unsupported host server version %s, will fallback to virtual cluster version v1.16", serverVersionString)
				image = K3SVersionMap["1.16"]
				serverVersionString = "1.16"
			}
		}
	}

	// build values
	values := `vcluster:
  image: ##IMAGE##
##BASEARGS##
`
	values = strings.ReplaceAll(values, "##IMAGE##", image)
	if chartOptions.K3SImage == "" {
		baseArgs := baseArgsMap[serverVersionString]
		values = strings.ReplaceAll(values, "##BASEARGS##", baseArgs)
	}

	return addCommonReleaseValues(values, chartOptions)
}

func addCommonReleaseValues(values string, chartOptions *helm.ChartOptions) (string, error) {
	values += `
serviceCIDR: ##CIDR##
storage:
  size: 5Gi`
	if chartOptions.DisableIngressSync {
		values += `
syncer:
  extraArgs: ["--disable-sync-resources=ingresses"]`
	}
	if chartOptions.CreateClusterRole {
		values += `
rbac:
  clusterRole:
    create: true`
	}

	if chartOptions.Expose {
		values += `
service:
  type: LoadBalancer`
	}

	values = strings.ReplaceAll(values, "##CIDR##", chartOptions.CIDR)
	values = strings.TrimSpace(values)
	return values, nil
}

func ParseKubernetesVersionInfo(versionStr string) (*version.Info, error) {
	if versionStr[0] == 'v' {
		versionStr = versionStr[1:]
	}

	splittedVersion := strings.Split(versionStr, ".")
	if len(splittedVersion) != 2 && len(splittedVersion) != 3 {
		return nil, fmt.Errorf("unrecognized kubernetes version %s, please use format vX.X", versionStr)
	}

	major := splittedVersion[0]
	minor := splittedVersion[1]

	return &version.Info{
		Major: major,
		Minor: minor,
	}, nil
}

func GetKubernetesVersion(serverVersion *version.Info) string {
	return replaceRegEx.ReplaceAllString(serverVersion.Major, "") + "." + replaceRegEx.ReplaceAllString(serverVersion.Minor, "")
}

func GetKubernetesMinorVersion(serverVersion *version.Info) (int, error) {
	return strconv.Atoi(replaceRegEx.ReplaceAllString(serverVersion.Minor, ""))
}

func GetServiceCIDR(client kubernetes.Interface, namespace string, ipv6 bool) (string, error) {
	clusterIP := "4.4.4.4"
	if ipv6 {
		// https://www.ietf.org/rfc/rfc3849.txt
		clusterIP = "2001:DB8::1"
	}
	_, err := client.CoreV1().Services(namespace).Create(context.Background(), &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "test-service-",
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port: 80,
				},
			},
			ClusterIP: clusterIP,
		},
	}, metav1.CreateOptions{})
	if err == nil {
		return "", fmt.Errorf("couldn't find cluster service cidr, will fallback to 10.96.0.0/12, however this is probably wrong, please make sure the host cluster service cidr and virtual cluster service cidr match")
	}

	errorMessage := err.Error()
	idx := strings.Index(errorMessage, errorMessageFind)
	if idx == -1 {
		return "", fmt.Errorf("couldn't find cluster service cidr (" + errorMessage + "), will fallback to 10.96.0.0/12, however this is probably wrong, please make sure the host cluster service cidr and virtual cluster service cidr match")
	}

	return strings.TrimSpace(errorMessage[idx+len(errorMessageFind):]), nil
}
