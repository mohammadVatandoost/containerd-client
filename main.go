package main

import (
	"context"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

const WordPressNameSpace = "word-press"


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

	nameSpaces, err := client.NamespaceService().List(ctx)
	if err != nil {
		logrus.WithError(err).Error("client nameSpaces")
		return
	}

	for _, nameSpace := range nameSpaces {
		logrus.Infof("nameSpace: %v", nameSpace)

		imagesList, err := client.ListImages( namespaces.WithNamespace(ctx, nameSpace))
		if err != nil {
			logrus.WithError(err).Error("client ListImages")
			return
		}

		for _, image := range imagesList {
			logrus.Infof("image name: %v", image.Name())
		}
		logrus.Info("============")
	}
}

func CreateNameSpace(client *containerd.Client, nameSpace string) error {
	ctx,cnl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cnl()
    err := client.NamespaceService().Create(namespaces.WithNamespace(ctx, nameSpace),
		nameSpace, make(map[string]string))
	s, ok := status.FromError(err)
	if ok && s.Code() == codes.AlreadyExists {
		return nil
	}
	return err
}