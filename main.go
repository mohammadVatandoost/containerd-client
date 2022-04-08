package main

import (
	"context"
	"encoding/json"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

const (
	WordPressNameSpace = "wordpress"
	WordPressImageName = "docker.io/library/wordpress:php8.1-apache"
	SnapShotterFormat = "native"
	SpecFileName = "config.json"
	ContainerName = "wordpress-containerd"
)

func main() {
	startTime := time.Now()
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

	ctxWithNameSpace := namespaces.WithNamespace(ctx, WordPressNameSpace)
	wordPressImage, err := client.GetImage(ctxWithNameSpace,WordPressImageName)
	if err != nil {
		logrus.WithError(err).Error("client GetImage")
		return
	}

	ok, err := wordPressImage.IsUnpacked(ctxWithNameSpace, SnapShotterFormat)
	if err != nil {
		logrus.WithError(err).Error("client IsUnpacked")
		return
	}
	if !ok {
		logrus.Infof("unpack image: %v", wordPressImage.Name())
		err = wordPressImage.Unpack(ctxWithNameSpace, "")
		if err != nil {
			logrus.WithError(err).Error("client Unpack")
			return
		}
	}

	container, err := client.NewContainer(ctxWithNameSpace, ContainerName,
		containerd.WithNewSnapshot(WordPressImageName+"-rootfs", wordPressImage),
		containerd.WithNewSpec(oci.WithImageConfig(wordPressImage)))
	if err != nil {
		logrus.WithError(err).Error("client NewContainer")
		return
	}
	defer func() {
		err := client.ContainerService().Delete(ctxWithNameSpace, ContainerName)
		if err != nil {
			logrus.WithError(err).Error("client ContainerService Delete")
		}
	}()
	logrus.Infof("container ID: %v", container.ID())
	spec, err := container.Spec(ctxWithNameSpace)
	if err != nil {
		logrus.WithError(err).Error("client NewContainer")
		return
	}
	b, err := json.Marshal(spec)
	if err != nil {
		logrus.WithError(err).Error("client NewContainer")
		return
	}

	err = WriteSpecToFile(SpecFileName, b)
	if err != nil {
		logrus.WithError(err).Error("WriteSpecToFile")
		return
	}
	//logrus.Info("============")
	//for _, image := range imagesList {
	//	logrus.Infof("image name: %v", image.Name())
	//	ok, err := image.IsUnpacked(ctxWithNameSpace, SnapShotterFormat)
	//	if err != nil {
	//		logrus.WithError(err).Error("client IsUnpacked")
	//		return
	//	}
	//	//if !ok {
	//	//	image.Unpack(ctxWithNameSpace, SnapShotterFormat)
	//	//}
	//	image.
	//	logrus.Infof("IsUnpacked btrfs: %v", ok)
	//}
	//client.NewContainer(ctxWithNameSpace, WordPressImageName+"test", )
	logrus.Infof("====== duration: %v ms ======", time.Now().Sub(startTime).Milliseconds())

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
	_, err := client.ImageService().Get(namespaces.WithNamespace(context.Background(), nameSpace),imageName)
	if err == nil {
		return nil
	}
	_, err = client.Pull(namespaces.WithNamespace(context.Background(), nameSpace),imageName)
	return err
}

func WriteSpecToFile(fileName string, data []byte) error {
	fo, err := os.Create(fileName)
	if err != nil {
		return err
	}

	if _, err := fo.Write(data); err != nil {
		return err
	}

	if err := fo.Close(); err != nil {
		return err
	}
	return nil
}