/*
 * MinIO Cloud Storage, (C) 2018 MinIO, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package audit

import (
	"net/http"
	"strings"
	"time"

	xhttp "minio/cmd/http"
	"minio/pkg/handlers"
)

// Version - represents the current version of audit log structure.
const Version = "1"

// Entry - audit entry logs.
type Entry struct {
	Version      string `json:"version"`
	DeploymentID string `json:"deploymentid,omitempty"`
	Time         string `json:"time"`
	Trigger      string `json:"trigger"`
	API          struct {
		Name            string `json:"name,omitempty"`
		Bucket          string `json:"bucket,omitempty"`
		Object          string `json:"object,omitempty"`
		Status          string `json:"status,omitempty"`
		StatusCode      int    `json:"statusCode,omitempty"`
		TimeToFirstByte string `json:"timeToFirstByte,omitempty"`
		TimeToResponse  string `json:"timeToResponse,omitempty"`
	} `json:"api"`
	RemoteHost string                 `json:"remotehost,omitempty"`
	RequestID  string                 `json:"requestID,omitempty"`
	UserAgent  string                 `json:"userAgent,omitempty"`
	ReqClaims  map[string]interface{} `json:"requestClaims,omitempty"`
	ReqQuery   map[string]string      `json:"requestQuery,omitempty"`
	ReqHeader  map[string]string      `json:"requestHeader,omitempty"`
	RespHeader map[string]string      `json:"responseHeader,omitempty"`
	Tags       map[string]interface{} `json:"tags,omitempty"`
}

// NewEntry - constructs an audit entry object with some fields filled
func NewEntry(deploymentID string) Entry {
	return Entry{
		Version:      Version,
		DeploymentID: deploymentID,
		Time:         time.Now().UTC().Format(time.RFC3339Nano),
	}
}

// ToEntry - constructs an audit entry from a http request
func ToEntry(w http.ResponseWriter, r *http.Request, reqClaims map[string]interface{}, deploymentID string) Entry {

	entry := NewEntry(deploymentID)

	entry.RemoteHost = handlers.GetSourceIP(r)
	entry.UserAgent = r.UserAgent()
	entry.ReqClaims = reqClaims

	q := r.URL.Query()
	reqQuery := make(map[string]string, len(q))
	for k, v := range q {
		reqQuery[k] = strings.Join(v, ",")
	}
	entry.ReqQuery = reqQuery

	reqHeader := make(map[string]string, len(r.Header))
	for k, v := range r.Header {
		reqHeader[k] = strings.Join(v, ",")
	}
	entry.ReqHeader = reqHeader

	wh := w.Header()
	entry.RequestID = wh.Get(xhttp.AmzRequestID)
	respHeader := make(map[string]string, len(wh))
	for k, v := range wh {
		respHeader[k] = strings.Join(v, ",")
	}
	entry.RespHeader = respHeader

	if etag := respHeader[xhttp.ETag]; etag != "" {
		respHeader[xhttp.ETag] = strings.Trim(etag, `"`)
	}

	return entry
}
