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

package subnet

import (
	"archive/zip"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"

	xhttp "github.com/minio/console/pkg/http"

	"github.com/minio/madmin-go/v2"
	"github.com/minio/mc/cmd"
	mc "github.com/minio/mc/cmd"
	"github.com/minio/pkg/env"
)

const (
	subnetRespBodyLimit = 1 << 20 // 1 MiB
)

func subnetBaseURL() string {
	return env.Get(ConsoleSubnetURL, "https://subnet.min.io")
}

func subnetRegisterURL() string {
	return subnetBaseURL() + "/api/cluster/register"
}

func subnetLoginURL() string {
	return subnetBaseURL() + "/api/auth/login"
}

func subnetOrgsURL() string {
	return subnetBaseURL() + "/api/auth/organizations"
}

func subnetMFAURL() string {
	return subnetBaseURL() + "/api/auth/mfa-login"
}

func subnetAPIKeyURL() string {
	return subnetBaseURL() + "/api/auth/api-key"
}

func LogWebhookURL() string {
	return subnetBaseURL() + "/api/logs"
}

func UploadURL(uploadType string, filename string) string {
	return fmt.Sprintf("%s/api/%s/upload?filename=%s", subnetBaseURL(), uploadType, filename)
}

func UploadAuthHeaders(apiKey string) map[string]string {
	return map[string]string{"x-subnet-api-key": apiKey}
}

/*
	func UploadFileToSubnet(info []interface{}, client *xhttp.Client, filename string, reqURL string, headers map[string]string) (string, error) {
		fmt.Println("In UploadFiletoSubnet - this should br perfinfo -  info[0]: ", info[0])
		fmt.Println("In UploadFiletoSubnet - this should be cluster info - info[1]: ", info[1])
		infoBody, e := zipResult(info, filename)
		req, e := subnetUploadReq(info, reqURL, filename)
		if e != nil {
			return "", e
		}
		resp, e := subnetReqDo(client, req, headers)
		return resp, e
	}
*/
func UploadFileToSubnet(info interface{}, client *xhttp.Client, filename string, reqURL string, headers map[string]string) (string, error) {
	infoBytes, e := json.Marshal(info)
	r, w := io.Pipe()
	w.Write(infoBytes)
	req, e := subnetUploadReq(r, reqURL, filename)
	if e != nil {
		return "", e
	}
	resp, e := subnetReqDo(client, req, headers)
	return resp, e
}

type SubnetPerfUploadInfo struct {
	ClusterInfo   cmd.ClusterRegistrationInfo `json:"cluster_info,omitempty"`
	SpeedtestInfo madmin.SpeedTestResult      `json:"speedtest_info,omitempty"`
}

func UploadPerfFileToSubnet(info SubnetPerfUploadInfo, client *xhttp.Client, filename string, reqURL string, headers map[string]string) (string, error) {
	fmt.Println("In UploadFiletoSubnet - info: ", info)
	infoBody, e := zipPerfResult(info, filename)
	if e != nil {
		fmt.Println("THERE WAS AN ERROR  ", e, "infoBody: ", infoBody)
		if e.Error() != "EOF " {
			return "", e
		}
		fmt.Println("I should print if it's an EOF error! e: ", e)
	}
	fmt.Println("we made it! infoBody: ", infoBody)
	req, e := subnetUploadReq(infoBody, reqURL, filename)
	if e != nil {
		fmt.Println("this is an error you have")
		return "", e
	}
	fmt.Println("we made it out of subnetUploadReq! req.Body: ", req.Body)
	if e != nil {
		return "", e
	}
	resp, e := subnetReqDo(client, req, headers)
	return resp, e
}

func convertObjStatServers(ss []madmin.SpeedTestStatServer) []mc.ObjStatServer {
	out := []mc.ObjStatServer{}
	for _, s := range ss {
		out = append(out, mc.ObjStatServer{
			Endpoint: s.Endpoint,
			Perf: mc.ObjStats{
				Throughput:    s.ThroughputPerSec,
				ObjectsPerSec: s.ObjectsPerSec,
			},
			Error: s.Err,
		})
	}
	return out
}

