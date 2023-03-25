/*
 * MinIO Cloud Storage, (C) 2019 MinIO, Inc.
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

package cmd

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/tinylib/msgp/msgp"

	"minio/cmd/logger"
	b "minio/pkg/bucket/bandwidth"
	"minio/pkg/event"
	"minio/pkg/madmin"
	trace "minio/pkg/trace"
)

// To abstract a node over network.
type peerRESTServer struct{}

// GetLocksHandler - returns list of older lock from the server.
func (s *peerRESTServer) GetLocksHandler(w http.ResponseWriter, r *http.Request) {
	if !s.IsValid(w, r) {
		s.WriteErrorResponse(w, errors.New("Invalid request"))
		return
	}

	ctx := NewContext(r, w, "GetLocks")
	logger.LogIf(ctx, gob.NewEncoder(w).Encode(globalLockServer.DupLockMap()))

	w.(http.Flusher).Flush()

}

// DeletePolicyHandler - deletes a policy on the server.
func (s *peerRESTServer) DeletePolicyHandler(w http.ResponseWriter, r *http.Request) {
	if !s.IsValid(w, r) {
		s.WriteErrorResponse(w, errors.New("Invalid request"))
		return
	}

	objAPI := newObjectLayerFn()
	if objAPI == nil {
		s.WriteErrorResponse(w, errServerNotInitialized)
		return
	}

	vars := mux.Vars(r)
	policyName := vars[peerRESTPolicy]
	if policyName == "" {
		s.WriteErrorResponse(w, errors.New("policyName is missing"))
		return
	}

	if err := GlobalIAMSys.DeletePolicy(policyName); err != nil {
		s.WriteErrorResponse(w, err)
		return
	}

	w.(http.Flusher).Flush()
}

// LoadPolicyHandler - reloads a policy on the server.
func (s *peerRESTServer) LoadPolicyHandler(w http.ResponseWriter, r *http.Request) {
	if !s.IsValid(w, r) {
		s.WriteErrorResponse(w, errors.New("Invalid request"))
		return
	}

	objAPI := newObjectLayerFn()
	if objAPI == nil {
		s.WriteErrorResponse(w, errServerNotInitialized)
		return
	}

	vars := mux.Vars(r)
	policyName := vars[peerRESTPolicy]
	if policyName == "" {
		s.WriteErrorResponse(w, errors.New("policyName is missing"))
		return
	}

	if err := GlobalIAMSys.LoadPolicy(objAPI, policyName); err != nil {
		s.WriteErrorResponse(w, err)
		return
	}

	w.(http.Flusher).Flush()
}

// LoadPolicyMappingHandler - reloads a policy mapping on the server.
func (s *peerRESTServer) LoadPolicyMappingHandler(w http.ResponseWriter, r *http.Request) {
	if !s.IsValid(w, r) {
		s.WriteErrorResponse(w, errors.New("Invalid request"))
		return
	}

	objAPI := newObjectLayerFn()
	if objAPI == nil {
		s.WriteErrorResponse(w, errServerNotInitialized)
		return
	}

	vars := mux.Vars(r)
	userOrGroup := vars[peerRESTUserOrGroup]
	if userOrGroup == "" {
		s.WriteErrorResponse(w, errors.New("user-or-group is missing"))
		return
	}
	_, isGroup := vars[peerRESTIsGroup]

	if err := GlobalIAMSys.LoadPolicyMapping(objAPI, userOrGroup, isGroup); err != nil {
		s.WriteErrorResponse(w, err)
		return
	}

	w.(http.Flusher).Flush()
}

// DeleteServiceAccountHandler - deletes a service account on the server.
func (s *peerRESTServer) DeleteServiceAccountHandler(w http.ResponseWriter, r *http.Request) {
	if !s.IsValid(w, r) {
		s.WriteErrorResponse(w, errors.New("Invalid request"))
		return
	}

	objAPI := newObjectLayerFn()
	if objAPI == nil {
		s.WriteErrorResponse(w, errServerNotInitialized)
		return
	}

	vars := mux.Vars(r)
	accessKey := vars[peerRESTUser]
	if accessKey == "" {
		s.WriteErrorResponse(w, errors.New("service account name is missing"))
		return
	}

	if err := GlobalIAMSys.DeleteServiceAccount(r.Context(), accessKey); err != nil {
		s.WriteErrorResponse(w, err)
		return
	}

	w.(http.Flusher).Flush()
}

// LoadServiceAccountHandler - reloads a service account on the server.
func (s *peerRESTServer) LoadServiceAccountHandler(w http.ResponseWriter, r *http.Request) {
	if !s.IsValid(w, r) {
		s.WriteErrorResponse(w, errors.New("Invalid request"))
		return
	}

	objAPI := newObjectLayerFn()
	if objAPI == nil {
		s.WriteErrorResponse(w, errServerNotInitialized)
		return
	}

	vars := mux.Vars(r)
	accessKey := vars[peerRESTUser]
	if accessKey == "" {
		s.WriteErrorResponse(w, errors.New("service account parameter is missing"))
		return
	}

	if err := GlobalIAMSys.LoadServiceAccount(accessKey); err != nil {
		s.WriteErrorResponse(w, err)
		return
	}

	w.(http.Flusher).Flush()
}

// DeleteUserHandler - deletes a user on the server.
func (s *peerRESTServer) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	if !s.IsValid(w, r) {
		s.WriteErrorResponse(w, errors.New("Invalid request"))
		return
	}

	objAPI := newObjectLayerFn()
	if objAPI == nil {
		s.WriteErrorResponse(w, errServerNotInitialized)
		return
	}

	vars := mux.Vars(r)
	accessKey := vars[peerRESTUser]
	if accessKey == "" {
		s.WriteErrorResponse(w, errors.New("username is missing"))
		return
	}

	if err := GlobalIAMSys.DeleteUser(accessKey); err != nil {
		s.WriteErrorResponse(w, err)
		return
	}

	w.(http.Flusher).Flush()
}

// LoadUserHandler - reloads a user on the server.
func (s *peerRESTServer) LoadUserHandler(w http.ResponseWriter, r *http.Request) {
	if !s.IsValid(w, r) {
		s.WriteErrorResponse(w, errors.New("Invalid request"))
		return
	}

	objAPI := newObjectLayerFn()
	if objAPI == nil {
		s.WriteErrorResponse(w, errServerNotInitialized)
		return
	}

	vars := mux.Vars(r)
	accessKey := vars[peerRESTUser]
	if accessKey == "" {
		s.WriteErrorResponse(w, errors.New("username is missing"))
		return
	}

	temp, err := strconv.ParseBool(vars[peerRESTUserTemp])
	if err != nil {
		s.WriteErrorResponse(w, err)
		return
	}

	var userType = regularUser
	if temp {
		userType = stsUser
	}

	if err = GlobalIAMSys.LoadUser(objAPI, accessKey, userType); err != nil {
		s.WriteErrorResponse(w, err)
		return
	}

	w.(http.Flusher).Flush()
}

// LoadGroupHandler - reloads group along with members list.
func (s *peerRESTServer) LoadGroupHandler(w http.ResponseWriter, r *http.Request) {
	if !s.IsValid(w, r) {
		s.WriteErrorResponse(w, errors.New("Invalid request"))
		return
	}

	objAPI := newObjectLayerFn()
	if objAPI == nil {
		s.WriteErrorResponse(w, errServerNotInitialized)
		return
	}

	vars := mux.Vars(r)
	group := vars[peerRESTGroup]
	err := GlobalIAMSys.LoadGroup(objAPI, group)
	if err != nil {
		s.WriteErrorResponse(w, err)
		return
	}

	w.(http.Flusher).Flush()
}

// StartProfilingHandler - Issues the start profiling command.
func (s *peerRESTServer) StartProfilingHandler(w http.ResponseWriter, r *http.Request) {
	if !s.IsValid(w, r) {
		s.WriteErrorResponse(w, errors.New("Invalid request"))
		return
	}

	vars := mux.Vars(r)
	profiles := strings.Split(vars[peerRESTProfiler], ",")
	if len(profiles) == 0 {
		s.WriteErrorResponse(w, errors.New("profiler name is missing"))
		return
	}
	globalProfilerMu.Lock()
	defer globalProfilerMu.Unlock()
	if globalProfiler == nil {
		globalProfiler = make(map[string]minioProfiler, 10)
	}

	// Stop profiler of all types if already running
	for k, v := range globalProfiler {
		for _, p := range profiles {
			if p == k {
				v.Stop()
				delete(globalProfiler, k)
			}
		}
	}

	for _, profiler := range profiles {
		prof, err := startProfiler(profiler)
		if err != nil {
			s.WriteErrorResponse(w, err)
			return
		}
		globalProfiler[profiler] = prof
	}

	w.(http.Flusher).Flush()
}

// DownloadProfilingDataHandler - returns profiled data.
func (s *peerRESTServer) DownloadProfilingDataHandler(w http.ResponseWriter, r *http.Request) {
	if !s.IsValid(w, r) {
		s.WriteErrorResponse(w, errors.New("Invalid request"))
		return
	}

	ctx := NewContext(r, w, "DownloadProfiling")
	profileData, err := getProfileData()
	if err != nil {
		s.WriteErrorResponse(w, err)
		return
	}

	defer w.(http.Flusher).Flush()
	logger.LogIf(ctx, gob.NewEncoder(w).Encode(profileData))
}

// ServerInfoHandler - returns Server Info
func (s *peerRESTServer) ServerInfoHandler(w http.ResponseWriter, r *http.Request) {
	if !s.IsValid(w, r) {
		s.WriteErrorResponse(w, errors.New("Invalid request"))
		return
	}

	ctx := NewContext(r, w, "ServerInfo")
	info := getLocalServerProperty(globalEndpoints, r)

	defer w.(http.Flusher).Flush()
	logger.LogIf(ctx, gob.NewEncoder(w).Encode(info))
}

func (s *peerRESTServer) NetInfoHandler(w http.ResponseWriter, r *http.Request) {
	ctx := NewContext(r, w, "NetInfo")
	if !s.IsValid(w, r) {
		s.WriteErrorResponse(w, errors.New("Invalid request"))
		return
	}

	// Use this trailer to send additional headers after sending body
	w.Header().Set("Trailer", "FinalStatus")

	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)

	n, err := io.Copy(ioutil.Discard, r.Body)
	if err == io.ErrUnexpectedEOF {
		w.Header().Set("FinalStatus", err.Error())
		return
	}
	if err != nil && err != io.EOF {
		logger.LogIf(ctx, err)
		w.Header().Set("FinalStatus", err.Error())
		return
	}
	if n != r.ContentLength {
		err := fmt.Errorf("Subnet health: short read: expected %d found %d", r.ContentLength, n)

		logger.LogIf(ctx, err)
		w.Header().Set("FinalStatus", err.Error())
		return
	}
	w.Header().Set("FinalStatus", "Success")
	w.(http.Flusher).Flush()
}

func (s *peerRESTServer) DispatchNetInfoHandler(w http.ResponseWriter, r *http.Request) {
	if !s.IsValid(w, r) {
		s.WriteErrorResponse(w, errors.New("Invalid request"))
		return
	}

	done := keepHTTPResponseAlive(w)

	ctx := NewContext(r, w, "DispatchNetInfo")
	info := GlobalNotificationSys.NetInfo(ctx)

	done(nil)
	logger.LogIf(ctx, gob.NewEncoder(w).Encode(info))
	w.(http.Flusher).Flush()
}

// DriveInfoHandler - returns Drive info.
func (s *peerRESTServer) DriveInfoHandler(w http.ResponseWriter, r *http.Request) {
	if !s.IsValid(w, r) {
		s.WriteErrorResponse(w, errors.New("Invalid request"))
		return
	}

	ctx, cancel := context.WithCancel(NewContext(r, w, "DriveInfo"))
	defer cancel()
	infoSerial := getLocalDrives(ctx, false, globalEndpoints, r)
	infoParallel := getLocalDrives(ctx, true, globalEndpoints, r)

	errStr := ""
	if infoSerial.Error != "" {
		errStr = "serial: " + infoSerial.Error
	}
	if infoParallel.Error != "" {
		errStr = errStr + " parallel: " + infoParallel.Error
	}
	info := madmin.ServerDrivesInfo{
		Addr:     infoSerial.Addr,
		Serial:   infoSerial.Serial,
		Parallel: infoParallel.Parallel,
		Error:    errStr,
	}
	defer w.(http.Flusher).Flush()
	logger.LogIf(ctx, gob.NewEncoder(w).Encode(info))
}

// CPUInfoHandler - returns CPU info.
func (s *peerRESTServer) CPUInfoHandler(w http.ResponseWriter, r *http.Request) {
	if !s.IsValid(w, r) {
		s.WriteErrorResponse(w, errors.New("Invalid request"))
		return
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	info := getLocalCPUInfo(ctx, r)

	defer w.(http.Flusher).Flush()
	logger.LogIf(ctx, gob.NewEncoder(w).Encode(info))
}

// DiskHwInfoHandler - returns Disk HW info.
func (s *peerRESTServer) DiskHwInfoHandler(w http.ResponseWriter, r *http.Request) {
	if !s.IsValid(w, r) {
		s.WriteErrorResponse(w, errors.New("Invalid request"))
		return
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	info := getLocalDiskHwInfo(ctx, r)

	defer w.(http.Flusher).Flush()
	logger.LogIf(ctx, gob.NewEncoder(w).Encode(info))
}

// OsInfoHandler - returns Os info.
func (s *peerRESTServer) OsInfoHandler(w http.ResponseWriter, r *http.Request) {
	if !s.IsValid(w, r) {
		s.WriteErrorResponse(w, errors.New("Invalid request"))
		return
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	info := getLocalOsInfo(ctx, r)

	defer w.(http.Flusher).Flush()
	logger.LogIf(ctx, gob.NewEncoder(w).Encode(info))
}

// ProcInfoHandler - returns Proc info.
func (s *peerRESTServer) ProcInfoHandler(w http.ResponseWriter, r *http.Request) {
	if !s.IsValid(w, r) {
		s.WriteErrorResponse(w, errors.New("Invalid request"))
		return
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	info := getLocalProcInfo(ctx, r)

	defer w.(http.Flusher).Flush()
	logger.LogIf(ctx, gob.NewEncoder(w).Encode(info))
}

// MemInfoHandler - returns Memory info.
func (s *peerRESTServer) MemInfoHandler(w http.ResponseWriter, r *http.Request) {
	if !s.IsValid(w, r) {
		s.WriteErrorResponse(w, errors.New("Invalid request"))
		return
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	info := getLocalMemInfo(ctx, r)

	defer w.(http.Flusher).Flush()
	logger.LogIf(ctx, gob.NewEncoder(w).Encode(info))
}

// DeleteBucketMetadataHandler - Delete in memory bucket metadata
func (s *peerRESTServer) DeleteBucketMetadataHandler(w http.ResponseWriter, r *http.Request) {
	if !s.IsValid(w, r) {
		s.WriteErrorResponse(w, errors.New("Invalid request"))
		return
	}

	vars := mux.Vars(r)
	bucketName := vars[peerRESTBucket]
	if bucketName == "" {
		s.WriteErrorResponse(w, errors.New("Bucket name is missing"))
		return
	}

	globalReplicationStats.Delete(bucketName)
	globalBucketMetadataSys.Remove(bucketName)
	if localMetacacheMgr != nil {
		localMetacacheMgr.deleteBucketCache(bucketName)
	}
}

// GetBucketStatsHandler - fetches current in-memory bucket stats, currently only
// returns BucketReplicationStatus
func (s *peerRESTServer) GetBucketStatsHandler(w http.ResponseWriter, r *http.Request) {
	if !s.IsValid(w, r) {
		s.WriteErrorResponse(w, errors.New("Invalid request"))
		return
	}

	vars := mux.Vars(r)
	bucketName := vars[peerRESTBucket]
	if bucketName == "" {
		s.WriteErrorResponse(w, errors.New("Bucket name is missing"))
		return
	}

	bs := BucketStats{
		ReplicationStats: globalReplicationStats.Get(bucketName),
	}
	defer w.(http.Flusher).Flush()
	logger.LogIf(r.Context(), msgp.Encode(w, &bs))
}

// LoadBucketMetadataHandler - reloads in memory bucket metadata
func (s *peerRESTServer) LoadBucketMetadataHandler(w http.ResponseWriter, r *http.Request) {
	if !s.IsValid(w, r) {
		s.WriteErrorResponse(w, errors.New("Invalid request"))
		return
	}

	vars := mux.Vars(r)
	bucketName := vars[peerRESTBucket]
	if bucketName == "" {
		s.WriteErrorResponse(w, errors.New("Bucket name is missing"))
		return
	}

	objAPI := newObjectLayerFn()
	if objAPI == nil {
		s.WriteErrorResponse(w, errServerNotInitialized)
		return
	}

	meta, err := loadBucketMetadata(r.Context(), objAPI, bucketName)
	if err != nil {
		s.WriteErrorResponse(w, err)
		return
	}

	globalBucketMetadataSys.Set(bucketName, meta)

	if meta.notificationConfig != nil {
		GlobalNotificationSys.AddRulesMap(bucketName, meta.notificationConfig.ToRulesMap())
	}

	if meta.bucketTargetConfig != nil {
		globalBucketTargetSys.UpdateAllTargets(bucketName, meta.bucketTargetConfig)
	}
}

// CycleServerBloomFilterHandler cycles bloom filter on server.
func (s *peerRESTServer) CycleServerBloomFilterHandler(w http.ResponseWriter, r *http.Request) {
	if !s.IsValid(w, r) {
		s.WriteErrorResponse(w, errors.New("Invalid request"))
		return
	}

	ctx := NewContext(r, w, "CycleServerBloomFilter")

	var req bloomFilterRequest
	err := gob.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		s.WriteErrorResponse(w, err)
		return
	}
	bf, err := intDataUpdateTracker.cycleFilter(ctx, req)
	if err != nil {
		s.WriteErrorResponse(w, err)
		return
	}

	logger.LogIf(ctx, gob.NewEncoder(w).Encode(bf))
}

func (s *peerRESTServer) GetMetacacheListingHandler(w http.ResponseWriter, r *http.Request) {
	if !s.IsValid(w, r) {
		s.WriteErrorResponse(w, errors.New("Invalid request"))
		return
	}
	ctx := NewContext(r, w, "GetMetacacheListing")

	var opts listPathOptions
	err := gob.NewDecoder(r.Body).Decode(&opts)
	if err != nil && err != io.EOF {
		s.WriteErrorResponse(w, err)
		return
	}
	resp := localMetacacheMgr.getBucket(ctx, opts.Bucket).findCache(opts)
	logger.LogIf(ctx, msgp.Encode(w, &resp))
}

func (s *peerRESTServer) UpdateMetacacheListingHandler(w http.ResponseWriter, r *http.Request) {
	if !s.IsValid(w, r) {
		s.WriteErrorResponse(w, errors.New("Invalid request"))
		return
	}
	ctx := NewContext(r, w, "UpdateMetacacheListing")

	var req metacache
	err := msgp.Decode(r.Body, &req)
	if err != nil {
		s.WriteErrorResponse(w, err)
		return
	}
	cache, err := localMetacacheMgr.updateCacheEntry(req)
	if err != nil {
		s.WriteErrorResponse(w, err)
		return
	}
	// Return updated metadata.
	logger.LogIf(ctx, msgp.Encode(w, &cache))
}

// PutBucketNotificationHandler - Set bucket policy.
func (s *peerRESTServer) PutBucketNotificationHandler(w http.ResponseWriter, r *http.Request) {
	if !s.IsValid(w, r) {
		s.WriteErrorResponse(w, errors.New("Invalid request"))
		return
	}

	vars := mux.Vars(r)
	bucketName := vars[peerRESTBucket]
	if bucketName == "" {
		s.WriteErrorResponse(w, errors.New("Bucket name is missing"))
		return
	}

	var rulesMap event.RulesMap
	if r.ContentLength < 0 {
		s.WriteErrorResponse(w, errInvalidArgument)
		return
	}

	err := gob.NewDecoder(r.Body).Decode(&rulesMap)
	if err != nil {
		s.WriteErrorResponse(w, err)
		return
	}

	GlobalNotificationSys.AddRulesMap(bucketName, rulesMap)
	w.(http.Flusher).Flush()
}

// Return disk IDs of all the local disks.
func getLocalDiskIDs(z *erasureServerPools) []string {
	var ids []string

	for poolIdx := range z.serverPools {
		for _, set := range z.serverPools[poolIdx].sets {
			disks := set.getDisks()
			for _, disk := range disks {
				if disk == nil {
					continue
				}
				if disk.IsLocal() {
					id, err := disk.GetDiskID()
					if err != nil {
						continue
					}
					if id == "" {
						continue
					}
					ids = append(ids, id)
				}
			}
		}
	}

	return ids
}

// HealthHandler - returns true of health
func (s *peerRESTServer) HealthHandler(w http.ResponseWriter, r *http.Request) {
	s.IsValid(w, r)
}

// GetLocalDiskIDs - Return disk IDs of all the local disks.
func (s *peerRESTServer) GetLocalDiskIDs(w http.ResponseWriter, r *http.Request) {
	if !s.IsValid(w, r) {
		s.WriteErrorResponse(w, errors.New("Invalid request"))
		return
	}

	ctx := NewContext(r, w, "GetLocalDiskIDs")

	objLayer := newObjectLayerFn()

	// Service not initialized yet
	if objLayer == nil {
		s.WriteErrorResponse(w, errServerNotInitialized)
		return
	}

	z, ok := objLayer.(*erasureServerPools)
	if !ok {
		s.WriteErrorResponse(w, errServerNotInitialized)
		return
	}

	ids := getLocalDiskIDs(z)
	logger.LogIf(ctx, gob.NewEncoder(w).Encode(ids))
	w.(http.Flusher).Flush()
}

// ServerUpdateHandler - updates the current server.
func (s *peerRESTServer) ServerUpdateHandler(w http.ResponseWriter, r *http.Request) {
	if !s.IsValid(w, r) {
		s.WriteErrorResponse(w, errors.New("Invalid request"))
		return
	}

	if r.ContentLength < 0 {
		s.WriteErrorResponse(w, errInvalidArgument)
		return
	}

	var info serverUpdateInfo
	err := gob.NewDecoder(r.Body).Decode(&info)
	if err != nil {
		s.WriteErrorResponse(w, err)
		return
	}

	if _, err = updateServer(info.URL, info.Sha256Sum, info.Time, info.ReleaseInfo, getMinioMode()); err != nil {
		s.WriteErrorResponse(w, err)
		return
	}
}

var errUnsupportedSignal = fmt.Errorf("unsupported signal: only restart and stop signals are supported")

// SignalServiceHandler - signal service handler.
func (s *peerRESTServer) SignalServiceHandler(w http.ResponseWriter, r *http.Request) {
	if !s.IsValid(w, r) {
		s.WriteErrorResponse(w, errors.New("Invalid request"))
		return
	}

	vars := mux.Vars(r)
	signalString := vars[peerRESTSignal]
	if signalString == "" {
		s.WriteErrorResponse(w, errors.New("signal name is missing"))
		return
	}
	si, err := strconv.Atoi(signalString)
	if err != nil {
		s.WriteErrorResponse(w, err)
		return
	}
	signal := serviceSignal(si)
	defer w.(http.Flusher).Flush()
	switch signal {
	case serviceRestart:
		globalServiceSignalCh <- signal
	case serviceStop:
		globalServiceSignalCh <- signal
	case serviceReloadDynamic:
		objAPI := newObjectLayerFn()
		if objAPI == nil {
			s.WriteErrorResponse(w, errServerNotInitialized)
			return
		}
		srvCfg, err := getValidConfig(objAPI)
		if err != nil {
			s.WriteErrorResponse(w, err)
			return
		}
		if err = applyDynamicConfig(r.Context(), objAPI, srvCfg); err != nil {
			s.WriteErrorResponse(w, err)
		}
		return
	default:
		s.WriteErrorResponse(w, errUnsupportedSignal)
		return
	}
}

// ListenHandler sends http trace messages back to peer rest client
func (s *peerRESTServer) ListenHandler(w http.ResponseWriter, r *http.Request) {
	if !s.IsValid(w, r) {
		s.WriteErrorResponse(w, errors.New("invalid request"))
		return
	}

	values := r.URL.Query()

	var prefix string
	if len(values[peerRESTListenPrefix]) > 1 {
		s.WriteErrorResponse(w, errors.New("invalid request"))
		return
	}

	if len(values[peerRESTListenPrefix]) == 1 {
		if err := event.ValidateFilterRuleValue(values[peerRESTListenPrefix][0]); err != nil {
			s.WriteErrorResponse(w, err)
			return
		}

		prefix = values[peerRESTListenPrefix][0]
	}

	var suffix string
	if len(values[peerRESTListenSuffix]) > 1 {
		s.WriteErrorResponse(w, errors.New("invalid request"))
		return
	}

	if len(values[peerRESTListenSuffix]) == 1 {
		if err := event.ValidateFilterRuleValue(values[peerRESTListenSuffix][0]); err != nil {
			s.WriteErrorResponse(w, err)
			return
		}

		suffix = values[peerRESTListenSuffix][0]
	}

	pattern := event.NewPattern(prefix, suffix)

	var eventNames []event.Name
	for _, ev := range values[peerRESTListenEvents] {
		eventName, err := event.ParseName(ev)
		if err != nil {
			s.WriteErrorResponse(w, err)
			return
		}

		eventNames = append(eventNames, eventName)
	}

	rulesMap := event.NewRulesMap(eventNames, pattern, event.TargetID{ID: mustGetUUID()})

	w.WriteHeader(http.StatusOK)
	w.(http.Flusher).Flush()

	doneCh := make(chan struct{})
	defer close(doneCh)

	// Listen Publisher uses nonblocking publish and hence does not wait for slow subscribers.
	// Use buffered channel to take care of burst sends or slow w.Write()
	ch := make(chan interface{}, 2000)

	globalHTTPListen.Subscribe(ch, doneCh, func(evI interface{}) bool {
		ev, ok := evI.(event.Event)
		if !ok {
			return false
		}
		if ev.S3.Bucket.Name != "" && values.Get(peerRESTListenBucket) != "" {
			if ev.S3.Bucket.Name != values.Get(peerRESTListenBucket) {
				return false
			}
		}
		return rulesMap.MatchSimple(ev.EventName, ev.S3.Object.Key)
	})

	keepAliveTicker := time.NewTicker(500 * time.Millisecond)
	defer keepAliveTicker.Stop()

	enc := gob.NewEncoder(w)
	for {
		select {
		case ev := <-ch:
			if err := enc.Encode(ev); err != nil {
				return
			}
			w.(http.Flusher).Flush()
		case <-keepAliveTicker.C:
			if err := enc.Encode(&event.Event{}); err != nil {
				return
			}
			w.(http.Flusher).Flush()
		}
	}
}

func extractTraceOptsFromPeerRequest(r *http.Request) (opts madmin.ServiceTraceOpts, err error) {

	q := r.URL.Query()
	opts.OnlyErrors = q.Get(peerRESTTraceErr) == "true"
	opts.Storage = q.Get(peerRESTTraceStorage) == "true"
	opts.Internal = q.Get(peerRESTTraceInternal) == "true"
	opts.S3 = q.Get(peerRESTTraceS3) == "true"
	opts.OS = q.Get(peerRESTTraceOS) == "true"

	if t := q.Get(peerRESTTraceThreshold); t != "" {
		d, err := time.ParseDuration(t)
		if err != nil {
			return opts, err
		}
		opts.Threshold = d
	}
	return
}

// TraceHandler sends http trace messages back to peer rest client
func (s *peerRESTServer) TraceHandler(w http.ResponseWriter, r *http.Request) {
	if !s.IsValid(w, r) {
		s.WriteErrorResponse(w, errors.New("Invalid request"))
		return
	}

	traceOpts, err := extractTraceOptsFromPeerRequest(r)
	if err != nil {
		s.WriteErrorResponse(w, errors.New("Invalid request"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.(http.Flusher).Flush()

	doneCh := make(chan struct{})
	defer close(doneCh)

	// Trace Publisher uses nonblocking publish and hence does not wait for slow subscribers.
	// Use buffered channel to take care of burst sends or slow w.Write()
	ch := make(chan interface{}, 2000)

	globalTrace.Subscribe(ch, doneCh, func(entry interface{}) bool {
		return mustTrace(entry, traceOpts)
	})

	keepAliveTicker := time.NewTicker(500 * time.Millisecond)
	defer keepAliveTicker.Stop()

	enc := gob.NewEncoder(w)
	for {
		select {
		case entry := <-ch:
			if err := enc.Encode(entry); err != nil {
				return
			}
			w.(http.Flusher).Flush()
		case <-keepAliveTicker.C:
			if err := enc.Encode(&trace.Info{}); err != nil {
				return
			}
			w.(http.Flusher).Flush()
		}
	}
}

func (s *peerRESTServer) BackgroundHealStatusHandler(w http.ResponseWriter, r *http.Request) {
	if !s.IsValid(w, r) {
		s.WriteErrorResponse(w, errors.New("invalid request"))
		return
	}
	ctx := NewContext(r, w, "BackgroundHealStatus")

	state, ok := getBackgroundHealStatus(ctx, newObjectLayerFn())
	if !ok {
		s.WriteErrorResponse(w, errServerNotInitialized)
		return
	}

	defer w.(http.Flusher).Flush()
	logger.LogIf(ctx, gob.NewEncoder(w).Encode(state))
}

// ConsoleLogHandler sends console logs of this node back to peer rest client
func (s *peerRESTServer) ConsoleLogHandler(w http.ResponseWriter, r *http.Request) {
	if !s.IsValid(w, r) {
		s.WriteErrorResponse(w, errors.New("Invalid request"))
		return
	}

	w.Header().Set("Connection", "close")
	w.WriteHeader(http.StatusOK)
	w.(http.Flusher).Flush()

	doneCh := make(chan struct{})
	defer close(doneCh)

	ch := make(chan interface{}, 2000)
	globalConsoleSys.Subscribe(ch, doneCh, "", 0, string(logger.All), nil)

	enc := gob.NewEncoder(w)
	for {
		select {
		case entry := <-ch:
			if err := enc.Encode(entry); err != nil {
				return
			}
			w.(http.Flusher).Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func (s *peerRESTServer) WriteErrorResponse(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte(err.Error()))
}

// IsValid - To authenticate and verify the time difference.
func (s *peerRESTServer) IsValid(w http.ResponseWriter, r *http.Request) bool {
	if err := storageServerRequestValidate(r); err != nil {
		s.WriteErrorResponse(w, err)
		return false
	}
	return true
}

// GetBandwidth gets the bandwidth for the buckets requested.
func (s *peerRESTServer) GetBandwidth(w http.ResponseWriter, r *http.Request) {
	if !s.IsValid(w, r) {
		s.WriteErrorResponse(w, errors.New("invalid request"))
		return
	}
	bucketsString := r.URL.Query().Get("buckets")
	w.WriteHeader(http.StatusOK)
	w.(http.Flusher).Flush()

	doneCh := make(chan struct{})
	defer close(doneCh)

	selectBuckets := b.SelectBuckets(strings.Split(bucketsString, ",")...)
	report := globalBucketMonitor.GetReport(selectBuckets)

	enc := gob.NewEncoder(w)
	if err := enc.Encode(report); err != nil {
		s.WriteErrorResponse(w, errors.New("Encoding report failed: "+err.Error()))
		return
	}
	w.(http.Flusher).Flush()
}

// GetPeerMetrics gets the metrics to be federated across peers.
func (s *peerRESTServer) GetPeerMetrics(w http.ResponseWriter, r *http.Request) {
	if !s.IsValid(w, r) {
		s.WriteErrorResponse(w, errors.New("invalid request"))
	}
	w.WriteHeader(http.StatusOK)
	w.(http.Flusher).Flush()

	doneCh := make(chan struct{})
	defer close(doneCh)

	enc := gob.NewEncoder(w)

	ch := ReportMetrics(r.Context(), GetGeneratorsForPeer)
	for m := range ch {
		if err := enc.Encode(m); err != nil {
			s.WriteErrorResponse(w, errors.New("Encoding metric failed: "+err.Error()))
			return
		}
	}
	w.(http.Flusher).Flush()
}

// registerPeerRESTHandlers - register peer rest router.
func registerPeerRESTHandlers(router *mux.Router) {
	server := &peerRESTServer{}
	subrouter := router.PathPrefix(peerRESTPrefix).Subrouter()
	subrouter.Methods(http.MethodPost).Path(peerRESTVersionPrefix + peerRESTMethodHealth).HandlerFunc(HTTPTraceHdrs(server.HealthHandler))
	subrouter.Methods(http.MethodPost).Path(peerRESTVersionPrefix + peerRESTMethodGetLocks).HandlerFunc(HTTPTraceHdrs(server.GetLocksHandler))
	subrouter.Methods(http.MethodPost).Path(peerRESTVersionPrefix + peerRESTMethodServerInfo).HandlerFunc(HTTPTraceHdrs(server.ServerInfoHandler))
	subrouter.Methods(http.MethodPost).Path(peerRESTVersionPrefix + peerRESTMethodProcInfo).HandlerFunc(HTTPTraceHdrs(server.ProcInfoHandler))
	subrouter.Methods(http.MethodPost).Path(peerRESTVersionPrefix + peerRESTMethodMemInfo).HandlerFunc(HTTPTraceHdrs(server.MemInfoHandler))
	subrouter.Methods(http.MethodPost).Path(peerRESTVersionPrefix + peerRESTMethodOsInfo).HandlerFunc(HTTPTraceHdrs(server.OsInfoHandler))
	subrouter.Methods(http.MethodPost).Path(peerRESTVersionPrefix + peerRESTMethodDiskHwInfo).HandlerFunc(HTTPTraceHdrs(server.DiskHwInfoHandler))
	subrouter.Methods(http.MethodPost).Path(peerRESTVersionPrefix + peerRESTMethodCPUInfo).HandlerFunc(HTTPTraceHdrs(server.CPUInfoHandler))
	subrouter.Methods(http.MethodPost).Path(peerRESTVersionPrefix + peerRESTMethodDriveInfo).HandlerFunc(HTTPTraceHdrs(server.DriveInfoHandler))
	subrouter.Methods(http.MethodPost).Path(peerRESTVersionPrefix + peerRESTMethodNetInfo).HandlerFunc(HTTPTraceHdrs(server.NetInfoHandler))
	subrouter.Methods(http.MethodPost).Path(peerRESTVersionPrefix + peerRESTMethodDispatchNetInfo).HandlerFunc(HTTPTraceHdrs(server.DispatchNetInfoHandler))
	subrouter.Methods(http.MethodPost).Path(peerRESTVersionPrefix + peerRESTMethodCycleBloom).HandlerFunc(HTTPTraceHdrs(server.CycleServerBloomFilterHandler))
	subrouter.Methods(http.MethodPost).Path(peerRESTVersionPrefix + peerRESTMethodDeleteBucketMetadata).HandlerFunc(HTTPTraceHdrs(server.DeleteBucketMetadataHandler)).Queries(restQueries(peerRESTBucket)...)
	subrouter.Methods(http.MethodPost).Path(peerRESTVersionPrefix + peerRESTMethodLoadBucketMetadata).HandlerFunc(HTTPTraceHdrs(server.LoadBucketMetadataHandler)).Queries(restQueries(peerRESTBucket)...)
	subrouter.Methods(http.MethodPost).Path(peerRESTVersionPrefix + peerRESTMethodGetBucketStats).HandlerFunc(HTTPTraceHdrs(server.GetBucketStatsHandler)).Queries(restQueries(peerRESTBucket)...)
	subrouter.Methods(http.MethodPost).Path(peerRESTVersionPrefix + peerRESTMethodSignalService).HandlerFunc(HTTPTraceHdrs(server.SignalServiceHandler)).Queries(restQueries(peerRESTSignal)...)
	subrouter.Methods(http.MethodPost).Path(peerRESTVersionPrefix + peerRESTMethodServerUpdate).HandlerFunc(HTTPTraceHdrs(server.ServerUpdateHandler))
	subrouter.Methods(http.MethodPost).Path(peerRESTVersionPrefix + peerRESTMethodDeletePolicy).HandlerFunc(HTTPTraceAll(server.DeletePolicyHandler)).Queries(restQueries(peerRESTPolicy)...)
	subrouter.Methods(http.MethodPost).Path(peerRESTVersionPrefix + peerRESTMethodLoadPolicy).HandlerFunc(HTTPTraceAll(server.LoadPolicyHandler)).Queries(restQueries(peerRESTPolicy)...)
	subrouter.Methods(http.MethodPost).Path(peerRESTVersionPrefix + peerRESTMethodLoadPolicyMapping).HandlerFunc(HTTPTraceAll(server.LoadPolicyMappingHandler)).Queries(restQueries(peerRESTUserOrGroup)...)
	subrouter.Methods(http.MethodPost).Path(peerRESTVersionPrefix + peerRESTMethodDeleteUser).HandlerFunc(HTTPTraceAll(server.DeleteUserHandler)).Queries(restQueries(peerRESTUser)...)
	subrouter.Methods(http.MethodPost).Path(peerRESTVersionPrefix + peerRESTMethodDeleteServiceAccount).HandlerFunc(HTTPTraceAll(server.DeleteServiceAccountHandler)).Queries(restQueries(peerRESTUser)...)
	subrouter.Methods(http.MethodPost).Path(peerRESTVersionPrefix + peerRESTMethodLoadUser).HandlerFunc(HTTPTraceAll(server.LoadUserHandler)).Queries(restQueries(peerRESTUser, peerRESTUserTemp)...)
	subrouter.Methods(http.MethodPost).Path(peerRESTVersionPrefix + peerRESTMethodLoadServiceAccount).HandlerFunc(HTTPTraceAll(server.LoadServiceAccountHandler)).Queries(restQueries(peerRESTUser)...)
	subrouter.Methods(http.MethodPost).Path(peerRESTVersionPrefix + peerRESTMethodLoadGroup).HandlerFunc(HTTPTraceAll(server.LoadGroupHandler)).Queries(restQueries(peerRESTGroup)...)

	subrouter.Methods(http.MethodPost).Path(peerRESTVersionPrefix + peerRESTMethodStartProfiling).HandlerFunc(HTTPTraceAll(server.StartProfilingHandler)).Queries(restQueries(peerRESTProfiler)...)
	subrouter.Methods(http.MethodPost).Path(peerRESTVersionPrefix + peerRESTMethodDownloadProfilingData).HandlerFunc(HTTPTraceHdrs(server.DownloadProfilingDataHandler))
	subrouter.Methods(http.MethodPost).Path(peerRESTVersionPrefix + peerRESTMethodTrace).HandlerFunc(server.TraceHandler)
	subrouter.Methods(http.MethodPost).Path(peerRESTVersionPrefix + peerRESTMethodListen).HandlerFunc(HTTPTraceHdrs(server.ListenHandler))
	subrouter.Methods(http.MethodPost).Path(peerRESTVersionPrefix + peerRESTMethodBackgroundHealStatus).HandlerFunc(server.BackgroundHealStatusHandler)
	subrouter.Methods(http.MethodPost).Path(peerRESTVersionPrefix + peerRESTMethodLog).HandlerFunc(server.ConsoleLogHandler)
	subrouter.Methods(http.MethodPost).Path(peerRESTVersionPrefix + peerRESTMethodGetLocalDiskIDs).HandlerFunc(HTTPTraceHdrs(server.GetLocalDiskIDs))
	subrouter.Methods(http.MethodPost).Path(peerRESTVersionPrefix + peerRESTMethodGetBandwidth).HandlerFunc(HTTPTraceHdrs(server.GetBandwidth))
	subrouter.Methods(http.MethodPost).Path(peerRESTVersionPrefix + peerRESTMethodGetMetacacheListing).HandlerFunc(HTTPTraceHdrs(server.GetMetacacheListingHandler))
	subrouter.Methods(http.MethodPost).Path(peerRESTVersionPrefix + peerRESTMethodUpdateMetacacheListing).HandlerFunc(HTTPTraceHdrs(server.UpdateMetacacheListingHandler))
	subrouter.Methods(http.MethodPost).Path(peerRESTVersionPrefix + peerRESTMethodGetPeerMetrics).HandlerFunc(HTTPTraceHdrs(server.GetPeerMetrics))
}
