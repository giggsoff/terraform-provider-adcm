package adcm

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var (
	ADCM_URL      string
	ADCM_LOGIN    string
	ADCM_PASSWORD string
	ADCM_ID       string
)
var testAccProviderServers map[string]func() (tfprotov6.ProviderServer, error)
var testAccProvider provider.Provider
var bundleStoreAddr string

func readENV() {
	ADCM_URL = os.Getenv("ADCM_URL")
	ADCM_LOGIN = os.Getenv("ADCM_LOGIN")
	ADCM_PASSWORD = os.Getenv("ADCM_PASSWORD")
}

func createListener() (l net.Listener, close func()) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	return l, func() {
		_ = l.Close()
	}
}

func initBundleStore(port chan int) error {
	l, closeFunction := createListener()
	port <- l.Addr().(*net.TCPAddr).Port
	defer closeFunction()
	return http.Serve(l, http.FileServer(http.Dir("./test_bundles")))
}

func init() {
	/*
		port, adcmID := testRunADCM()
		if err := os.Setenv("ADCM_URL", fmt.Sprintf("http://127.0.0.1:%s", port)); err != nil {
			panic(err)
		}
		if err := os.Setenv("ADCM_LOGIN", "admin"); err != nil {
			panic(err)
		}
		if err := os.Setenv("ADCM_PASSWORD", "admin"); err != nil {
			panic(err)
		}
		ADCM_ID = adcmID
	*/
	readENV()
	testAccProvider = New()
	testAccProviderServers = map[string]func() (tfprotov6.ProviderServer, error){
		// newProvider is an example function that returns a provider.Provider
		"adcm": providerserver.NewProtocol6WithError(testAccProvider),
	}
	testPrepareLocalBundleStore()
}

func testAccPreCheckRequiredEnvVars(t *testing.T) {
	if ADCM_URL == "" {
		t.Fatal("ADCM_URL must be set for acceptance tests")
	}
	if ADCM_LOGIN == "" {
		t.Fatal("ADCM_LOGIN must be set for acceptance tests")
	}
	if ADCM_PASSWORD == "" {
		t.Fatal("ADCM_PASSWORD must be set for acceptance tests")
	}
}

func testAccPreCheck(t *testing.T) {
	testAccPreCheckRequiredEnvVars(t)
}

func testPrepareLocalBundleStore() {
	port := make(chan int)
	go func() {
		if err := initBundleStore(port); err != nil {
			log.Panicln(err)
		}
	}()
	bundleStoreAddr = fmt.Sprintf("http://127.0.0.1:%d", <-port)
}

func testRunADCM() (string, string) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	imageName := "arenadata/adcm"
	port, err := nat.NewPort("tcp", "8000")
	if err != nil {
		panic(err)
	}
	portBinding := nat.PortMap{port: []nat.PortBinding{
		{
			HostIP:   "0.0.0.0",
			HostPort: "0",
		},
	}}

	out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	defer out.Close()
	_, err = io.Copy(os.Stdout, out)
	if err != nil {
		panic(err)
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
	}, &container.HostConfig{
		PortBindings: portBinding}, nil, nil, "")
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}
	c, err := cli.ContainerInspect(ctx, resp.ID)
	if err != nil {
		panic(err)
	}
	for _, ports := range c.NetworkSettings.Ports {
		for _, p := range ports {
			return p.HostPort, resp.ID
		}
	}
	return "", resp.ID
}

func destroyADCM(id string) error {
	fmt.Println("1")
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	return cli.ContainerRemove(ctx, id, types.ContainerRemoveOptions{Force: true, RemoveVolumes: true})
}