func convertPUTStats(stats madmin.SpeedTestStats) mc.ObjPUTStats {
	return mc.ObjPUTStats{
		Throughput:    stats.ThroughputPerSec,
		ObjectsPerSec: stats.ObjectsPerSec,
		Response:      stats.Response,
	}
}

func writeJSONObjToZip(zipWriter *zip.Writer, obj interface{}, filename string) error {
	writer, e := zipWriter.Create(filename)
	if e != nil {
		return e
	}

	return json.NewEncoder(writer).Encode(obj)
}
func convertPUTResults(stats madmin.SpeedTestStats) mc.ObjPUTPerfResults {
	fmt.Println("in convertPUTResults")
	return mc.ObjPUTPerfResults{
		Perf:    convertPUTStats(stats),
		Servers: convertObjStatServers(stats.Servers),
	}
}

func convertGETResults(stats madmin.SpeedTestStats) mc.ObjGETPerfResults {
	fmt.Println("in convertgetResults  ")
	return mc.ObjGETPerfResults{
		Perf: mc.ObjGETStats{
			ObjPUTStats: convertPUTStats(stats),
			TTFB:        stats.TTFB,
		},
		Servers: convertObjStatServers(stats.Servers),
	}
}
func convertObjTestResults(objResult *madmin.SpeedTestResult) mc.ObjTestResults {
	fmt.Println("in convertobjTestResults objResult: ", objResult)
	//if objResult == nil {
	//	return nil
	//}
	result := mc.ObjTestResults{
		ObjectSize: objResult.Size,
		Threads:    objResult.Concurrent,
	}
	result.PUTResults = convertPUTResults(objResult.PUTStats)
	result.GETResults = convertGETResults(objResult.GETStats)
	fmt.Println("got all the stuuf result: ", result)
	return result
}

func convertPerfResult(r mc.PerfTestResult) mc.PerfTestOutput {
	var out mc.PerfTestOutput
	thisThing := convertObjTestResults(r.ObjectResult)

	out.ObjectResults = &thisThing
	return out
}

func zipPerfResult(perfOutput SubnetPerfUploadInfo, resultFilename string) (*io.PipeReader, error) {
	fmt.Println("in zipPerfResult")
	var perfTestSubnetReport cmd.PerfTestResult
	perfTestSubnetReport.ObjectResult = &perfOutput.SpeedtestInfo

	perfTestOutput := convertPerfResult(perfTestSubnetReport)
	fmt.Println("we got out! perfTestOutput.Objectresults: ", perfTestOutput.ObjectResults)
	r, w := io.Pipe()

	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()
	e := writeJSONObjToZip(zipWriter, perfTestOutput.ObjectResults, resultFilename)
	if e != nil {
		return nil, e
	}
	e = writeJSONObjToZip(zipWriter, perfOutput.ClusterInfo, "cluster.info")
	if e != nil {
		return nil, e
	}
	//w.Close()

	fmt.Println("you finished this part! good job! r: ", r)
	//fmt.Println("make a byte[]")
	//var bytesThing []byte

	//fmt.Println("read into it")
	//_, e = r.Read(bytesThing)
	//if e != nil {
	//	fmt.Println("THIS IS THE ERROR!")
	//	if e.Error() != "EOF " {
	//		return nil, e
	//	}
	//}
	w.Close()
	fmt.Println("try to return")
	return r, nil
}

func subnetUploadReq(info *io.PipeReader, url string, filename string) (*http.Request, error) {

	fmt.Println("In subnetUploadReq")
	r, w := io.Pipe()
	mwriter := multipart.NewWriter(w)
	contentType := mwriter.FormDataContentType()

	defer func() {
		mwriter.Close()
		w.Close()
	}()

	//file, e := os.Open(filename)
	//if e != nil {
	//	return
	//}
	//defer file.Close()
	//_, e = w.Write(info)
	_, e := io.Copy(w, info)
	fmt.Println("I have a thing! r: ", r, " info: ", info)
	req, e := http.NewRequest(http.MethodPost, url, r)
	if e != nil {
		fmt.Println("There was an error here! e: ", e)

	}
	fmt.Println("There was no error here! ")
	req.Header.Add("Content-Type", contentType)
	fmt.Println("made the header! req.Body: ", req.Body)
	testR, testW := io.Pipe()
	b := make([]byte, 50)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}
	testW.Write(b)
	testReq, e := http.NewRequest(http.MethodGet, url, testR)
	return testReq, nil
}

