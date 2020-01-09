package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v2"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	//"github.com/Juniper/contrail-operator"
	contrailOperatorTypes "github.com/Juniper/contrail-operator/pkg/apis/contrail/v1alpha1"
)

//NodeType definition
type NodeType string

const (
	//VROUTER nodetype
	VROUTER NodeType = "vrouter"
	//CONFIG nodetype
	CONFIG = "config"
	//ANALYTICS nodetype
	ANALYTICS = "analytics"
	//CONTROL nodetype
	CONTROL = "control"
)

var ControlStatus contrailOperatorTypes.ControlServiceStatus

/*
type ControlStatus struct {
	Connections              []Connection
	NumberOfXMPPPeers        string
	NumberOfRoutingInstances string
	StaticRoutes             StaticRoutes
	BGPPeer                  BGPPeer
	State                    string
}

type StaticRoutes struct {
	Down   string
	Number string
}

type BGPPeer struct {
	Up     string
	Number string
}

type Connection struct {
	Type   string
	Name   string
	Status string
	Nodes  []string
}
*/

//Config struct
type Config struct {
	APIServerList  []string   `yaml:"apiServerList,omitempty"`
	Encryption     encryption `yaml:"encryption,omitempty"`
	NodeType       NodeType   `yaml:"nodeType,omitempty"`
	Interval       int64      `yaml:"interval,omitempty"`
	Hostname       string     `yaml:"hostname,omitempty"`
	InCluster      *bool      `yaml:"inCluster,omitempty"`
	KubeConfigPath string     `yaml:"kubeConfigPath,omitempty"`
	NodeName       string     `yaml:"nodeName,omitempty"`
	Namespace      string     `yaml:"namespace,omitempty"`
}

type controlUVEStatus struct {
	NodeStatus struct {
		T             [][]interface{} `json:"__T"`
		ProcessStatus struct {
			Aggtype string `json:"@aggtype"`
			List    struct {
				ProcessStatus []struct {
					ModuleID struct {
						Type string `json:"@type"`
						Text string `json:"#text"`
					} `json:"module_id"`
					InstanceID struct {
						Type string `json:"@type"`
						Text string `json:"#text"`
					} `json:"instance_id"`
					State struct {
						Type string `json:"@type"`
						Text string `json:"#text"`
					} `json:"state"`
					ConnectionInfos struct {
						Type string `json:"@type"`
						List struct {
							Type           string `json:"@type"`
							Size           string `json:"@size"`
							ConnectionInfo []struct {
								Type struct {
									Type string `json:"@type"`
									Text string `json:"#text"`
								} `json:"type"`
								Name struct {
									Type string `json:"@type"`
									Text string `json:"#text,omitempty"`
								} `json:"name,omitempty"`
								ServerAddrs struct {
									Type string `json:"@type"`
									List struct {
										Type    string      `json:"@type"`
										Size    string      `json:"@size"`
										Element interface{} `json:"element"`
									} `json:"list"`
								} `json:"server_addrs"`
								Status struct {
									Type string `json:"@type"`
									Text string `json:"#text"`
								} `json:"status"`
								Description struct {
									Type string `json:"@type"`
									Text string `json:"#text"`
								} `json:"description"`
								/*
									NameA struct {
										Type string `json:"@type"`
										Text string `json:"#text"`
									} `json:"name,omitempty"`
									NameB struct {
										Type string `json:"@type"`
										Text string `json:"#text"`
									} `json:"name,omitempty"`
								*/
							} `json:"ConnectionInfo"`
						} `json:"list"`
					} `json:"connection_infos"`
					Description struct {
						Type string `json:"@type"`
					} `json:"description"`
				} `json:"ProcessStatus"`
				Type string `json:"@type"`
				Size string `json:"@size"`
			} `json:"list"`
			Type string `json:"@type"`
		} `json:"process_status"`
	} `json:"NodeStatus"`

	BgpRouterState struct {
		NumDownServiceChains      [][]interface{}         `json:"num_down_service_chains"`
		BgpRouterIPList           [][]interface{}         `json:"bgp_router_ip_list"`
		NumUpXMPPPeer             [][]NumUpXMPPPeer       `json:"num_up_xmpp_peer"`
		OutputQueueDepth          [][]interface{}         `json:"output_queue_depth"`
		NumDownStaticRoutes       [][]NumDownStaticRoutes `json:"num_down_static_routes"`
		Uptime                    [][]interface{}         `json:"uptime"`
		NumDeletingXMPPPeer       [][]interface{}         `json:"num_deleting_xmpp_peer"`
		LocalAsn                  [][]interface{}         `json:"local_asn"`
		DbConnInfo                [][]interface{}         `json:"db_conn_info"`
		NumXMPPPeer               [][]interface{}         `json:"num_xmpp_peer"`
		NumDeletingBgpPeer        [][]interface{}         `json:"num_deleting_bgp_peer"`
		NumStaticRoutes           [][]NumStaticRoutes     `json:"num_static_routes"`
		RouterID                  [][]interface{}         `json:"router_id"`
		AdminDown                 [][]interface{}         `json:"admin_down"`
		NumUpBgpaasPeer           [][]interface{}         `json:"num_up_bgpaas_peer"`
		T                         [][]interface{}         `json:"__T"`
		NumDeletedRoutingInstance [][]interface{}         `json:"num_deleted_routing_instance"`
		NumServiceChains          [][]interface{}         `json:"num_service_chains"`
		GlobalAsn                 [][]interface{}         `json:"global_asn"`
		NumRoutingInstance        [][]NumRoutingInstance  `json:"num_routing_instance"`
		BuildInfo                 [][]interface{}         `json:"build_info"`
		IfmapServerInfo           [][]interface{}         `json:"ifmap_server_info"`
		NumUpBgpPeer              [][]NumUpBgpPeer        `json:"num_up_bgp_peer"`
		AmqpConnInfo              [][]interface{}         `json:"amqp_conn_info"`
		NumBgpaasPeer             [][]interface{}         `json:"num_bgpaas_peer"`
		NumBgpPeer                [][]NumBgpPeer          `json:"num_bgp_peer"`
		NumDeletingBgpaasPeer     [][]interface{}         `json:"num_deleting_bgpaas_peer"`
	} `json:"BgpRouterState"`
	ContrailConfig struct {
		Deleted  [][]interface{} `json:"deleted"`
		T        [][]interface{} `json:"__T"`
		Elements [][]interface{} `json:"elements"`
	} `json:"ContrailConfig"`
}

