package syncer

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var app Syncer

func Init() {
	ctx := context.Background()
	client, err := getAwsClient(ctx)
	if err != nil {
		panic(err)
	}
	app.FolderPath = "C:\\Users\\pratersm\\Documents\\WebSites"
	app.Bucket = "pratermade-gotest"
	app.S3Client = client
	err = app.InitDb("..\\manifest.db")
	if err != nil {
		panic(err)
	}
}

func TestWalkAndHash(t *testing.T) {
	Init()

	testFilter := []string{"jpg", "txt"}
	ret, err := app.WalkAndHash(testFilter)
	if err != nil {
		t.Fatal(err)
	}
	_ = ret

}

func TestSelectUploadList(t *testing.T) {
	Init()
	_, err := app.GetUploadList()
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpdateRecords(t *testing.T) {
	Init()
	testFilter := []string{"jpg"}
	retMap, err := app.WalkAndHash(testFilter)
	if err != nil {
		t.Fatal(err)
	}
	err = app.UpdateManifest(retMap)
	if err != nil {
		t.Fatal(err)
	}

}

func TestRecordParts(t *testing.T) {
	Init()
	parts := []string{
		"C:\\Users\\pratersm\\AppData\\Local\\Temp\\s3sync3030611028\\Battlestar Galactica (2003)  S04e19e20  Daybreak (1080P Bluray X265 Rzerox)-1.mp4.part0",
		"C:\\Users\\pratersm\\AppData\\Local\\Temp\\s3sync3030611028\\Battlestar Galactica (2003)  S04e19e20  Daybreak (1080P Bluray X265 Rzerox)-1.mp4.part1",
		"C:\\Users\\pratersm\\AppData\\Local\\Temp\\s3sync3030611028\\Battlestar Galactica (2003)  S04e19e20  Daybreak (1080P Bluray X265 Rzerox)-1.mp4.part2",
		"C:\\Users\\pratersm\\AppData\\Local\\Temp\\s3sync3030611028\\Battlestar Galactica (2003)  S04e19e20  Daybreak (1080P Bluray X265 Rzerox)-1.mp4.part3",
		"C:\\Users\\pratersm\\AppData\\Local\\Temp\\s3sync3030611028\\Battlestar Galactica (2003)  S04e19e20  Daybreak (1080P Bluray X265 Rzerox)-1.mp4.part4",
		"C:\\Users\\pratersm\\AppData\\Local\\Temp\\s3sync3030611028\\Battlestar Galactica (2003)  S04e19e20  Daybreak (1080P Bluray X265 Rzerox)-1.mp4.part5",
		"C:\\Users\\pratersm\\AppData\\Local\\Temp\\s3sync3030611028\\Battlestar Galactica (2003)  S04e19e20  Daybreak (1080P Bluray X265 Rzerox)-1.mp4.part6",
	}
	err := app.recordParts(1, parts)
	if err != nil {
		t.Fatal(err)
	}

}

func TestPutObject(t *testing.T) {
	Init()
	obj := "X:\\shows\\Battlestar Galactica (2004)\\Season 4\\Battlestar Galactica (2003)  S04e19e20  Daybreak (1080P Bluray X265 Rzerox)-1.mp4"
	err := app.updateRecord(obj, 123456)
	if err != nil {
		t.Fatal(err)
	}

	err = app.putObject(context.Background(), obj, false)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpdateUploadStatusPart(t *testing.T) {
	Init()
	part := "C:\\Users\\pratersm\\AppData\\Local\\Temp\\s3sync3030611028\\Battlestar Galactica (2003)  S04e19e20  Daybreak (1080P Bluray X265 Rzerox)-1.mp4.part0"
	err := app.updateUploadStatusPart(part)
	if err != nil {
		t.Fatal(err)
	}
}

func getAwsClient(ctx context.Context) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg)
	return client, nil
}

func TestRecordExists(t *testing.T) {
	Init()
	exists, err := app.recordExists("C:\\Users\\pratersm\\Documents\\WebSites\\yci-www\\fa\\less\\fixed-width.less", 1480701261)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("exists = %t", exists)
	t.Fatal()

}

func TestUploadDiffs(t *testing.T) {
	Init()
	res, err := app.GetUploadList()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res)
	ctx := context.Background()
	err = app.UploadDiffs(ctx, res, false)
	if err != nil {
		t.Fatal(err)
	}

}