//_, e := io.Copy(w, info)
//fmt.Println("I have no part! r: ", r)
//req, e := http.NewRequest(http.MethodPost, url, r)
//if e != nil {
//	fmt.Println("There was an error here! e: ", e)

//}
//req.Header.Add("Content-Type", contentType)
//fmt.Println("in console perf &req.Body: ", req.Body)
//return req, nil
//tmpArchive, e := os.CreateTemp("", "test-perf-")
//if e != nil {
//	return nil, e
//}
//defer tmpArchive.Close()

//var body bytes.Buffer
//r, w := io.Pipe()
//fmt.Println("made the pipe!")
//writer := multipart.NewWriter(w)
//fmt.Println("made the writer!")
//zipWriter := gzip.NewWriter(w)
//fmt.Println("made the zipwriter! I think something is wrong here!")
//zipFileWriter := gzip.NewWriter(tmpArchive)
//	version := "3"
//fmt.Println("encode it!")
//enc := json.NewEncoder(zipWriter)
//fileEnc := json.NewEncoder(zipFileWriter)

//	header := struct {
//		Version string `json:"version"`
//	}{Version: version}
//	fmt.Println("encode the header! header: ", header)
//	if e := enc.Encode(header); e != nil {
//	return nil, e
//	}
//fmt.Println("do the loop!")
//for i := range info {
//
//	if e := enc.Encode(info[i]); e != nil {
//			return nil, e
//		}
//}
//if e := fileEnc.Encode(header); e != nil {
//	return nil, e
//}
//for i := range info {
//	if e := fileEnc.Encode(info[i]); e != nil {
//		return nil, e
//	}
//}
//zipWriter.Close()
//zipFileWriter.Close()
//	temp := r
//	fmt.Println("make a part!")
//	part, e := writer.CreateFormFile("file", filename)
//if e != nil {
//		return nil, e
//	}
//	if _, e = io.Copy(part, temp); e != nil {
//		return nil, e
//	}

// writer.Close()
// fmt.Println("making the request body: ", r)
// req, e := http.NewRequest(http.MethodPost, url, r)
//
//	if e != nil {
//		return nil, e
//	}
//
// req.Header.Add("Content-Type", writer.FormDataContentType())
// fmt.Println("in request req.body: ", req.Body)
// return req, nil

func GenerateRegToken(clusterRegInfo mc.ClusterRegistrationInfo) (string, error) {
	token, e := json.Marshal(clusterRegInfo)
	if e != nil {
		return "", e
	}

	return base64.StdEncoding.EncodeToString(token), nil
}

func subnetAuthHeaders(authToken string) map[string]string {
	return map[string]string{"Authorization": "Bearer " + authToken}
}

func httpDo(client xhttp.ClientI, req *http.Request) (*http.Response, error) {
	return client.Do(req)
}

func subnetReqDo(client xhttp.ClientI, r *http.Request, headers map[string]string) (string, error) {
	for k, v := range headers {
		r.Header.Add(k, v)
	}

	ct := r.Header.Get("Content-Type")
	if len(ct) == 0 {
		r.Header.Add("Content-Type", "application/json")
	}

	resp, e := httpDo(client, r)
	if e != nil {
		return "", e
	}

	defer resp.Body.Close()
	respBytes, e := ioutil.ReadAll(io.LimitReader(resp.Body, subnetRespBodyLimit))
	if e != nil {
		return "", e
	}
	respStr := string(respBytes)

	if resp.StatusCode == http.StatusOK {
		return respStr, nil
	}
	return respStr, fmt.Errorf("Request failed with code %d and errors: %s", resp.StatusCode, respStr)
}

