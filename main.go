package main

import (
	"context"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/namespaces"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	WordPressNameSpace = "wordpress"
	WordPressImageName = "docker.io/library/wordpress:php8.1-apache"
)


func main() {
	client, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		logrus.WithError(err).Error("containerd new client")
		return
	}
	defer client.Close()

	err = CreateNameSpace(client, WordPressNameSpace)
	if err != nil {
		logrus.WithError(err).Error("client CreateNameSpace")
		return
	}

	ctx,cnl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cnl()

	err = PullImage(client, WordPressNameSpace, WordPressImageName)
	if err != nil {
		logrus.WithError(err).Error("client PullImage")
		return
	}



		imagesList, err := client.ListImages( namespaces.WithNamespace(ctx, WordPressNameSpace))
		if err != nil {
			logrus.WithError(err).Error("client ListImages")
			return
		}
		logrus.Info("============")
		for _, image := range imagesList {
			logrus.Infof("image name: %v", image.Name())
			ok, err := image.IsUnpacked(ctx, "btrfs")
			if err != nil {
				logrus.WithError(err).Error("client IsUnpacked")
				return
			}
			logrus.Infof("IsUnpacked btrfs: %v", ok)
		}
		logrus.Info("============")
}

func CreateNameSpace(client *containerd.Client, nameSpace string) error {
	ctx,cnl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cnl()
    err := client.NamespaceService().Create(namespaces.WithNamespace(ctx, nameSpace),
		nameSpace, make(map[string]string))

	if errdefs.IsAlreadyExists(err) {
		return nil
	}
	return err
}

func PullImage(client *containerd.Client, nameSpace string, imageName string) error {
	//ctx,cnl := context.WithTimeout(context.Background(), 1000*time.Second)
	//defer cnl()


	_, err := client.Pull(namespaces.WithNamespace(context.Background(), nameSpace),imageName)

	return err
}