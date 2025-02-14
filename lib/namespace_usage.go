package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/utils"
)

type NamespaceCosts struct {
	Namespace map[string]struct {
		Breakdown map[string]float32 `json:"breakdown"`
		Total     float32            `json:"total"`
	} `json:"namespace"`
	LastUpdated string `json:"last_updated"`
}

type NamespaceUsage struct {
	Data []struct {
		Requested struct {
			CPU    int `json:"CPU"`
			Memory int `json:"Memory"`
			Pods   int `json:"Pods"`
		} `json:"Requested"`
		Used struct {
			CPU    int `json:"CPU"`
			Memory int `json:"Memory"`
			Pods   int `json:"Pods"`
		} `json:"Used"`
		Hardlimits struct {
			CPU    int `json:"CPU"`
			Memory int `json:"Memory"`
			Pods   int `json:"Pods"`
		} `json:"Hardlimits"`
		ContainerCount int    `json:"ContainerCount"`
		Name           string `json:"Name"`
	} `json:"data"`
	LastUpdated string `json:"updated_at"`
}

type Usage struct {
	Namespace string
	Breakdown map[string]float32
	Total     float32

	CPU struct {
		Requested  int
		Used       int
		HardLimits int
	}
	Memory struct {
		Requested  int
		Used       int
		HardLimits int
	}
	Pods struct {
		Requested  int
		Used       int
		HardLimits int
	}
	Name           string
	ContainerCount int
	LastUpdated    string
	Tags           struct {
		Environment  string
		Team         string
		Application  string
		SlackChannel string
		SourceCode   string
		DomainNames  string
	}
}

func NamespaceUsagePage(w http.ResponseWriter, bucket, namespace string, wantJson bool, client *s3.Client) {
	t := template.Must(template.ParseFiles("lib/templates/namespaces.html"))

	byteValue, filestamp, err := utils.ImportS3File(client, bucket, "namespace_costs.json")
	if err != nil {
		fmt.Println(err)
	}
	if wantJson {
		w.Header().Set("Content-Type", "application/json")
		w.Write(byteValue)
		return
	}

	var namespaceCosts NamespaceCosts
	json.Unmarshal(byteValue, &namespaceCosts)
	namespaceCosts.LastUpdated = filestamp

	byteValue, filestamp, err = utils.ImportS3File(client, bucket, "namespace_usage.json")
	if err != nil {
		fmt.Println(err)
	}
	if wantJson {
		w.Header().Set("Content-Type", "application/json")
		w.Write(byteValue)
		return
	}

	var namespaceUsage NamespaceUsage
	json.Unmarshal(byteValue, &namespaceUsage)
	namespaceUsage.LastUpdated = filestamp

	var usage Usage
	for ns, v := range namespaceCosts.Namespace {
		if ns == namespace {
			usage.Namespace = ns
			usage.Breakdown = v.Breakdown
			usage.Total = v.Total
		}
	}

	for _, v := range namespaceUsage.Data {
		if v.Name == namespace {
			usage.CPU.Requested = v.Requested.CPU
			usage.CPU.Used = v.Used.CPU
			usage.CPU.HardLimits = v.Hardlimits.CPU

			usage.Memory.Requested = v.Requested.Memory
			usage.Memory.Used = v.Used.Memory
			usage.Memory.HardLimits = v.Hardlimits.Memory

			usage.Pods.Requested = v.Requested.Pods
			usage.Pods.Used = v.Used.Pods
			usage.Pods.HardLimits = v.Hardlimits.Pods

			usage.ContainerCount = v.ContainerCount
			usage.Name = v.Name
			usage.LastUpdated = namespaceUsage.LastUpdated
		}
	}

	fmt.Println(usage)

	if err := t.ExecuteTemplate(w, "namespaces.html", usage); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
