package goAws

import (
	"flag"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	t "github.com/tom"
)


 c := make(chan error)
type config struct {
	Aws struct {
		AcKeyID  string
		SecKeyID string
	}
}

func (cfg *config) Default() {
	cfg.Aws.AcKeyID = ""
	cfg.Aws.SecKeyID = ""
}

var gConfig config

func Init() error {
	gConfig.Default()

	err := t.LoadConfig("collection.json", &gConfig)
	if nil != err {
		t.Log(t.Error, err)
	}
	return err
}

// StopInstance ???
func StopInstance(svc ec2iface.EC2API, instanceID *string, stsInput ec2.DescribeInstancesInput) error {
	// snippet-start:[ec2.go.start_stop_instances.stop]
	input := &ec2.StopInstancesInput{
		InstanceIds: []*string{
			instanceID,
		},
		DryRun: aws.Bool(true),
	}
	_, err := svc.StopInstances(input)
	awsErr, ok := err.(awserr.Error)
	if ok && awsErr.Code() == "DryRunOperation" {
		input.DryRun = aws.Bool(false)
		_, err = svc.StopInstances(input)
		// snippet-end:[ec2.go.start_stop_instances.stop]
		if err != nil {
			return err
		}
		err = svc.WaitUntilInstanceStopped(&stsInput)
		return nil
	}

	return err
}

// StartInstance ???
func StartInstance(svc ec2iface.EC2API, instanceID *string, stsInput ec2.DescribeInstancesInput) (string, error) {
	// snippet-start:[ec2.go.start_stop_instances.start]
	//	var waitGroup sync.WaitGroup

	input := &ec2.StartInstancesInput{
		InstanceIds: []*string{
			instanceID,
		},
		DryRun: aws.Bool(true),
	}
	_, err := svc.StartInstances(input)
	awsErr, ok := err.(awserr.Error)

	if ok && awsErr.Code() == "DryRunOperation" {
		// Set DryRun to be false to enable starting the instances
		input.DryRun = aws.Bool(false)
		_, err := svc.StartInstances(input)
		// snippet-end:[ec2.go.start_stop_instances.start]
		if err != nil {
			return "", err
		}

		err = svc.WaitUntilInstanceRunning(&stsInput)
		description, err := svc.DescribeInstances(&stsInput)
		publicIp := description.Reservations[0].Instances[0].PublicIpAddress
		publicIP := *publicIp
		return publicIP, nil
	}

	return "", err
}

func RestartEc2(insID, accessKey, secretKey, region string) (string, error) {
	var publicIp string
	var err error
	instanceID := flag.String("i", insID, "the id of instance to reboot")

	flag.Parse()

	if *instanceID == "" {
		fmt.Println("(-s START | STOP -i INSTANCE-ID")
		return "", nil
	}
	// setting configs
	//access key, secret key token key(optional)
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			CredentialsChainVerboseErrors: aws.Bool(true),
			Credentials:                   credentials.NewStaticCredentials(accessKey, secretKey, ""),
			Region:                        aws.String(region),
		},
	}))

	svc := ec2.New(sess)
	// snippet-end:[ec2.go.start_stop_instances.session]
	statusInput := ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			instanceID,
		},
	}
	if err = StopInstance(svc, instanceID, statusInput); nil == err {
		if publicIp, err = StartInstance(svc, instanceID, statusInput); nil == err {
			return publicIp, err
		}
	}
	return "", err
}
