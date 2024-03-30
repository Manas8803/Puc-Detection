package main

import (
	"log"
	"os"

	event "github.com/Manas8803/puc-detection/deploy-scripts/events"
	"github.com/Manas8803/puc-detection/deploy-scripts/roles"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	dynamodb "github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/joho/godotenv"
)

type PucDetectionStackProps struct {
	awscdk.StackProps
}

func NewPucDetectionStack(scope constructs.Construct, id string, props *PucDetectionStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	//^ Vehicle-TABLE
	vehicle_table := dynamodb.NewTable(stack, jsii.String("vehicle-table"), &dynamodb.TableProps{
		PartitionKey: &dynamodb.Attribute{
			Name: jsii.String("reg_no"),
			Type: dynamodb.AttributeType_STRING,
		},
		TableName: jsii.String("PUC-Detection-vehicle-table"),
	})

	//^ User-TABLE
	user_table := dynamodb.NewTable(stack, jsii.String("user-table"), &dynamodb.TableProps{
		PartitionKey: &dynamodb.Attribute{
			Name: jsii.String("email"),
			Type: dynamodb.AttributeType_STRING,
		},
		TableName: jsii.String("PUC-Detection-user-table-1"),
	})

	//^ Log group of vrc handler
	logGroup_vrc := awslogs.NewLogGroup(stack, jsii.String("VRC-LogGroup"), &awslogs.LogGroupProps{
		LogGroupName:  jsii.String("/aws/lambda/PucDetectionStack-VRC"),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	//^ VRC handler
	vrc_handler := awslambda.NewFunction(stack, jsii.String("VRC-Lambda"), &awslambda.FunctionProps{
		Code:    awslambda.Code_FromAsset(jsii.String("../vrc-service"), nil),
		Runtime: awslambda.Runtime_PROVIDED_AL2023(),
		Handler: jsii.String("main"),
		Timeout: awscdk.Duration_Seconds(jsii.Number(10)),
		Role:    roles.CreateVRCHandlerRole(stack, vehicle_table),
		Environment: &map[string]*string{
			"REGION":            jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
			"API_KEY":           jsii.String(os.Getenv("API_KEY")),
			"VEHICLE_TABLE_ARN": jsii.String(*vehicle_table.TableArn()),
		},
		FunctionName: jsii.String("PUC-Detection-VRC-Lambda"),
		LogGroup:     logGroup_vrc,
	})

	//^ Log group of reg_ex_cron_job handler
	logGroup_reg_ex := awslogs.NewLogGroup(stack, jsii.String("RegExpCronJob-LogGroup"), &awslogs.LogGroupProps{
		LogGroupName:  jsii.String("/aws/lambda/PucDetectionStack-RegExCronJob"),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	//^ Registration expiration cron job handler
	reg_exp_cron_job := awslambda.NewFunction(stack, jsii.String("RegExpCronJob-Lambda"), &awslambda.FunctionProps{
		Code:    awslambda.Code_FromAsset(jsii.String("../reg_expiration_job-service"), nil),
		Runtime: awslambda.Runtime_PROVIDED_AL2023(),
		Handler: jsii.String("main"),
		Timeout: awscdk.Duration_Seconds(jsii.Number(10)),
		Role:    roles.CreateRegExpCronJobRole(stack, vrc_handler, vehicle_table),
		Environment: &map[string]*string{
			"REGION":            jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
			"VEHICLE_TABLE_ARN": jsii.String(*vehicle_table.TableArn()),
			"VRC_HANDLER_ARN":   jsii.String(*vrc_handler.FunctionArn()),
		},
		FunctionName: jsii.String("PUC-Detection-RegExp-Lambda"),
		LogGroup:     logGroup_reg_ex,
	})

	event.CreateDailyScheduler(stack, reg_exp_cron_job)

	//^ Log group of reg_renewal_reminder handler
	logGroup_reg := awslogs.NewLogGroup(stack, jsii.String("Reg_Renewal_Reminder-LogGroup"), &awslogs.LogGroupProps{
		LogGroupName:  jsii.String("/aws/lambda/PucDetectionStack-Reg_Renewal_Reminder"),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	//^ Reg_Renewal_Reminder handler
	reg_renewal_reminder_handler := awslambda.NewFunction(stack, jsii.String("Reg_Renewal_Reminder-Lambda"), &awslambda.FunctionProps{
		Code:    awslambda.Code_FromAsset(jsii.String("../reg_renewal_reminder-service"), nil),
		Runtime: awslambda.Runtime_PROVIDED_AL2023(),
		Handler: jsii.String("main"),
		Timeout: awscdk.Duration_Seconds(jsii.Number(10)),
		Role:    roles.CreateRegReminderHandlerRole(stack, vehicle_table),
		Environment: &map[string]*string{
			"REGION":            jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
			"VRC_HANDLER_ARN":   jsii.String(*vrc_handler.FunctionArn()),
			"VEHICLE_TABLE_ARN": jsii.String(*vehicle_table.TableArn()),
		},
		FunctionName: jsii.String("PUC-Detection-Reg_Renewal_Reminder-Lambda"),
		LogGroup:     logGroup_reg,
	})

	//^ Log group of ocr handler
	logGroup_ocr := awslogs.NewLogGroup(stack, jsii.String("OCR-Lambda-LogGroup"), &awslogs.LogGroupProps{
		LogGroupName:  jsii.String("/aws/lambda/PucDetectionStack-OCRLambda"),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	//^ Ocr handler
	awslambda.NewFunction(stack, jsii.String("OCR-Lambda"), &awslambda.FunctionProps{
		Code:    awslambda.Code_FromAsset(jsii.String("../ocr-service"), nil),
		Runtime: awslambda.Runtime_PROVIDED_AL2023(),
		Handler: jsii.String("main"),
		Timeout: awscdk.Duration_Seconds(jsii.Number(10)),
		Role:    roles.CreateOCRHandlerRole(stack, vrc_handler),
		Environment: &map[string]*string{
			"REGION":          jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
			"REG_RENEWAL_ARN": jsii.String(*reg_renewal_reminder_handler.FunctionArn()),
		},
		LogGroup:     logGroup_ocr,
		FunctionName: jsii.String("PUC-Detection-OCR-Lambda"),
	})

	auth_handler := awslambda.NewFunction(stack, jsii.String("auth-service"), &awslambda.FunctionProps{
		Code:    awslambda.Code_FromAsset(jsii.String("../auth-service"), nil),
		Runtime: awslambda.Runtime_PROVIDED_AL2023(),
		Handler: jsii.String("main"),
		Timeout: awscdk.Duration_Seconds(jsii.Number(10)),
		Environment: &map[string]*string{
			"JWT_SECRET_KEY": jsii.String(os.Getenv("JWT_SECRET_KEY")),
			"JWT_LIFETIME":   jsii.String(os.Getenv("JWT_LIFETIME")),
			"EMAIL":          jsii.String(os.Getenv("EMAIL")),
			"PASSWORD":       jsii.String(os.Getenv("PASSWORD")),
			"ADMIN":          jsii.String(os.Getenv("ADMIN")),
			"USER_TABLE_ARN": jsii.String(*user_table.TableArn()),
		},
		Role: roles.CreateDbRole(stack, user_table),
	})

	awsapigateway.NewLambdaRestApi(stack, jsii.String("Puc_Detection_Auth"), &awsapigateway.LambdaRestApiProps{
		Handler: auth_handler,
	})

	return stack

}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewPucDetectionStack(app, "PucDetectionStack", &PucDetectionStackProps{
		awscdk.StackProps{
			StackName: jsii.String("PucDetectionStack"),
			Env:       env(),
		},
	})

	app.Synth(nil)
}

func env() *awscdk.Environment {

	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatalln("Error loading .env file : ", err)
	}

	return &awscdk.Environment{
		Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
		Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	}

}