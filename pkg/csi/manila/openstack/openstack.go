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

package openstack

import (
	"os"

	"github.com/golang/glog"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	//"github.com/gophercloud/gophercloud/openstack/sharedfilesystems/v2/shares"
	"gopkg.in/gcfg.v1"
)

type IOpenStack interface {
	CreateShare(name string, size int, proto string, stype string, availability string) (Share, error)
	DeleteShare(shareID string) error
	//AttachShare(instanceID, shareID string) (string, error)
	//WaitDiskAttached(instanceID string, shareID string) error
	//DetachShare(instanceID, shareID string) error
	//WaitDiskDetached(instanceID string, shareID string) error
	//GetAttachmentDiskPath(instanceID, shareID string) (string, error)
	GetShareByName(name string) (Share, error)
	//CreateSnapshot(name, volID, description string, tags *map[string]string) (*snapshots.Snapshot, error)
	//ListSnapshots(limit, offset int, filters map[string]string) ([]snapshots.Snapshot, error)
	//DeleteSnapshot(snapID string) error
}

type OpenStack struct {
	sharestorage *gophercloud.ServiceClient
}

type Config struct {
	Global struct {
		AuthUrl    string `gcfg:"auth-url"`
		Username   string
		UserId     string `gcfg:"user-id"`
		Password   string
		TenantId   string `gcfg:"tenant-id"`
		TenantName string `gcfg:"tenant-name"`
		DomainId   string `gcfg:"domain-id"`
		DomainName string `gcfg:"domain-name"`
		Region     string
	}
}

func (cfg Config) toAuthOptions() gophercloud.AuthOptions {
	return gophercloud.AuthOptions{
		IdentityEndpoint: cfg.Global.AuthUrl,
		Username:         cfg.Global.Username,
		UserID:           cfg.Global.UserId,
		Password:         cfg.Global.Password,
		TenantID:         cfg.Global.TenantId,
		TenantName:       cfg.Global.TenantName,
		DomainID:         cfg.Global.DomainId,
		DomainName:       cfg.Global.DomainName,

		// Persistent service, so we need to be able to renew tokens.
		AllowReauth: true,
	}
}

func GetConfigFromFile(configFilePath string) (gophercloud.AuthOptions, gophercloud.EndpointOpts, error) {
	// Get config from file
	var authOpts gophercloud.AuthOptions
	var epOpts gophercloud.EndpointOpts
	config, err := os.Open(configFilePath)
	if err != nil {
		glog.V(3).Infof("Failed to open OpenStack configuration file: %v", err)
		return authOpts, epOpts, err
	}
	defer config.Close()

	// Read configuration
	var cfg Config
	err = gcfg.FatalOnly(gcfg.ReadInto(&cfg, config))
	if err != nil {
		glog.V(3).Infof("Failed to read OpenStack configuration file: %v", err)
		return authOpts, epOpts, err
	}

	authOpts = cfg.toAuthOptions()
	epOpts = gophercloud.EndpointOpts{
		Region: cfg.Global.Region,
	}

	return authOpts, epOpts, nil
}

func GetConfigFromEnv() (gophercloud.AuthOptions, gophercloud.EndpointOpts, error) {
	// Get config from env
	authOpts, err := openstack.AuthOptionsFromEnv()
	var epOpts gophercloud.EndpointOpts
	if err != nil {
		glog.V(3).Infof("Failed to read OpenStack configuration from env: %v", err)
		return authOpts, epOpts, err
	}

	epOpts = gophercloud.EndpointOpts{
		Region: os.Getenv("OS_REGION_NAME"),
	}

	return authOpts, epOpts, nil
}

var OsInstance IOpenStack = nil
var configFile string = "/etc/cloud.conf"

func InitOpenStackProvider(cfg string) {
	configFile = cfg
	glog.V(2).Infof("InitOpenStackProvider configFile: %s", configFile)
}

func GetOpenStackProvider() (IOpenStack, error) {

	if OsInstance == nil {
		// Get config from file
		authOpts, epOpts, err := GetConfigFromFile(configFile)
		if err != nil {
			// Get config from env
			authOpts, epOpts, err = GetConfigFromEnv()
			if err != nil {
				return nil, err
			}
		}

		// Authenticate Client
		provider, err := openstack.AuthenticatedClient(authOpts)
		if err != nil {
			return nil, err
		}

		// Init Cinder ServiceClient
		sharestorageclient, err := openstack.NewSharedFileSystemV2(provider, epOpts)
		if err != nil {
			return nil, err
		}

		// Init OpenStack
		OsInstance = &OpenStack{
            sharestorage: sharestorageclient,
		}
	}

	return OsInstance, nil
}