func subnetGetReq(client xhttp.ClientI, reqURL string, headers map[string]string) (string, error) {
	r, e := http.NewRequest(http.MethodGet, reqURL, nil)
	if e != nil {
		return "", e
	}
	return subnetReqDo(client, r, headers)
}

func subnetPostReq(client xhttp.ClientI, reqURL string, payload interface{}, headers map[string]string) (string, error) {
	body, e := json.Marshal(payload)
	if e != nil {
		return "", e
	}
	r, e := http.NewRequest(http.MethodPost, reqURL, bytes.NewReader(body))
	if e != nil {
		return "", e
	}
	return subnetReqDo(client, r, headers)
}

func GetClusterRegInfo(admInfo madmin.InfoMessage) mc.ClusterRegistrationInfo {
	noOfPools := 1
	noOfDrives := 0
	for _, srvr := range admInfo.Servers {
		if srvr.PoolNumber > noOfPools {
			noOfPools = srvr.PoolNumber
		}
		noOfDrives += len(srvr.Disks)
	}

	totalSpace, usedSpace := getDriveSpaceInfo(admInfo)

	return mc.ClusterRegistrationInfo{
		DeploymentID: admInfo.DeploymentID,
		ClusterName:  admInfo.DeploymentID,
		UsedCapacity: admInfo.Usage.Size,
		Info: mc.ClusterInfo{
			MinioVersion:    admInfo.Servers[0].Version,
			NoOfServerPools: noOfPools,
			NoOfServers:     len(admInfo.Servers),
			NoOfDrives:      noOfDrives,
			TotalDriveSpace: totalSpace,
			UsedDriveSpace:  usedSpace,
			NoOfBuckets:     admInfo.Buckets.Count,
			NoOfObjects:     admInfo.Objects.Count,
		},
	}
}

func getDriveSpaceInfo(admInfo madmin.InfoMessage) (uint64, uint64) {
	total := uint64(0)
	used := uint64(0)
	for _, srvr := range admInfo.Servers {
		for _, d := range srvr.Disks {
			total += d.TotalSpace
			used += d.UsedSpace
		}
	}
	return total, used
}

// ClusterRegistrationInfo - Information stored in the cluster registration token
type ClusterRegistrationInfo struct {
	DeploymentID string      `json:"deployment_id"`
	ClusterName  string      `json:"cluster_name"`
	UsedCapacity uint64      `json:"used_capacity"`
	Info         ClusterInfo `json:"info"`
}

// ClusterInfo - The "info" sub-node of the cluster registration information struct
// Intended to be extensible i.e. more fields will be added as and when required
type ClusterInfo struct {
	MinioVersion    string `json:"minio_version"`
	NoOfServerPools int    `json:"no_of_server_pools"`
	NoOfServers     int    `json:"no_of_servers"`
	NoOfDrives      int    `json:"no_of_drives"`
	NoOfBuckets     uint64 `json:"no_of_buckets"`
	NoOfObjects     uint64 `json:"no_of_objects"`
	TotalDriveSpace uint64 `json:"total_drive_space"`
	UsedDriveSpace  uint64 `json:"used_drive_space"`
}

func getClusterRegInfo(admInfo madmin.InfoMessage, clusterName string) ClusterRegistrationInfo {
	noOfPools := 1
	noOfDrives := 0
	for _, srvr := range admInfo.Servers {
		if srvr.PoolNumber > noOfPools {
			noOfPools = srvr.PoolNumber
		}
		noOfDrives += len(srvr.Disks)
	}

	totalSpace, usedSpace := getDriveSpaceInfo(admInfo)

	return ClusterRegistrationInfo{
		DeploymentID: admInfo.DeploymentID,
		ClusterName:  clusterName,
		UsedCapacity: admInfo.Usage.Size,
		Info: ClusterInfo{
			MinioVersion:    admInfo.Servers[0].Version,
			NoOfServerPools: noOfPools,
			NoOfServers:     len(admInfo.Servers),
			NoOfDrives:      noOfDrives,
			TotalDriveSpace: totalSpace,
			UsedDriveSpace:  usedSpace,
			NoOfBuckets:     admInfo.Buckets.Count,
			NoOfObjects:     admInfo.Objects.Count,
		},
	}
}
