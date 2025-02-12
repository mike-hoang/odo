package registry

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/config"
	envcontext "github.com/redhat-developer/odo/pkg/config/context"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

func TestGetDevfileRegistries(t *testing.T) {
	tempConfigFile, err := ioutil.TempFile("", "odoconfig")
	if err != nil {
		t.Fatal("Fail to create temporary config file")
	}
	defer os.Remove(tempConfigFile.Name())
	defer tempConfigFile.Close()
	_, err = tempConfigFile.Write([]byte(
		`kind: Preference
apiversion: odo.openshift.io/v1alpha1
OdoSettings:
  RegistryList:
  - Name: DefaultDevfileRegistry
    URL: https://registry.devfile.io
  - Name: CheDevfileRegistry
    URL: https://che-devfile-registry.openshift.io/`,
	))
	if err != nil {
		t.Error(err)
	}
	tempConfigFileName := tempConfigFile.Name()

	tests := []struct {
		name         string
		registryName string
		want         []api.Registry
	}{
		{
			name:         "Case 1: Test get all devfile registries",
			registryName: "",
			want: []api.Registry{
				{
					Name:   "CheDevfileRegistry",
					URL:    "https://che-devfile-registry.openshift.io/",
					Secure: false,
				},
				{
					Name:   "DefaultDevfileRegistry",
					URL:    "https://registry.devfile.io",
					Secure: false,
				},
			},
		},
		{
			name:         "Case 2: Test get specific devfile registry",
			registryName: "CheDevfileRegistry",
			want: []api.Registry{
				{
					Name:   "CheDevfileRegistry",
					URL:    "https://che-devfile-registry.openshift.io/",
					Secure: false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctx = envcontext.WithEnvConfig(ctx, config.Configuration{
				Globalodoconfig: &tempConfigFileName,
			})
			prefClient, _ := preference.NewClient(ctx)
			catClient := NewRegistryClient(filesystem.NewFakeFs(), prefClient)
			got, err := catClient.GetDevfileRegistries(tt.registryName)
			if err != nil {
				t.Errorf("Error message is %v", err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("RegistryClient.GetDevfileRegistries() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestListDevfileStacks(t *testing.T) {
	// Start a local HTTP server
	// to test getting multiple devfiles via ListDevfileStacks
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Send response to be tested
		_, err := rw.Write([]byte(
			`
			[
				{
					"name": "nodejs",
					"displayName": "NodeJS Angular Web Application",
					"description": "Stack for developing NodeJS Angular Web Application",
					"tags": [
						"NodeJS",
						"Angular",
						"Alpine"
					],
					"language": "nodejs",
					"icon": "/images/angular.svg",
					"globalMemoryLimit": "2686Mi",
					"links": {
						"self": "/devfiles/angular/devfile.yaml"
					}
				},
				{
					"name": "python",
					"displayName": "Python",
					"description": "Python Stack with Python 3.7",
					"tags": [
						"Python",
						"pip"
					],
					"language": "python",
					"icon": "/images/foobar.svg",
					"globalMemoryLimit": "2686Mi",
					"links": {
						"self": "/devfiles/python/devfile.yaml"
					}
				}
			]
			`,
		))
		if err != nil {
			t.Error(err)
		}
	}))
	// Close the server when test finishes
	defer server.Close()

	const registryName = "TestRegistry"
	tests := []struct {
		name         string
		registryName string
		devfileName  string
		filter       string
		want         DevfileStackList
	}{
		{
			name:         "Case 1: Test getting ALL registries and looking for nodejs",
			registryName: "",
			want: DevfileStackList{
				DevfileRegistries: []api.Registry{
					{
						Name:   "TestRegistry",
						URL:    server.URL,
						Secure: false,
					},
				},
				Items: []api.DevfileStack{
					{
						Name:        "nodejs",
						DisplayName: "NodeJS Angular Web Application",
						Description: "Stack for developing NodeJS Angular Web Application",
						Registry: api.Registry{
							Name: registryName,
							URL:  server.URL,
						},
						Language: "nodejs",
						Tags:     []string{"NodeJS", "Angular", "Alpine"},
					},
					{
						Name:        "python",
						DisplayName: "Python",
						Description: "Python Stack with Python 3.7",
						Registry: api.Registry{
							Name: registryName,
							URL:  server.URL,
						},
						Language: "python",
						Tags:     []string{"Python", "pip"},
					},
				},
			},
		},
		{
			name:         "Case 2: Test getting from only one specific devfile and from a specific registry",
			registryName: "TestRegistry",
			devfileName:  "nodejs",
			want: DevfileStackList{
				DevfileRegistries: []api.Registry{
					{
						Name:   "TestRegistry",
						URL:    server.URL,
						Secure: false,
					},
				},
				Items: []api.DevfileStack{
					{
						Name:        "nodejs",
						DisplayName: "NodeJS Angular Web Application",
						Description: "Stack for developing NodeJS Angular Web Application",
						Registry: api.Registry{
							Name: registryName,
							URL:  server.URL,
						},
						Language: "nodejs",
						Tags:     []string{"NodeJS", "Angular", "Alpine"},
					},
				},
			},
		},
		{
			name:         "Case 3: Test getting a devfile using a filter from the description",
			registryName: "TestRegistry",
			filter:       "Python Stack",
			want: DevfileStackList{
				DevfileRegistries: []api.Registry{
					{
						Name:   "TestRegistry",
						URL:    server.URL,
						Secure: false,
					},
				},
				Items: []api.DevfileStack{
					{
						Name:        "python",
						DisplayName: "Python",
						Description: "Python Stack with Python 3.7",
						Registry: api.Registry{
							Name: registryName,
							URL:  server.URL,
						},
						Language: "python",
						Tags:     []string{"Python", "pip"},
					},
				},
			},
		},
		{
			name:         "Case 4: Expect nothing back if registry is not found",
			registryName: "Foobar",
			want:         DevfileStackList{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			prefClient := preference.NewMockClient(ctrl)
			prefClient.EXPECT().RegistryList().Return([]preference.Registry{
				{
					Name: "TestRegistry",
					URL:  server.URL,
				},
			}).AnyTimes()
			catClient := NewRegistryClient(filesystem.NewFakeFs(), prefClient)
			ctx := context.Background()
			ctx = envcontext.WithEnvConfig(ctx, config.Configuration{})
			got, err := catClient.ListDevfileStacks(ctx, tt.registryName, tt.devfileName, tt.filter, false)
			if err != nil {
				t.Error(err)
			}

			if diff := cmp.Diff(tt.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("RegistryClient.ListDevfileStacks() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGetRegistryDevfiles(t *testing.T) {
	// Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Send response to be tested
		_, err := rw.Write([]byte(
			`
			[
				{
					"name": "nodejs",
					"displayName": "NodeJS Angular Web Application",
					"description": "Stack for developing NodeJS Angular Web Application",
					"tags": [
						"NodeJS",
						"Angular",
						"Alpine"
					],
					"language": "nodejs",
					"icon": "/images/angular.svg",
					"globalMemoryLimit": "2686Mi",
					"links": {
						"self": "/devfiles/angular/devfile.yaml"
					}
				}
			]
			`,
		))
		if err != nil {
			t.Error(err)
		}
	}))
	// Close the server when test finishes
	defer server.Close()

	const registryName = "some registry"
	tests := []struct {
		name     string
		registry api.Registry
		want     []api.DevfileStack
	}{
		{
			name:     "Test NodeJS devfile index",
			registry: api.Registry{Name: registryName, URL: server.URL},
			want: []api.DevfileStack{
				{
					Name:        "nodejs",
					DisplayName: "NodeJS Angular Web Application",
					Description: "Stack for developing NodeJS Angular Web Application",
					Registry: api.Registry{
						Name: registryName,
						URL:  server.URL,
					},
					Language: "nodejs",
					Tags:     []string{"NodeJS", "Angular", "Alpine"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			prefClient := preference.NewMockClient(ctrl)
			ctx := context.Background()
			ctx = envcontext.WithEnvConfig(ctx, config.Configuration{})
			got, err := getRegistryStacks(ctx, prefClient, tt.registry)

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("getRegistryStacks() mismatch (-want +got):\n%s", diff)
				t.Logf("Error message is: %v", err)
			}
		})
	}
}
