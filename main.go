package main

import (
	"time"
	"os"
	"fmt"
	"strconv"
	"log"
	"math/rand"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type TQuot struct {
	Slug     string `json:"Slug"`
	Category string `json:"Category"`
	Number   int    `json:"Number"`
	Content  string `json:"Content"`
	From     string `json:"From"`
}

type TCounter struct {
	Category string `json:"Category"`
	Count    int    `json:"Count"`
}

/*
func HandleRequest(ctx context.Context, name MyEvent) (string, error) {
	category := 
	return fmt.Sprintf("Hello %s!", name.Name ), nil
}
*/

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	category := request.QueryStringParameters["c"]
	//num := request.QueryStringParameters["n"]

	slug := request.QueryStringParameters["s"]

	var quots []TQuot
	if slug != "" {
		quots = getDetailQuot(slug)
	} else if category != "" {
		num := randomQuotation(category)
		quots = getQuotByCatNum(category, num)
	} else {
		quots = getListQuot()
	}

	jsonBytes, _ := json.Marshal(quots)

	return events.APIGatewayProxyResponse{
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "http://localhost:3000",
			"Access-Control-Allow-Headers": "origin,Accept,Authorization,Content-Type",
			"Content-Type":                 "application/json; charset=utf-8",
		},
        Body:       string(jsonBytes),
        StatusCode: 200,
    }, nil
}

func main() {
	lambda.Start(handler)
}

func AccessDB() *dynamodb.DynamoDB {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	db := dynamodb.New(sess, &aws.Config{
		Region:   aws.String("ap-northeast-1"),
		Endpoint: aws.String(os.Getenv("DYNAMO_ENDPOINT")),
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			os.Getenv("AWS_SESSION_TOKEN"),
		),
	})

	return db
}

func getQuotByCatNum(cat string, num int) []TQuot {
	db := AccessDB()
	input := &dynamodb.QueryInput{
		TableName: aws.String("Quotations"),
		IndexName: aws.String("Cat-Num-Index"),
		ExpressionAttributeNames: map[string]*string{
			"#Category": aws.String("Category"),
			"#Number":   aws.String("Number"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":cat": {
				S: aws.String(cat),
			},
			":num": {
				N: aws.String(strconv.Itoa(num)),
			},
		},
		KeyConditionExpression: aws.String("#Category = :cat AND #Number = :num"),
	}

	res, err := db.Query(input)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	quot := TQuot{}
	err = dynamodbattribute.UnmarshalMap(res.Items[0], &quot)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	ret := []TQuot{quot}

	return ret
}

func randomQuotation(cat string) int {
	count := DetailCounter(cat)

	var ret int
	for i := 0; i < 5; i++ {
		rand.Seed(time.Now().UnixNano())
		var n int
		n = rand.Intn(count) + 1
		quots := getQuotByCatNum(cat, n)
		if quots[0].Number != 0 {
			ret = quots[0].Number
			break
		}
	}

	return ret
}

// all
func getListQuot() []TQuot {
	db := AccessDB()
	tableName := "Quotations"
	input := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}

	scanOut, err := db.Scan(input)
	if err != nil {
		log.Fatalf("Query API call failed: %s", err)
		return nil
	}

	var quots []TQuot
	for _, n := range scanOut.Items {
		var quotTemp TQuot
		_ = dynamodbattribute.UnmarshalMap(n, &quotTemp)
		quots = append(quots, quotTemp)
	}

	return quots
}

// by slug
func getDetailQuot(slug string) []TQuot {
	db := AccessDB()
	input := &dynamodb.GetItemInput{
		TableName: aws.String("Quotations"),
		Key: map[string]*dynamodb.AttributeValue{
			"Slug": {
				S: aws.String(slug),
			},
		},
	}

	res, err := db.GetItem(input)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	quot := TQuot{}
	err = dynamodbattribute.UnmarshalMap(res.Item, &quot)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	ret := []TQuot{quot}
	return ret
}

func DetailCounter(cat string) int {
	db := AccessDB()
	input := &dynamodb.GetItemInput{
		TableName: aws.String("Counters"),
		Key: map[string]*dynamodb.AttributeValue{
			"Category": {
				S: aws.String(cat),
			},
		},
	}

	res, err := db.GetItem(input)
	if err != nil {
		fmt.Println(err.Error())
		return -1
	}
	count := TCounter{}
	err = dynamodbattribute.UnmarshalMap(res.Item, &count)
	if err != nil {
		fmt.Println(err.Error())
		return -1
	}

	return count.Count
}
