// This file is part of MinIO Console Server
// Copyright (c) 2021 MinIO, Inc.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package restapi

import (
	"context"

	"github.com/minio/minio-go/v7/pkg/set"

	"github.com/minio/console/cluster"

	"github.com/go-openapi/runtime/middleware"
	"github.com/minio/console/models"
	"github.com/minio/console/restapi/operations"
	"github.com/minio/console/restapi/operations/admin_api"

	//corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func registerAdminNodesHandlers(api *operations.ConsoleAPI) {
	// return usage stats
	api.AdminAPIAdminInfoHandler = admin_api.ListNodeLabelsHandlerFunc(func(params admin_api.AdminInfoParams, session *models.Principal) middleware.Responder {
		infoResp, err := getAdminInfoResponse(session, params)
		if err != nil {
			return admin_api.NewAdminInfoDefault(int(err.Code)).WithPayload(err)
		}
		return admin_api.NewAdminInfoOK().WithPayload(infoResp)
	})
}

type NodeResourceInfo struct {
	Name              string
	AllocatableMemory int64
	AllocatableCPU    int64
}

func getNodeLabels(ctx context.Context, clientset v1.CoreV1Interface) (*models.NodeLabels, error) {
	// get all nodes from cluster
	nodes, err := clientset.Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	// make a map[string]set to avoid duplicate values
	keyValueSet := map[string]set.StringSet{}

	for _, node := range nodes.Items {
		for k, v := range node.Labels {
			if _, ok := keyValueSet[k]; !ok {
				keyValueSet[k] = set.NewStringSet()
			}
			keyValueSet[k].Add(v)
		}
	}

	// convert to output
	res := models.NodeLabels{}
	for k, valSet := range keyValueSet {
		res[k] = valSet.ToSlice()
	}

	return &res, nil
}

func getNodeLabelsResponse(ctx context.Context, session *models.Principal) (*models.NodeLabels, *models.Error) {
	client, err := cluster.K8sClient(session.STSSessionToken)
	if err != nil {
		return nil, prepareError(err)
	}

	clusterResources, err := getNodeLabels(ctx, client.CoreV1())
	if err != nil {
		return nil, prepareError(err)
	}
	return clusterResources, nil
}
