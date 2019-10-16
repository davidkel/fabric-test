// Copyright IBM Corp. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package launcher

import (
	//"flag"
	"io/ioutil"

	"github.com/hyperledger/fabric-test/tools/operator/launcher/dockercompose"
	"github.com/hyperledger/fabric-test/tools/operator/launcher/k8s"
	"github.com/hyperledger/fabric-test/tools/operator/launcher/nl"
	"github.com/hyperledger/fabric-test/tools/operator/logger"
	"github.com/hyperledger/fabric-test/tools/operator/networkspec"
	"github.com/hyperledger/fabric-test/tools/operator/ytt"
	"github.com/hyperledger/fabric-test/tools/operator/paths"
)

// var networkSpecPath = flag.String("i", "", "Network spec input file path (Required)")
// var kubeConfigPath = flag.String("k", "", "Kube config file path (Optional for local network)")
// var action = flag.String("a", "up", "Set action(up or down) (default is up)")

func validateArguments(networkSpecPath string, kubeConfigPath string) {

	if networkSpecPath == "" {
		logger.CRIT(nil, "Config file not provided")
	} else if kubeConfigPath == "" {
		logger.INFO("Kube config file not provided, proceeding with local environment")
	}
}

func doAction(action, env, kubeConfigPath string, config networkspec.Config) error {

	var err error
	switch env {
	case "k8s":
		k8s := k8s.K8s{KubeConfigPath: kubeConfigPath, Config: config}
		err = k8s.K8sNetwork(action)
	case "docker":
		dc := dockercompose.DockerCompose{Config: config}
		err = dc.DockerNetwork(action)
	}
	if err != nil {
		logger.CRIT(err)
	}
	return nil
}

func validateBasicConsensusConfig(config networkspec.Config) {

	ordererType := config.Orderer.OrdererType
	if ordererType == "solo" {
		if !(len(config.OrdererOrganizations) == 1 && config.OrdererOrganizations[0].NumOrderers == 1) {
			logger.CRIT(nil, "Consensus type solo should have only one orderer organization and one orderer")
		}
	} else if ordererType == "kafka" {
		if len(config.OrdererOrganizations) != 1 {
			logger.CRIT(nil, "Consensus type kafka should have only one orderer organization")
		}
	}
}

func Launcher(action, env, kubeConfigPath, networkSpecPath string) error {

	var yttObject ytt.YTT
	err := yttObject.DownloadYtt()
	if err != nil {
		logger.CRIT(err)
	}
	validateArguments(networkSpecPath, kubeConfigPath)
	contents, _ := ioutil.ReadFile(networkSpecPath)
	contents = append([]byte("#@data/values \n"), contents...)
	inputPath := paths.JoinPath(paths.TemplatesDir(), "input.yaml")
	ioutil.WriteFile(inputPath, contents, 0644)

	var network nl.Network
	config, err := network.GetConfigData(inputPath)

	if err != nil {
		logger.CRIT(err)
	}

	validateBasicConsensusConfig(config)
	err = doAction(action, env, kubeConfigPath, config)
	if err != nil {
		return err
	}
	return nil
}