package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	ceTypes "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/ministryofjustice/cloud-platform-environments/pkg/authenticate"
	"github.com/ministryofjustice/cloud-platform-environments/pkg/namespace"
)

var (
	bucket         = flag.String("bucket", os.Getenv("KUBECONFIG_S3_BUCKET"), "AWS S3 bucket for kubeconfig")
	ctx            = flag.String("context", "live.cloud-platform.service.justice.gov.uk", "Kubernetes context specified in kubeconfig")
	hoodawApiKey   = flag.String("hoodawAPIKey", os.Getenv("HOODAW_API_KEY"), "API key to post data to the 'How out of date are we' API")
	hoodawEndpoint = flag.String("hoodawEndpoint", "/namespace_costs", "Endpoint to send the data to")
	hoodawHost     = flag.String("hoodawHost", os.Getenv("HOODAW_HOST"), "Hostname of the 'How out of date are we' API")
	kubeconfig     = flag.String("kubeconfig", "kubeconfig", "Name of kubeconfig file in S3 bucket")
	region         = flag.String("region", os.Getenv("AWS_REGION"), "AWS Region")

	endPoint = *hoodawHost + *hoodawEndpoint
)

const SHARED_COSTS string = "SHARED_COSTS"

func main() {
	flag.Parse()

	awsCostUsageData, err := GetAwsCostAndUsageData()
	if err != nil {
		log.Fatalln(err.Error())
	}
	// Get all namespaces from cluster

	// Gain access to a Kubernetes cluster using a config file stored in an S3 bucket.
	clientset, err := authenticate.CreateClientFromS3Bucket(*bucket, *kubeconfig, *region, *ctx)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// Get the list of namespaces from the cluster which is set in the clientset
	_, err = namespace.GetAllNamespacesFromCluster(clientset)
	if err != nil {
		log.Fatalln(err.Error())
	}

	_ = costsByNamespace(awsCostUsageData)

	// get the value of shared costs from the aws data and
	// delete the shared costs from each of namespace.
	// divide the shared costs by number of namespace and assign the cost back to per namespace
	// add shared team costs
	//

}

func GetAwsCostAndUsageData() ([][]string, error) {

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		// handle error
	}
	svc := costexplorer.NewFromConfig(cfg)
	now, monthBefore := timeNow(31)

	param := &costexplorer.GetCostAndUsageInput{
		Granularity: ceTypes.GranularityMonthly,
		TimePeriod: &ceTypes.DateInterval{
			Start: aws.String(monthBefore),
			End:   aws.String(now),
		},
		Metrics: []string{"BlendedCost"},
		GroupBy: []ceTypes.GroupDefinition{
			{
				Type: ceTypes.GroupDefinitionTypeDimension,
				Key:  aws.String("SERVICE"),
			},
			{
				Type: ceTypes.GroupDefinitionTypeTag,
				Key:  aws.String("namespace"),
			},
		},
	}

	GetCostAndUsageOutput, err := svc.GetCostAndUsage(context.TODO(), param)
	if err != nil {
		fmt.Println(err)
	}

	var resultsCosts [][]string
	for _, results := range GetCostAndUsageOutput.ResultsByTime {
		startDate := *results.TimePeriod.Start
		for _, groups := range results.Groups {
			for _, metrics := range groups.Metrics {
				tag_value := strings.Split(groups.Keys[1], "$")
				if tag_value[1] == "" {
					tag_value[1] = SHARED_COSTS
				}
				info := []string{startDate, groups.Keys[0], tag_value[1], *metrics.Amount}

				resultsCosts = append(resultsCosts, info)

			}
		}
	}
	return resultsCosts, nil
}

func timeNow(x int) (string, string) {
	dt := time.Now()
	now := dt.Format("2006-01-02")
	month := dt.AddDate(0, 0, -x).Format("2006-01-02")
	return now, month
}

// Use repository interface isntead https://blog.canopas.com/approach-to-avoid-accessing-variables-globally-in-golang-2019b234762
var costsPerNamespaceMap = map[string]map[string]float64{}

func costsByNamespace(awsCostUsageData [][]string) map[string]map[string]float64 {
	service := make(map[string]float64, 0)
	for _, col := range awsCostUsageData {

		// just test with example namespace
		if col[2] == "prisoner-content-hub-production" {
			cost, err := strconv.ParseFloat(col[3], 64)
			if err != nil {
				fmt.Println(err)
				return nil
			}
			if existing, ok := costsPerNamespaceMap[col[2]][col[1]]; ok {
				fmt.Println("existing", col[2], "resource", col[1], "cost", costsPerNamespaceMap[col[2]][col[1]])
				service[col[1]] = cost + existing
			} else {
				service[col[1]] = cost
			}
			costsPerNamespaceMap[col[2]] = service
		}

	}

	for k, v := range costsPerNamespaceMap {
		fmt.Println("key[%s] value[%s]\n", k, v)
	}
	return costsPerNamespaceMap
}