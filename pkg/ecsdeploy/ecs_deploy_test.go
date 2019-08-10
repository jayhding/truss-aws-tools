package ecsdeploy

import (
	"errors"
	// "fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

////////////////////
// Shared mocks
////////////////////
type ecsMock struct {
	mock.Mock
	ecsiface.ECSAPI
}

func (m *ecsMock) UpdateService(upi *ecs.UpdateServiceInput) (*ecs.UpdateServiceOutput, error) {
	args := m.Called(upi)
	return args.Get(0).(*ecs.UpdateServiceOutput), args.Error(1)
}

func (m *ecsMock) DescribeServices(input *ecs.DescribeServicesInput) (*ecs.DescribeServicesOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*ecs.DescribeServicesOutput), args.Error(1)
}

func (m *ecsMock) DescribeTaskDefinition(input *ecs.DescribeTaskDefinitionInput) (*ecs.DescribeTaskDefinitionOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*ecs.DescribeTaskDefinitionOutput), args.Error(1)

}

////////////////////
// Testing Func UpdateService
////////////////////

type GetTaskDefinitionSuite struct {
	suite.Suite
	ecsMockClient *ecsMock
	e             ECSClusterServiceDeployer
	taskDef       string
}

func TestGetTaskDefinitionSuite(t *testing.T) {
	suite.Run(t, new(GetTaskDefinitionSuite))
}

func (suite *GetTaskDefinitionSuite) SetupTest() {
	var logger *zap.Logger // not for sure what to do about zap
	logger, _ = zap.NewDevelopment()
	suite.ecsMockClient = new(ecsMock)
	suite.e = ECSClusterServiceDeployer{
		Logger:     logger,
		ECSCluster: "mycluster",
		ECSService: "myservice",
		ECSClient:  suite.ecsMockClient,
	}
	suite.taskDef = "thisismytaskdef"
}

// TestGetServiceTaskDefinition testing normal operation
func (suite *GetTaskDefinitionSuite) TestGetServiceTaskDefinition() {
	serviceOutput := ecs.DescribeServicesOutput{
		Services: []*ecs.Service{
			&ecs.Service{
				TaskDefinition: &suite.taskDef,
			}},
	}

	taskDefOutput := ecs.DescribeTaskDefinitionOutput{
		TaskDefinition: &ecs.TaskDefinition{
			TaskDefinitionArn: &suite.taskDef,
		},
	}

	suite.ecsMockClient.On("DescribeServices", mock.Anything).Once().Return(&serviceOutput, nil)
	suite.ecsMockClient.On("DescribeTaskDefinition", mock.Anything).Once().Return(&taskDefOutput, nil)

	resp, err := suite.e.GetServiceTaskDefinition()
	assert.Equal(suite.T(), resp.TaskDefinitionArn, &suite.taskDef)
	assert.Nil(suite.T(), err)
	suite.ecsMockClient.AssertExpectations(suite.T())
}

// TestGetServiceTaskDefinitionServicesFailed test when the DescribeServices call fails
func (suite *GetTaskDefinitionSuite) TestGetServiceTaskDefinitionServicesFailed() {
	output := ecs.DescribeServicesOutput{}
	suite.ecsMockClient.On("DescribeServices", mock.Anything).Once().Return(&output, errors.New("poof AWS died"))

	_, err := suite.e.GetServiceTaskDefinition()
	assert.Error(suite.T(), err)
	suite.ecsMockClient.AssertExpectations(suite.T())
}

// TestGetServiceTaskDefinitionNoMatchingServices test when service serach return no matching services
func (suite *GetTaskDefinitionSuite) TestGetServiceTaskDefinitionNoMatchingServices() {
	output := ecs.DescribeServicesOutput{
		Services: make([]*ecs.Service, 0),
	}
	suite.ecsMockClient.On("DescribeServices", mock.Anything).Once().Return(&output, nil)

	_, err := suite.e.GetServiceTaskDefinition()
	assert.Error(suite.T(), err)
	suite.ecsMockClient.AssertExpectations(suite.T())
}

