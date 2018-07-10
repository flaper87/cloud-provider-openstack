/*
Copyright 2017 The Kubernetes Authors.

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

package manila

import (
	//"errors"

	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/golang/glog"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
	"github.com/pborman/uuid"
	"golang.org/x/net/context"
	"k8s.io/cloud-provider-openstack/pkg/csi/manila/openstack"
	volumeutil "k8s.io/kubernetes/pkg/volume/util"

    // Taken from pkg/share/manila
	//"github.com/gophercloud/gophercloud/openstack/sharedfilesystems/v2/shares"
)

type controllerServer struct {
	*csicommon.DefaultControllerServer
}

func (cs *controllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {

	// Volume Name
	shareName := req.GetName()
	if len(shareName) == 0 {
		shareName = uuid.NewUUID().String()
	}

	// Volume Size - Default is 1 GiB
	shareSizeBytes := int64(1 * 1024 * 1024 * 1024)
	if req.GetCapacityRange() != nil {
		shareSizeBytes = int64(req.GetCapacityRange().GetRequiredBytes())
	}
	shareSizeGB := int(volumeutil.RoundUpSize(shareSizeBytes, 1024*1024*1024))

	// Volume Type
	shareType := req.GetParameters()["type"]

	// Volume Type
	shareProto := req.GetParameters()["proto"]

	// Volume Availability - Default is nova
	shareAvailability := req.GetParameters()["availability"]

	// Get OpenStack Provider
	cloud, err := openstack.GetOpenStackProvider()
	if err != nil {
		glog.V(3).Infof("Failed to GetOpenStackProvider: %v", err)
		return nil, err
	}

	// Verify a share with the provided name doesn't already exist for this tenant
	share, err := cloud.GetShareByName(shareName)
	if err != nil {
		glog.V(3).Infof("Failed to query for existing Volume during CreateVolume: %v", err)
	}

	resID := ""
	resAvailability := ""

	if share.Name != "" {
		resID = share.ID
		resAvailability = share.AZ
	} else {
		// Volume Create
		share, err = cloud.CreateShare(shareName, shareSizeGB, shareProto, shareType, shareAvailability)
		if err != nil {
			glog.V(3).Infof("Failed to CreateVolume: %v", err)
			return nil, err
		}

		glog.V(4).Infof("Create share %s in Availability Zone: %s", share)

	}


	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			Id: resID,
			Attributes: map[string]string{
				"availability": resAvailability,
			},
		},
	}, nil
}

func (cs *controllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {

	// Get OpenStack Provider
	cloud, err := openstack.GetOpenStackProvider()
	if err != nil {
		glog.V(3).Infof("Failed to GetOpenStackProvider: %v", err)
		return nil, err
	}

	// Share Delete
    shareID := req.GetVolumeId()
    err = cloud.DeleteShare(shareID)
	return &csi.DeleteVolumeResponse{}, err
}

func (cs *controllerServer) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {

	// Get OpenStack Provider
	_, err := openstack.GetOpenStackProvider()
	if err != nil {
		glog.V(3).Infof("Failed to GetOpenStackProvider: %v", err)
		return nil, err
	}

	// Volume Attach
	_ = req.GetNodeId()
	_ = req.GetVolumeId()

	pvInfo := map[string]string{}

	return &csi.ControllerPublishVolumeResponse{
		PublishInfo: pvInfo,
	}, nil
}

func (cs *controllerServer) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {

	// Get OpenStack Provider
	_, err := openstack.GetOpenStackProvider()
	if err != nil {
		glog.V(3).Infof("Failed to GetOpenStackProvider: %v", err)
		return nil, err
	}

	// Volume Detach
	_ = req.GetNodeId()
	_ = req.GetVolumeId()

	return &csi.ControllerUnpublishVolumeResponse{}, nil
}