type NumUpXMPPPeer struct {
	Type string `json:"@type"`
	Text string `json:"#text"`
}

type NumDownStaticRoutes struct {
	Type string `json:"@type"`
	Text string `json:"#text"`
}

type NumStaticRoutes struct {
	Type string `json:"@type"`
	Text string `json:"#text"`
}

type NumRoutingInstance struct {
	Type string `json:"@type"`
	Text string `json:"#text"`
}

type NumUpBgpPeer struct {
	Type string `json:"@type"`
	Text string `json:"#text"`
}

type NumBgpPeer struct {
	Type string `json:"@type"`
	Text string `json:"#text"`
}

type BgpRouterIPList struct {
	Type string `json:"@type"`
	List struct {
		Type    string   `json:"@type"`
		Size    string   `json:"@size"`
		Element []string `json:"element"`
	} `json:"list"`
}

type encryption struct {
	CA       *string `yaml:"ca,omitempty"`
	Cert     *string `yaml:"cert,omitempty"`
	Key      *string `yaml:"key,omitempty"`
	Insecure bool    `yaml:"insecure,omitempty"`
}

func check(err error) {
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}

func main() {
	configPtr := flag.String("config", "/config.yaml", "path to config yaml file")
	intervalPtr := flag.Int64("interval", 1, "interval for getting status")
	flag.Parse()

	ticker := time.NewTicker(time.Duration(*intervalPtr) * time.Second)
	done := make(chan bool)
	go func() {
		for {
			select {
			case t := <-ticker.C:
				fmt.Println("Tick at", t)
				var config Config
				configYaml, err := ioutil.ReadFile(*configPtr)
				if err != nil {
					panic(err)
				}
				err = yaml.Unmarshal(configYaml, &config)
				if err != nil {
					panic(err)
				}
				ticker = time.NewTicker(time.Duration(config.Interval) * time.Second)
				getStatus(config)
			}
		}
	}()
	done <- true
	fmt.Println("Ticker stopped")
}