func (suite *GetTaskDefinitionSuite) TestGetServiceTaskDefinitionDescribeTaskFails() {
	serviceOutput := ecs.DescribeServicesOutput{
		Services: []*ecs.Service{
			&ecs.Service{
				TaskDefinition: &suite.taskDef,
			}},
	}

	taskDefOutput := ecs.DescribeTaskDefinitionOutput{
		TaskDefinition: &ecs.TaskDefinition{
			TaskDefinitionArn: &suite.taskDef,
		},
	}

	suite.ecsMockClient.On("DescribeServices", mock.Anything).Once().Return(&serviceOutput, nil)
	suite.ecsMockClient.On("DescribeTaskDefinition", mock.Anything).Once().Return(&taskDefOutput, errors.New("aws failed"))

	_, err := suite.e.GetServiceTaskDefinition()
	assert.Error(suite.T(), err)
	suite.ecsMockClient.AssertExpectations(suite.T())
}

////////////////////
// Testing Func UpdateService
////////////////////

type UpdateServiceSuite struct {
	suite.Suite
	ecsMockClient *ecsMock
	e             ECSClusterServiceDeployer
}

func (suite *UpdateServiceSuite) SetupTest() {
	var logger *zap.Logger // not for sure what to do about zap
	logger, _ = zap.NewDevelopment()
	suite.ecsMockClient = new(ecsMock)
	suite.e = ECSClusterServiceDeployer{
		Logger:     logger,
		ECSCluster: "mycluster",
		ECSService: "myservice",
		ECSClient:  suite.ecsMockClient,
	}
}

func TestUpdateServiceSuite(t *testing.T) {
	suite.Run(t, new(UpdateServiceSuite))
}

// TestUpdateService tests the case where the ECS API are a success
func (suite *UpdateServiceSuite) TestUpdateService() {
	service := new(ecs.Service)
	output := ecs.UpdateServiceOutput{
		Service: service,
	}
	suite.ecsMockClient.On("UpdateService", mock.Anything).Return(&output, nil)

	resp, err := suite.e.UpdateService("here is my taskDefin")
	assert.Equal(suite.T(), resp, service)
	assert.Nil(suite.T(), err)

	suite.ecsMockClient.AssertExpectations(suite.T())
}

// TestUpdateServiceError tests the case where the ECS API call throws an error
func (suite *UpdateServiceSuite) TestUpdateServiceError() {
	output := ecs.UpdateServiceOutput{
		Service: nil,
	}
	suite.ecsMockClient.On("UpdateService", mock.Anything).Return(&output, errors.New("poof AWS died"))

	resp, err := suite.e.UpdateService("here is my taskDefin")
	assert.Nil(suite.T(), resp)
	assert.Error(suite.T(), err)

	suite.ecsMockClient.AssertExpectations(suite.T())
}

// ////////////////////
// // Test Data
// ////////////////////

// var goodMatchingContainerMap = map[string]map[string]string{
// 	"atlantis":            {"image": "updated/imagepath:latest"},
// 	"nginxbutwithracoons": {"image": "nginxbutwithracoons/imagepath:latest"},
// }

// var goodNonMatchingContainerMap = map[string]map[string]string{
// 	"nginx":               {"image": "nginx/imagepath:latest"},
// 	"nginxbutwithracoons": {"image": "nginxbutwithracoons/imagepath:latest"},
// }

// var emptyContainerMap = map[string]map[string]string{}

// var goodContainerDef = &ecs.ContainerDefinition{
// 	Name:                   aws.String("atlantis"),
// 	Image:                  aws.String("runatlantis/atlantis:latest"),
// 	Cpu:                    aws.Int64(256),
// 	Memory:                 aws.Int64(512),
// 	MemoryReservation:      aws.Int64(128),
// 	PortMappings:           nil,
// 	Essential:              nil,
// 	Environment:            nil,
// 	MountPoints:            nil,
// 	VolumesFrom:            nil,
// 	Secrets:                nil,
// 	ReadonlyRootFilesystem: aws.Bool(false),
// 	LogConfiguration:       nil,
// }

// var goodTaskDefinition = &ecs.TaskDefinition{
// 	ContainerDefinitions:    []*ecs.ContainerDefinition{goodContainerDef},
// 	Family:                  aws.String("atlantis"),
// 	TaskRoleArn:             aws.String("arn:aws:iam::accountID:role/atlantis-ecs_task_execution"),
// 	ExecutionRoleArn:        aws.String("arn:aws:iam::accountID:role/atlantis-ecs_task_execution"),
// 	NetworkMode:             aws.String("awsvpc"),
// 	Revision:                aws.Int64(000),
// 	Status:                  aws.String("ACTIVE"),
// 	RequiresCompatibilities: []*string{aws.String("FARGATE")},
// 	Cpu:                     aws.String("256"),
// 	Memory:                  aws.String("512"),
// }

