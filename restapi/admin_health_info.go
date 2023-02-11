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
	"bytes"
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zip"

	"github.com/minio/console/cluster"
	"github.com/minio/console/models"
	"github.com/minio/console/operatorapi/operations/operator_api"
	"github.com/minio/madmin-go/v2"
	"github.com/minio/websocket"
	corev1 "k8s.io/api/core/v1"
)

// startHealthInfo starts fetching mc.ServerHealthInfo and
// sends messages with the corresponding data on the websocket connection
func startHealthInfo(ctx context.Context, conn WSConn, client MinioAdmin, deadline *time.Duration) error {
	fmt.Println("In startHealthInfo")
	if deadline == nil {
		return errors.New("duration can't be nil on startHealthInfo")
	}

	// Fetch info of all servers (cluster or single server)
	healthDataTypes := []madmin.HealthDataType{
		madmin.HealthDataTypeMinioInfo,
		madmin.HealthDataTypeMinioConfig,
		madmin.HealthDataTypeSysCPU,
		madmin.HealthDataTypeSysDriveHw,
		madmin.HealthDataTypeSysDocker,
		madmin.HealthDataTypeSysOsInfo,
		madmin.HealthDataTypeSysLoad,
		madmin.HealthDataTypeSysMem,
		madmin.HealthDataTypeSysNet,
		madmin.HealthDataTypeSysProcess,
	}

	var err error
	// Fetch info of all servers (cluster or single server)
	healthInfo, version, err := client.serverHealthInfo(ctx, healthDataTypes, *deadline)
	if err != nil {
		return err
	}

	compressedDiag, err := tarGZ(healthInfo, version)
	if err != nil {
		return err
	}
	encodedDiag := b64.StdEncoding.EncodeToString(compressedDiag)

	type messageReport struct {
		Encoded          string      `json:"encoded"`
		ServerHealthInfo interface{} `json:"serverHealthInfo"`
	}

	report := messageReport{
		Encoded:          encodedDiag,
		ServerHealthInfo: healthInfo,
	}
	message, err := json.Marshal(report)
	if err != nil {
		return err
	}

	// Send Message through websocket connection
	return conn.writeMessage(websocket.TextMessage, message)
}

// compress and tar MinIO diagnostics output
func tarGZ(healthInfo interface{}, version string) ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	gzWriter := gzip.NewWriter(buffer)

	enc := json.NewEncoder(gzWriter)

	header := struct {
		Version string `json:"version"`
	}{Version: version}

	if err := enc.Encode(header); err != nil {
		return nil, err
	}

	if err := enc.Encode(healthInfo); err != nil {
		return nil, err
	}
	err := gzWriter.Close()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// getHealthInfoOptionsFromReq gets duration for startHealthInfo request
// path come as : `/health-info?deadline=2h`
func getDurationFromReq(req *http.Request) (*time.Duration, error) {
	deadlineDuration, err := time.ParseDuration(req.FormValue("deadline"))
	if err != nil {
		return nil, err
	}
	return &deadlineDuration, nil
}
func getPodLogsResponse(session *models.Principal, params operator_api.GetPodLogsParams) (string, *models.Error) {
	ctx, cancel := context.WithCancel(params.HTTPRequest.Context())
	defer cancel()
	clientset, err := cluster.K8sClient(session.STSSessionToken)
	if err != nil {
		return "", ErrorWithContext(ctx, err)
	}
	listOpts := &corev1.PodLogOptions{}
	logs := clientset.CoreV1().Pods(params.Namespace).GetLogs(params.PodName, listOpts)
	buff, err := logs.DoRaw(ctx)
	if err != nil {
		return "", ErrorWithContext(ctx, err)
	}
	makeLogReport(buff, params.Tenant)
	return string(buff), nil
}

func makeLogReport(logs []byte, name string) error {
	w, err := os.Create(name + "-report.zip")
	if err != nil {
		return err
	}
	zipw := zip.NewWriter(w)
	f, err := zipw.Create(name + ".log")
	if err != nil {
		return err
	}
	f.Write(logs)
	return nil
}

// startLogReport generates the Tenant log report and
// sends messages with the corresponding data on the websocket connection
func startLogReport(ctx context.Context, conn WSConn, client MinioAdmin, deadline *time.Duration) error {
	if deadline == nil {
		return errors.New("duration can't be nil on startLogReport")
	}

	//do the thing here
	fmt.Println("You hit this function - the thing should be put here")
	var tempParams operator_api.GetPodLogsParams
	tempParams.Namespace = "ns-1"
	tempParams.Tenant = "newtenant"

	var tempSession *models.Principal

	getPodLogsResponse(tempSession, tempParams)
	var err error
	/*
		// Fetch info of all servers (cluster or single server)
		healthInfo, version, err := client.serverHealthInfo(ctx, healthDataTypes, *deadline)
		if err != nil {
			return err
		}

		compressedDiag, err := tarGZ(healthInfo, version)
		if err != nil {
			return err
		}
		encodedDiag := b64.StdEncoding.EncodeToString(compressedDiag)

		type messageReport struct {
			Encoded          string      `json:"encoded"`
			ServerHealthInfo interface{} `json:"serverHealthInfo"`
		}

		report := messageReport{
			Encoded:          encodedDiag,
			ServerHealthInfo: healthInfo,
		}
	*/
	message, err := json.Marshal("this")
	if err != nil {
		return err
	}

	// Send Message through websocket connection
	return conn.writeMessage(websocket.TextMessage, message)
}
