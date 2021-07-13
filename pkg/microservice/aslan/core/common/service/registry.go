/*
Copyright 2021 The KodeRover Authors.

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

package service

import (
	"fmt"

	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/koderover/zadig/pkg/microservice/aslan/config"
	"github.com/koderover/zadig/pkg/microservice/aslan/core/common/repository/models"
	"github.com/koderover/zadig/pkg/microservice/aslan/core/common/repository/mongodb"
	"github.com/koderover/zadig/pkg/microservice/aslan/core/common/service/kube"
	e "github.com/koderover/zadig/pkg/tool/errors"
	"github.com/koderover/zadig/pkg/util"
)

func FindDefaultRegistry(log *zap.SugaredLogger) (*models.RegistryNamespace, error) {
	// TODO: 多租户适配
	resp, err := mongodb.NewRegistryNamespaceColl().Find(&mongodb.FindRegOps{
		IsDefault: true,
	})

	if err != nil {
		log.Warnf("RegistryNamespace.Find error: %v", err)
		resp = &models.RegistryNamespace{
			RegAddr:   config.RegistryAddress(),
			AccessKey: config.RegistryAccessKey(),
			SecretKey: config.RegistrySecretKey(),
			Namespace: config.RegistryNamespace(),
		}
	}

	ak := resp.AccessKey
	sk := resp.SecretKey
	if resp.RegProvider == config.SWRProvider {
		ak = fmt.Sprintf("%s@%s", resp.Region, resp.AccessKey)
		sk = util.ComputeHmacSha256(resp.AccessKey, resp.SecretKey)
	}
	resp.AccessKey = ak
	resp.SecretKey = sk

	return resp, nil
}

func GetDefaultRegistryNamespace(log *zap.SugaredLogger) (*models.RegistryNamespace, error) {
	resp, err := mongodb.NewRegistryNamespaceColl().Find(&mongodb.FindRegOps{IsDefault: true})
	if err != nil {
		log.Errorf("get default registry error: %v", err)
		return resp, fmt.Errorf("get default registry error: %v", err)
	}
	return resp, nil
}

func ListRegistryNamespaces(log *zap.SugaredLogger) ([]*models.RegistryNamespace, error) {
	resp, err := mongodb.NewRegistryNamespaceColl().FindAll(&mongodb.FindRegOps{})
	if err != nil {
		log.Errorf("RegistryNamespace.List error: %v", err)
		return resp, fmt.Errorf("RegistryNamespace.List error: %v", err)
	}
	return resp, nil
}

func EnsureDefaultRegistrySecret(namespace string, kubeClient client.Client, log *zap.SugaredLogger) error {
	reg, err := FindDefaultRegistry(log)
	if err != nil {
		log.Errorf(
			"service.EnsureRegistrySecret: failed to find default candidate registry: %s %v",
			namespace, err,
		)
		return err
	}

	err = kube.CreateOrUpdateRegistrySecret(namespace, reg, kubeClient)
	if err != nil {
		log.Errorf("[%s] CreateDockerSecret error: %v", namespace, err)
		return e.ErrUpdateSecret.AddDesc(e.CreateDefaultRegistryErrMsg)
	}

	return nil
}
