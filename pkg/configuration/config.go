/* 
*Copyright (c) 2019-2021, Alibaba Group Holding Limited;
*Licensed under the Apache License, Version 2.0 (the "License");
*you may not use this file except in compliance with the License.
*You may obtain a copy of the License at

*   http://www.apache.org/licenses/LICENSE-2.0

*Unless required by applicable law or agreed to in writing, software
*distributed under the License is distributed on an "AS IS" BASIS,
*WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
*See the License for the specific language governing permissions and
*limitations under the License.
 */


package configuration

import (
	"flag"
	"sync"

	"gitlab.alibaba-inc.com/polar-as/polar-mpd-controller/pkg/define"

	commonconfig "gitlab.alibaba-inc.com/polar-as/polar-common-domain/configuration"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog"
)

var (
	once   sync.Once
	config *Config
)

// Config ...
type Config struct {
	FilterOperatorName      string
	PodPort                 int
	WebhookPort             int
	MaxConcurrentReconciles int
	EnableLeaderElection    bool
	DbClusterLogDir         string
	WorkFlowMetaDir         string
	CmCpuReqLimit           string
	CmMemReqLimit           string
	HwCheck                 bool
	ImagePullPolicy         corev1.PullPolicy
	RunPodInPrivilegedMode  bool
	CmLogHostPath           string
	CmLogMountPath          string
	DbTypeLabel             string
}

// GetConfig ...
func GetConfig() *Config {
	once.Do(func() {
		conf := &Config{}
		flag.BoolVar(&conf.EnableLeaderElection, "enable-leader-election", false,
			"Enable leader election for controller manager. "+
				"Enabling this will ensure there is only one active controller manager.")
		flag.StringVar(&conf.WorkFlowMetaDir, "work-flow-meta-dir", "./pkg/workflow", "work flow meta dir specify")
		flag.IntVar(&conf.PodPort, "port", 6079, "rest api port")
		flag.StringVar(&conf.DbClusterLogDir, "dbcluster-log-dir", "/disk1/polardb/", "dbcluster log path in host")
		flag.IntVar(&conf.WebhookPort, "webhook-port", 6070, "webhook port")
		flag.IntVar(&conf.MaxConcurrentReconciles, "max-concurrent-reconciles", 26, "Max Concurrent Reconciles.")
		flag.StringVar(&conf.FilterOperatorName, "filter-operator-name", "polar-mpd", "Filter operatorName in spec.")
		flag.BoolVar(&conf.HwCheck, "hw_check", true, "whether to check hardware when create pod")
		flag.BoolVar(&conf.RunPodInPrivilegedMode, "run-pod-in-privileged-mode", conf.RunPodInPrivilegedMode, "run pod in privileged mode")
		flag.StringVar(&conf.CmCpuReqLimit, "cm-cpu-req-limit", "200m/500m", "cluster manager cpu request/limit")
		flag.StringVar(&conf.CmMemReqLimit, "cm-mem-req-limit", "200Mi/512Mi", "cluster manager mem request/limit")
		flag.StringVar(&conf.CmLogHostPath, "cm-log-host-path", "/var/log/polardb-box/polardb-cm", "cluster manager log host path")
		flag.StringVar(&conf.CmLogMountPath, "cm-log-mount-path", "/root/polardb_cluster_manager/log", "cluster manager log mount path")
		flag.StringVar(&conf.DbTypeLabel, "db-type-label", "PostgreSQL", "db type label")

		var imagePullPolicy = ""
		flag.StringVar(&imagePullPolicy, "image-pull-policy", "IfNotPresent", "image pull policy")
		flag.Parse()
		conf.ImagePullPolicy = corev1.PullPolicy(imagePullPolicy)
		config = conf
		commonconfig.SetConfig(&commonconfig.Config{
			DbClusterLogDir:        config.DbClusterLogDir,
			CmCpuReqLimit:          config.CmCpuReqLimit,
			CmMemReqLimit:          config.CmMemReqLimit,
			HwCheck:                config.HwCheck,
			ImagePullPolicy:        config.ImagePullPolicy,
			RunPodInPrivilegedMode: config.RunPodInPrivilegedMode,
			CmLogHostPath:          config.CmLogHostPath,
			CmLogMountPath:         config.CmLogMountPath,
			DbTypeLabel:            config.DbTypeLabel,

			ClusterInfoDbInsTypeRw:       define.InsTypeRw,
			ClusterInfoDbInsTypeRo:       define.InsTypeRo,
			ParamsTemplateName:           "postgresql-1-0-mycnf-template",
			MinorVersionCmName:           "postgresql-1-0-minor-version-info",
			ExternalPortRangeAnnotation:  "polarbox.mpd.external.port.range",
			ReservedPortRangesAnnotation: "polarbox.mpd.reserved.port.ranges",
			PortRangeCmName:              "polardb4mpd-controller",
			InsIdRangeCmName:             "polardb4mpd-controller",
			NodeMaintainLabelName:        "mpd.polardb.aliyun.com/maintainNode",
			AccountMetaClusterLabelName:  "mpdcluster_name",
			ControllerManagerRoleName:    "polardb4mpd-controller-manager-role",
		})
		klog.Infof("FLAG: %v", config)
	})
	return config
}