func getStatus(config Config) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: config.Encryption.Insecure,
	}
	if config.Encryption.Key != nil && config.Encryption.Cert != nil && config.Encryption.CA != nil {
		caCert, err := ioutil.ReadFile(*config.Encryption.CA)
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = pool
		cer, err := tls.LoadX509KeyPair(*config.Encryption.Cert, *config.Encryption.Key)
		if err != nil {
			log.Println(err)
			return
		}
		tlsConfig.Certificates = []tls.Certificate{cer}
	}
	transport := http.Transport{
		TLSClientConfig: tlsConfig,
	}
	client := http.Client{
		Transport: &transport,
	}
	var nodeType string
	switch config.NodeType {
	case "config":
		nodeType = "config-node"
	case "control":
		nodeType = "control-node"
	case "analytics":
		nodeType = "analytics-node"
	case "vrouter":
		nodeType = "vrouter"
	}
	for _, apiServer := range config.APIServerList {
		resp, err := client.Get("https://" + apiServer + "/analytics/uves/" + nodeType + "/" + config.Hostname)
		if err != nil {
			fmt.Println(err)
		}
		if resp != nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				bodyBytes, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Fatal(err)
				}

				switch config.NodeType {
				case "control":
					controlStatus := getControlStatus(bodyBytes)
					clientset, restClient, err := kubeClient(config)
					check(err)
					err = updateControlStatus(&config, *controlStatus, clientset, restClient)

					if err != nil {
						log.Fatal(err)
					}

				}
				break
			}
		}
	}
}
func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func kubeClient(config Config) (*kubernetes.Clientset, *rest.RESTClient, error) {
	var err error
	clientset := &kubernetes.Clientset{}
	restClient := &rest.RESTClient{}
	kubeConfig := &rest.Config{}
	if config.InCluster != nil && !*config.InCluster {
		var kubeConfigPath string
		if config.KubeConfigPath != "" {
			kubeConfigPath = config.KubeConfigPath
		} else {
			kubeConfigPath = filepath.Join(homeDir(), ".kube", "config")
		}
		kubeConfig, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			return clientset, restClient, err
		}

	} else {
		kubeConfig, err = rest.InClusterConfig()
		if err != nil {
			return clientset, restClient, err
		}
		kubeConfig.CAFile = ""
		kubeConfig.TLSClientConfig.Insecure = true
	}
	// create the clientset
	contrailOperatorTypes.SchemeBuilder.AddToScheme(scheme.Scheme)

	crdConfig := kubeConfig
	crdConfig.ContentConfig.GroupVersion = &schema.GroupVersion{Group: contrailOperatorTypes.SchemeGroupVersion.Group, Version: contrailOperatorTypes.SchemeGroupVersion.Version}
	crdConfig.APIPath = "/apis"

	crdConfig.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}
	crdConfig.UserAgent = rest.DefaultKubernetesUserAgent()

	restClient, err = rest.UnversionedRESTClientFor(crdConfig)
	if err != nil {
		return clientset, restClient, err
	}
	clientset, err = kubernetes.NewForConfig(crdConfig)
	if err != nil {
		return clientset, restClient, err
	}
	return clientset, restClient, nil
}

func updateControlStatus(config *Config, controlStatus contrailOperatorTypes.ControlServiceStatus, clientSet *kubernetes.Clientset, restClient *rest.RESTClient) error {
	/*
		controlObjectList := contrailOperatorTypes.ControlList{}
		err := restClient.Get().Resource("controls").Do().Into(&controlObjectList)
		check(err)
		if len(controlObjectList.Items) > 0 {
			for _, controlObject := range controlObjectList.Items {
				fmt.Printf("controlResource: %s \n", controlObject.Name)
			}
		}
	*/
	//restClient := clientSet.RESTClient()
	controlObject := contrailOperatorTypes.Control{}
	err := restClient.Get().Namespace(config.Namespace).Resource("controls").Name(config.NodeName).Do().Into(&controlObject)
	check(err)

	if controlObject.Status.ServiceStatus == nil {
		var serviceStatus = make(map[string]contrailOperatorTypes.ControlServiceStatus)
		serviceStatus[config.Hostname] = controlStatus
		controlObject.Status.ServiceStatus = serviceStatus
	} else {
		controlObject.Status.ServiceStatus[config.Hostname] = controlStatus
	}
	result := contrailOperatorTypes.Control{}
	err = restClient.Put().Namespace(config.Namespace).Resource("controls").Name(config.NodeName).SubResource("status").Body(&controlObject).Do().Into(&result)

	check(err)

	controlObject = contrailOperatorTypes.Control{}
	err = restClient.Get().Namespace(config.Namespace).Resource("controls").Name(config.NodeName).Do().Into(&controlObject)
	check(err)
	fmt.Println(controlObject.Status.ServiceStatus)

	return nil
}

