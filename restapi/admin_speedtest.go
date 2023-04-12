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
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/minio/console/models"
	xhttp "github.com/minio/console/pkg/http"
	subnet "github.com/minio/console/pkg/subnet"
	"github.com/minio/madmin-go/v2"
	"github.com/minio/websocket"
)

// getSpeedtesthOptionsFromReq gets duration, size & concurrent requests from a websocket
// path come as : `/speedtest?duration=2h&size=12MiB&concurrent=10`
func getSpeedtestOptionsFromReq(req *http.Request) (*madmin.SpeedtestOpts, error) {
	optionsSet := madmin.SpeedtestOpts{}

	queryPairs := req.URL.Query()

	paramDuration := queryPairs.Get("duration")

	if paramDuration == "" {
		paramDuration = "10s"
	}

	duration, err := time.ParseDuration(paramDuration)
	if err != nil {
		return nil, fmt.Errorf("unable to parse duration: %s", paramDuration)
	}

	if duration <= 0 {
		return nil, fmt.Errorf("duration cannot be 0 or negative")
	}

	optionsSet.Duration = duration

	paramSize := queryPairs.Get("size")

	if paramSize == "" {
		paramSize = "64MiB"
	}

	size, err := humanize.ParseBytes(paramSize)
	if err != nil {
		return nil, fmt.Errorf("unable to parse object size")
	}

	if size < 0 {
		return nil, fmt.Errorf("size is expected to be atleast 0 bytes")
	}

	optionsSet.Size = int(size)

	paramConcurrent := queryPairs.Get("concurrent")

	if paramConcurrent == "" {
		paramConcurrent = "32"
	}

	concurrent, err := strconv.Atoi(paramConcurrent)
	if err != nil {
		return nil, fmt.Errorf("invalid concurrent value: %s", paramConcurrent)
	}

	if concurrent <= 0 {
		return nil, fmt.Errorf("concurrency cannot be '0' or negative")
	}

	optionsSet.Concurrency = concurrent

	autotune := queryPairs.Get("autotune")

	if autotune == "true" {
		optionsSet.Autotune = true
	}

	return &optionsSet, nil
}

func startSpeedtest(ctx context.Context, conn WSConn, client MinioAdmin, speedtestOpts *madmin.SpeedtestOpts) error {
	fmt.Println("Starting speedtest")
	speedtestRes, err := client.speedtest(ctx, *speedtestOpts)
	if err != nil {
		LogError("error initializing speedtest: %v", err)
		return err
	}

	for result := range speedtestRes {
		if len(result.Version) > 0 {
			// Serializing message
			bytes, err := json.Marshal(result)
			if err != nil {
				LogError("error serializing json: %v", err)
				return err
			}

			// Send Message through websocket connection
			err = conn.writeMessage(websocket.TextMessage, bytes)
			if err != nil {
				LogError("error writing speedtest response: %v", err)
				return err
			}
			subnetResp, err := sendPerfInfoToSubnet(ctx, result, client)

			fmt.Println("Sent to subnet! subnetresp: ", subnetResp, " err: ", err)
		}
	}

	return nil
}

func sendPerfInfoToSubnet(ctx context.Context, perfInfo interface{}, client MinioAdmin) (string, error) {
	fmt.Println("Sending pefinfotosubnet")
	perfUploadInfo := make([]interface{}, 2)

	filename := fmt.Sprintf("perf_%d.json", time.Now().Unix())

	subnetUploadURL := subnet.UploadURL("perf", filename)
	subnetHTTPClient := &xhttp.Client{Client: GetConsoleHTTPClient("")}

	subnetTokenConfig, e := GetSubnetKeyFromMinIOConfig(ctx, client)
	if e != nil {
		return "", e
	}
	apiKey := subnetTokenConfig.APIKey
	session := models.Principal{
		STSSessionToken: apiKey,
	}
	fmt.Println(" subnet key: ", apiKey)
	mAdmin, err := NewMinioAdminClient(&session)
	if err != nil {
		return "", err
	}
	fmt.Println("Got madmin: ", mAdmin)
	// create a minioClient interface implementation
	// defining the client to be used
	//adminClient := AdminClient{Client: mAdmin}
	//fmt.Println("Got AdminClient: ", adminClient)
	headers := subnet.UploadAuthHeaders(apiKey)
	//fmt.Println("Getting serverinfo ctx: ", ctx)
	//serverCtx, cancel := context.WithCancel(context.Background())
	//defer cancel()
	//fmt.Println("Getting serverinfo serverCtx: ", serverCtx)
	serverInfo, err := client.serverInfo(ctx)
	//serverInfo, err := getServerInfo(ctx, adminClient)
	if err != nil {
		return "", err
	}
	fmt.Println("Getting clusterinfo serverInfo: ", serverInfo)
	//{"deployment_id":,"cluster_name":,"used_capacity":,"info":{"minio_version":"2023-04-11T18:17:46Z","no_of_server_pools":1,"no_of_servers":4,"no_of_drives":16,"no_of_buckets":1,"no_of_objects":5,"total_drive_space":952142528512,"used_drive_space":827428569088}}
	clusterInfo := subnet.GetClusterRegInfo(serverInfo)
	fmt.Println("clustrInfo: ", clusterInfo)
	perfUploadInfo[1] = clusterInfo
	perfUploadInfo[0] = perfInfo
	fmt.Println("lets upload to subnet! perfUploadInfo: ", perfUploadInfo, " subnetHTTPClient: ", subnetHTTPClient, " subnetUploadURL: ", subnetUploadURL, " headers: ", headers)
	resp, e := subnet.UploadFileToSubnet(perfUploadInfo, subnetHTTPClient, filename, subnetUploadURL, headers)
	if e != nil {
		return "", e
	}
	type SubnetResponse struct {
		ClusterURL string `json:"cluster_url,omitempty"`
	}

	var subnetResp SubnetResponse
	e = json.Unmarshal([]byte(resp), &subnetResp)
	if e != nil {
		return "", e
	}
	if len(subnetResp.ClusterURL) != 0 {
		subnetClusterURL := strings.ReplaceAll(subnetResp.ClusterURL, "%2f", "/")
		return subnetClusterURL, nil
	}

	return "", ErrSubnetUploadFail
}
func getServerInfo(ctx context.Context, adminClient MinioAdmin) (*models.AdminInfoResponse, error) {
	/*
		prometheusStatus := models.AdminInfoResponseAdvancedMetricsStatusAvailable
		if prometheusURL == "" {
			prometheusStatus = models.AdminInfoResponseAdvancedMetricsStatusNotConfigured
		}
		if prometheusURL != "" && !testPrometheusURL(ctx, prometheusURL) {
			prometheusStatus = models.AdminInfoResponseAdvancedMetricsStatusUnavailable
		}
	*/
	sessionResp := &models.AdminInfoResponse{}

	doneCh := make(chan error)
	go func() {
		defer close(doneCh)
		// serialize output
		fmt.Println("get the info adminClient: ", adminClient)
		usage, err := GetAdminInfo(ctx, adminClient)
		if err != nil {
			doneCh <- err
		}
		if usage != nil {
			sessionResp.Buckets = usage.Buckets
			sessionResp.Objects = usage.Objects
			sessionResp.Usage = usage.Usage
			sessionResp.Servers = usage.Servers
			sessionResp.Backend = usage.Backend
		}
	}()

	var wdgts []*models.Widget

	sessionResp.Widgets = wdgts

	// wait for mc admin info
	err := <-doneCh
	if err != nil {
		return nil, err
	}

	return sessionResp, nil
}
