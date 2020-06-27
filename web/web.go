package web

import (
	"encoding/json"
	"fmt"
	"github.com/jabolina/go-mcast/pkg/mcast"
	"io/ioutil"
	"net/http"
)

type peer struct {
	Name string `json:name`
}

type clusterConfig struct {
	Partitions   []peer `json:partitions`
}

type Server struct {
	name string
	config clusterConfig
	unity mcast.Unity
}

// Creates a new structure that holds information about
// all other partitions and about the HTTP server.
func NewServer(confPath string, name string) (*Server, error) {
	buf, err := ioutil.ReadFile(confPath)
	if err != nil {
		return nil, err
	}

	var config clusterConfig
	if err = json.Unmarshal(buf, &config); err != nil {
		return nil, err
	}

	partitionName := mcast.CreatePartitionName(name)
	conf := mcast.DefaultConfiguration(partitionName)
	conf.Logger.ToggleDebug(true)
	unity, err := mcast.NewMulticastConfigured(conf)
	if err != nil {
		return nil, err
	}

	conf.Logger.Infof("using configuration. %v", config)
	return &Server{config: config, name: name, unity: unity}, nil
}

func (s Server) destinations() []string {
	var partitions []string
	for _, partition := range s.config.Partitions {
		partitions = append(partitions, partition.Name)
	}
	return partitions
}

func (s *Server) GetRequest(writer http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	key := request.Form.Get("key")
	writer.WriteHeader(200)
	read := mcast.NewReadRequest([]byte(key), []string{s.name})
	res, err := s.unity.Read(*read)
	if err != nil {
		writer.WriteHeader(500)
		fmt.Fprintf(writer, "error = %v", err)
		return
	}

	if !res.Success {
		writer.WriteHeader(500)
		fmt.Fprintf(writer, "not success", err)
		return
	}

	fmt.Fprintf(writer, "[key = %s, value = %v]", key, string(res.Data))
}

func (s *Server) SetRequest(writer http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	key := request.Form.Get("key")
	value := request.Form.Get("value")
	write := mcast.NewWriteRequest([]byte(key), []byte(value), nil, s.destinations())
	_, err := s.unity.Write(*write)
	if err != nil {
		writer.WriteHeader(500)
		fmt.Fprintf(writer, "error = %v", err)
		return
	}
	writer.WriteHeader(200)
	fmt.Fprintf(writer, "[key = %s, value = %s]", key, value)
}
