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


/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"gitlab.alibaba-inc.com/polar-as/polar-mpd-controller/pkg/business/monitor"

	"k8s.io/component-base/logs"
	"k8s.io/klog/klogr"

	globalmgr "gitlab.alibaba-inc.com/polar-as/polar-common-domain/manager"
	mpdv1 "gitlab.alibaba-inc.com/polar-as/polar-mpd-controller/apis/mpd/v1"
	"gitlab.alibaba-inc.com/polar-as/polar-mpd-controller/pkg/bizapis"
	"gitlab.alibaba-inc.com/polar-as/polar-mpd-controller/pkg/configuration"
	controllers "gitlab.alibaba-inc.com/polar-as/polar-mpd-controller/pkg/controllers/mpd"
	v "gitlab.alibaba-inc.com/polar-as/polar-mpd-controller/pkg/version"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = mpdv1.AddToScheme(scheme)
	_ = mpdv1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	ctrl.SetLogger(klogr.New())
	logs.InitLogs()
	defer logs.FlushLogs()

	parseFlags()
	printCodeBranchInfo()
	startHTTPServer(configuration.GetConfig().PodPort)
	mgr := createManager(configuration.GetConfig().EnableLeaderElection)
	globalmgr.RegisterManager(mgr)
	addLeaderRunners(mgr)
	setupMpdClusterReconcile(mgr)
	signals := ctrl.SetupSignalHandler()
	go monitor.CreateSharedStorageClusterStatusSyncMonitor(setupLog, 60*time.Second).Start(signals)

	// +kubebuilder:scaffold:builder
	setupLog.Info("starting manager")
	if err := mgr.Start(signals); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func parseFlags() {
	configuration.GetConfig()
}

func setupMpdClusterReconcile(mgr manager.Manager) {
	if err := (&controllers.MPDClusterReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("MpdCluster"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "MpdCluster")
		os.Exit(1)
	}
}

func createManager(enableLeaderElection bool) manager.Manager {
	mgr, err := manager.New(ctrl.GetConfigOrDie(), manager.Options{
		Scheme:             scheme,
		LeaderElection:     enableLeaderElection,
		Port:               configuration.GetConfig().WebhookPort,
		MetricsBindAddress: "0",
		LeaderElectionID:   "mpd.polardb.aliyun.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to set up overall controller manager")
		os.Exit(1)
	}
	return mgr
}

func addLeaderRunners(mgr manager.Manager) {
	leaderRunnable := &controllers.PolarOptLead{}
	leaderRunnable.OwnedMgr = mgr
	leadErr := mgr.Add(leaderRunnable)
	if leadErr != nil {
		setupLog.Error(leadErr, "-------*****-----leadErr err: %v---")
	}
}

func startHTTPServer(port int) {
	go func() {
		bizapis.Start()
		addr := fmt.Sprintf(":%d", port)
		setupLog.Info(fmt.Sprintf("http server listen: %s", addr))
		err := http.ListenAndServe(addr, nil)
		setupLog.Error(err, "http listen error")
	}()
}

func printCodeBranchInfo() {
	fmt.Printf("---------------------------------------------------------------------------------------------\n")
	fmt.Printf("|                                                                                           |\n")
	fmt.Printf("| branch: %v commitId :%v \n", v.GitBranch, v.GitCommitID)
	fmt.Printf("| repo: %v\n", v.GitCommitRepo)
	fmt.Printf("| commitDate: %v\n", v.GitCommitDate)
	fmt.Printf("|                                                                                           |\n")
	fmt.Printf("---------------------------------------------------------------------------------------------\n")
}
