/*
Copyright 2018 The Kubernetes Authors.

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

package openstack

import (
	//"fmt"
	"time"

	"github.com/gophercloud/gophercloud/openstack/sharedfilesystems/v2/shares"
	//"k8s.io/apimachinery/pkg/util/wait"

	//"github.com/golang/glog"
)

const (
	ShareAvailableStatus       = "available"
	ShareInUseStatus           = "in-use"
	ShareDeletedStatus         = "deleted"
	ShareErrorStatus           = "error"
	fsOperationFinishInitDelay = 1 * time.Second
	fsOperationFinishFactor    = 1.1
	fsOperationFinishSteps     = 10
	fsAttachInitDelay          = 1 * time.Second
	fsAttachFactor             = 1.2
	fsAttachSteps              = 15
	fsDetachInitDelay          = 1 * time.Second
	fsDetachFactor             = 1.2
	fsDetachSteps              = 13
)

type Share struct {
	// ID of the instance, to which this volume is attached. "" if not attached
	AttachedServerId string
	// Device file path
	AttachedDevice string
	// Unique identifier for the volume.
	ID string
	// Human-readable display name for the volume.
	Name string
	// Current status of the volume.
	Status string
	// Volume size in GB
	Size int
	// Availability Zone the volume belongs to
	AZ string
}

// GetVolumesByName is a wrapper around ListVolumes that creates a Name filter to act as a GetByName
// Returns a list of Volume references with the specified name
func (os *OpenStack) GetShareByName(n string) (Share, error) {
	s, err := shares.Get(os.sharestorage, n).Extract()

	if err != nil {
		return Share{}, err
	}

	return Share{
			ID:     s.ID,
			Name:   s.Name,
			Status: s.Status,
			AZ:     s.AvailabilityZone,
		}, nil
}

// CreateShare creates a share of given size
func (os *OpenStack) CreateShare(name string, size int, proto string, stype string, availability string) (Share, error) {
	opts := &shares.CreateOpts{
		Name:             name,
		Size:             size,
		ShareType:        stype,
		ShareProto:       proto,
		AvailabilityZone: availability,
	}

	s, err := shares.Create(os.sharestorage, opts).Extract()
	if err != nil {
		return Share{}, err
	}

	return Share{
			ID:     s.ID,
			Name:   s.Name,
			Status: s.Status,
			AZ:     s.AvailabilityZone,
		}, nil
}

// DeleteVolume delete a share
func (os *OpenStack) DeleteShare(shareID string) error {
	//used, err := os.diskIsUsed(shareID)
	//if err != nil {
	//	return err
	//}
	//if used {
	//	return fmt.Errorf("Cannot delete the share %q, it's still attached to a node", shareID)
	//}

    err := shares.Delete(os.sharestorage, shareID).ExtractErr()
	return err
}