// var goodTaskDefResponse = &ecs.RegisterTaskDefinitionOutput{
// 	TaskDefinition: goodTaskDefinition,
// }

// ////////////////////
// // ECS Client Mock
// ////////////////////
// type mockECSClient struct {
// 	ecsiface.ECSAPI
// }

// func (m *mockECSClient) DescribeServices(input *ecs.DescribeServicesInput) (*ecs.DescribeServicesOutput, error) {
// 	//do the mocking
// 	return nil, nil
// }

// func (m *mockECSClient) DescribeTaskDefinition(input *ecs.DescribeTaskDefinitionInput) (*ecs.DescribeTaskDefinitionOutput, error) {
// 	//do the mocking
// 	return nil, nil
// }

// // You must mock this one for the thing you NEED to test
// func (m *mockECSClient) RegisterTaskDefinition(input *ecs.RegisterTaskDefinitionInput) (*ecs.RegisterTaskDefinitionOutput, error) {
// 	var taskDefinition = &ecs.TaskDefinition{
// 		ContainerDefinitions: input.ContainerDefinitions,
// 		Family:               input.Family,
// 		TaskRoleArn:          input.TaskRoleArn,
// 		ExecutionRoleArn:     input.TaskRoleArn,
// 		NetworkMode:          input.NetworkMode,
// 		// technically revision should be different this just makes it easier to test
// 		Revision:                aws.Int64(000),
// 		Status:                  aws.String("ACTIVE"),
// 		RequiresCompatibilities: input.RequiresCompatibilities,
// 		Cpu:                     input.Cpu,
// 		Memory:                  input.Memory,
// 	}

// 	output := &ecs.RegisterTaskDefinitionOutput{
// 		TaskDefinition: taskDefinition,
// 	}
// 	return output, nil
// }

// func (m *mockECSClient) UpdateService(input *ecs.UpdateServiceInput) (*ecs.UpdateServiceOutput, error) {
// 	//do the mocking
// 	return nil, nil
// }

// /////////////////
// // Tests
// /////////////////
// var logger, _ = zap.NewProduction()
// var mockClient = &mockECSClient{}

// func TestRegisterUpdatedTaskDefinition(t *testing.T) {
// 	// Setup for the test
// 	e := ECSClusterServiceDeployer{
// 		ECSCluster: "test",
// 		ECSService: "atlantis",
// 		Logger:     logger,
// 		ECSClient:  mockClient,
// 	}

// 	fmt.Println(e)
// 	// // Empty Container map
// 	// fmt.Println("Test empty container map")
// 	// fmt.Println("+++++++++++++++++++")
// 	// taskDefinition, err := e.RegisterUpdatedTaskDefinition(goodTaskDefinition, emptyContainerMap)
// 	// //	if !reflect.DeepEqual(taskDefinition, goodTaskDefinition) {
// 	// //		t.Errorf("ERROR: The task definition changed.")
// 	// //		fmt.Println(taskDefinition)
// 	// //		fmt.Println(goodTaskDefinition)
// 	// //		fmt.Println(err)
// 	// //	}
// 	// fmt.Println("did this mutate?")
// 	// fmt.Println(taskDefinition)
// 	// fmt.Println(err)

// 	// // Best case
// 	// fmt.Println("Test normal case")
// 	// fmt.Println("+++++++++++++++++++")
// 	// taskDefinition, err = e.RegisterUpdatedTaskDefinition(goodTaskDefinition, goodMatchingContainerMap)
// 	// fmt.Println(taskDefinition)
// 	// fmt.Println(goodTaskDefinition)
// 	// fmt.Println(err)
// 	// if !reflect.DeepEqual(taskDefinition, goodTaskDefinition) {
// 	// 	t.Errorf("ERROR: The task definition changed.")
// 	// }
// 	// fmt.Println("did this mutate?")
// 	// fmt.Println(goodTaskDefinition)
// 	//
