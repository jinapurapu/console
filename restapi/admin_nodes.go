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
	"fmt"
	"time"

	"github.com/minio/minio-go/v7/pkg/set"

	"github.com/minio/console/cluster"

	"github.com/minio/madmin-go"

	"github.com/go-openapi/runtime/middleware"
	"github.com/minio/console/models"
	"github.com/minio/console/restapi/operations"
	"github.com/minio/console/restapi/operations/admin_api"
	"github.com/minio/console/restapi/operations/user_api"

	//corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func registerAdminNodesHandlers(api *operations.ConsoleAPI) {
	// return usage stats
	api.UserAPIListNodeLabelsHandler = user_api.ListNodeLabelsHandlerFunc(func(params user_api.ListNodeLabelsParams, session *models.Principal) middleware.Responder {

		nodeLabelsResp, err := getNodeLabelsResponse(params, principal)
		if err != nil {
			return admin_api.NewAdminInfoDefault(int(err.Code)).WithPayload(err)
		}
		return user_api.NewListNodeLabelsOK().WithPayload(*nodeLabelsResp)
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

func getAdminResources(ctx context.Context, mAdmin *madmin.AdminClient) ([]*models.ServerDrives, error) {

	serverInfo, err := mAdmin.ServerInfo(ctx)
	if err != nil {
		return nil, err
	}

	var drives = []*models.ServerDrives{}
	for _, serv := range serverInfo.Servers {

		for _, drive := range serv.Disks {
			drives = append(drives, &models.ServerDrives{
				State:          drive.State,
				UUID:           drive.UUID,
				Endpoint:       drive.Endpoint,
				RootDisk:       drive.RootDisk,
				DrivePath:      drive.DrivePath,
				Healing:        drive.Healing,
				Model:          drive.Model,
				TotalSpace:     int64(drive.TotalSpace),
				UsedSpace:      int64(drive.UsedSpace),
				AvailableSpace: int64(drive.AvailableSpace),
			})
		}

	}
	return drives, nil
}

func getNodeLabelsResponse(params *models.ListNodeLabelsParams, session *models.Principal) (*models.NodeLabels, *models.Error) {

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	client, err := cluster.K8sClient(session.STSSessionToken)
	if err != nil {
		return nil, prepareError(err)
	}

	mAdmin, err := NewMinioAdminClient(session)
	if err != nil {
		return nil, prepareError(err)
	}

	clusterResources, err := getNodeLabels(ctx, client.CoreV1())
	adminResources, err := getAdminResources(ctx, mAdmin)

	fmt.Println("Admin resources is: ", adminResources)

	if err != nil {
		return nil, prepareError(err)
	}
	return clusterResources, nil
}