func getControlStatus(statusBody []byte) *contrailOperatorTypes.ControlServiceStatus {
	controlUVEStatus := &controlUVEStatus{}
	json.Unmarshal(statusBody, controlUVEStatus)
	connectionList := []contrailOperatorTypes.Connection{}

	bgpRouterList := typeSwitch(controlUVEStatus.BgpRouterState.BgpRouterIPList)
	bgpRouterConnection := contrailOperatorTypes.Connection{
		Type:  "BGPRouter",
		Nodes: bgpRouterList,
	}
	connectionList = append(connectionList, bgpRouterConnection)

	for _, connectionInfo := range controlUVEStatus.NodeStatus.ProcessStatus.List.ProcessStatus[0].ConnectionInfos.List.ConnectionInfo {
		nodeList := typeSwitch(connectionInfo.ServerAddrs.List.Element)
		connection := contrailOperatorTypes.Connection{
			Type:   connectionInfo.Type.Text,
			Name:   connectionInfo.Name.Text,
			Status: connectionInfo.Status.Text,
			Nodes:  nodeList,
		}
		connectionList = append(connectionList, connection)
	}

	staticRoutes := contrailOperatorTypes.StaticRoutes{
		Down:   controlUVEStatus.BgpRouterState.NumDownStaticRoutes[0][0].Text,
		Number: controlUVEStatus.BgpRouterState.NumStaticRoutes[0][0].Text,
	}

	bgpPeer := contrailOperatorTypes.BGPPeer{
		Up:     controlUVEStatus.BgpRouterState.NumUpBgpPeer[0][0].Text,
		Number: controlUVEStatus.BgpRouterState.NumBgpPeer[0][0].Text,
	}

	controlStatus := contrailOperatorTypes.ControlServiceStatus{
		Connections:              connectionList,
		NumberOfXMPPPeers:        controlUVEStatus.BgpRouterState.NumUpXMPPPeer[0][0].Text,
		NumberOfRoutingInstances: controlUVEStatus.BgpRouterState.NumRoutingInstance[0][0].Text,
		StaticRoutes:             staticRoutes,
		BGPPeer:                  bgpPeer,
		State:                    controlUVEStatus.NodeStatus.ProcessStatus.List.ProcessStatus[0].State.Text,
	}
	return &controlStatus
}

func typeSwitch(tst interface{}) []string {
	var nodeList []string
	switch v := tst.(type) {
	case interface{}:
		inter, ok := v.([][]interface{})
		if ok {
			for _, element := range inter {
				for _, element2 := range element {
					x, ok := element2.(map[string]interface{})
					if ok {
						y, ok := x["list"].(map[string]interface{})
						if ok {
							z, ok := y["element"].([]interface{})
							if ok {
								for _, zz := range z {
									nodeList = append(nodeList, zz.(string))
								}
							}
						}
					}
				}
			}
		} else {
			inter2, ok := v.([]interface{})
			if ok {
				for _, element := range inter2 {
					nodeList = append(nodeList, element.(string))
				}
			} else {
				inter3, ok := v.(interface{})
				if ok {
					nodeList = append(nodeList, inter3.(string))
				}
			}
		}
	case string:
		fmt.Println("String:", v)
	case [][]interface{}:
		fmt.Println("[][]interface{}:", v)
	default:
		fmt.Println("unknown")
	}
	return nodeList
}
