package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	fema_compute "github.com/HydrologicEngineeringCenter/go-fema-consequences/compute"
	"github.com/HydrologicEngineeringCenter/go-fema-consequences/config"
	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Config holds all runtime configuration provided via environment variables
type AWSConfig struct {
	AWSS3Endpoint       string `envconfig:"AWS_S3_ENDPOINT"`
	AWSS3Region         string `envconfig:"AWS_S3_REGION"`
	AWSS3DisableSSL     bool   `envconfig:"AWS_S3_DISABLE_SSL"`
	AWSS3ForcePathStyle bool   `envconfig:"AWS_S3_FORCE_PATH_STYLE"`
	AWSS3Bucket         string `envconfig:"AWS_S3_BUCKET"`
	AWSS3Prefix         string `envconfig:"AWS_S3_PREFIX"`
}

func main() {
	fmt.Println("Launching fema_consequences")
	var cfg AWSConfig
	if err := envconfig.Process("fema_consequences", &cfg); err != nil {
		log.Fatal(err.Error())
	}

	port:=os.Getenv("PORT")
	if port==""{
		port="8000"
	}
	// This should probably move elsewhere
	awsConfig := aws.NewConfig().WithRegion(cfg.AWSS3Region)
	// Used for "minio" during development
	awsConfig.WithDisableSSL(cfg.AWSS3DisableSSL)
	awsConfig.WithS3ForcePathStyle(cfg.AWSS3ForcePathStyle)
	if cfg.AWSS3Endpoint != "" {
		awsConfig.WithEndpoint(cfg.AWSS3Endpoint)
	}
	newSession, err1 := session.NewSession(awsConfig)
	if err1 != nil {
		fmt.Println(err1)
	}
	s3c := s3.New(newSession)

	e := echo.New()
	e.Use(
		middleware.Logger(),
		middleware.Recover(),
		//middleware.CORS(),
		middleware.GzipWithConfig(middleware.GzipConfig{Level: 5}),
	)

	// Public Routes
	public := e.Group("")

	// Private Routes
	/*private := e.Group("")
	if cfg.SkipJWT == true {
		private.Use(middleware.MockIsLoggedIn)
	} else {
		private.Use(middleware.JWT, middleware.IsLoggedIn)
	}*/

	// Public Routes
	// NOTE: ALL GET REQUESTS ARE ALLOWED WITHOUT AUTHENTICATION USING JWTConfig Skipper. See appconfig/jwt.go
	public.GET("fema-consequences", func(c echo.Context) error {
		return c.String(http.StatusOK, "fema-consequences-api v0.0.1") //should probably have this pick up from an env variable for version info.
	})
	public.GET("fema-consequences/events", func(c echo.Context) error {
		resp, err := s3c.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: &cfg.AWSS3Bucket, Prefix:&cfg.AWSS3Prefix})
		var list string
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case s3.ErrCodeNoSuchBucket:
					fmt.Println(s3.ErrCodeNoSuchBucket, aerr.Error())
				default:
					fmt.Println(aerr.Error())
				}
			} else {
				fmt.Println(err.Error())
			}
			return c.String(http.StatusBadRequest, "something bad happened.")
		}
		for _, item := range resp.Contents {
			path := *item.Key
			fmt.Println(path)
			if len(path) > 11 {
				fmt.Println(path)
				if path[len(path)-11:] == "eventconfig" {
					list += path + "\n"
				}
			}
		}
		return c.String(http.StatusOK, list)
	})
	public.POST("fema-consequences/compute", func(c echo.Context) error {
		var i config.Config
		if err := c.Bind(&i); err != nil {
			return c.String(http.StatusBadRequest, "Invalid Input")
		}
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		c.Response().WriteHeader(http.StatusOK)
		compute, err := fema_compute.Init(i)
		if err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
		compute.Compute() //compute and write to temp directory
		//move from temp to s3.
		if compute.OutputFolderPath != "" {
			parts := strings.Split(compute.TempFileOutput, "/")
			fname := parts[len(parts)-1]
			if i.Ot == "shp" {
				tmp := compute.TempFileOutput
				//here we have shapefiles.
				extensions := make([]string, 4)
				extensions[0] = ".shp"
				extensions[1] = ".shx"
				extensions[2] = ".dbf"
				extensions[3] = ".prj"
				for _, ext := range extensions {
					fname = fname[:len(fname)-4]
					fname = fname + ext
					tmp = tmp[:len(tmp)-4]
					tmp = tmp + ext
					writeToS3(tmp, compute.OutputFolderPath+"/"+fname, cfg, s3c)
				}
			} else {
				writeToS3(compute.TempFileOutput, compute.OutputFolderPath+"/"+fname, cfg, s3c)
			}

		}
		return c.String(http.StatusOK, "Compute Complete")
	})
	log.Print("starting fema-consequences server")
	log.Fatal(http.ListenAndServe(":"+port, e))
}
func writeToS3(localpath string, s3Path string, cfg AWSConfig, s3c *s3.S3) (string, error) {
	//read in the output file.
	b, err := ioutil.ReadFile(localpath)
	reader := bytes.NewReader(b)
	input := &s3.PutObjectInput{
		Bucket:        &cfg.AWSS3Bucket,
		Body:          reader,
		ContentLength: aws.Int64(int64(reader.Len())),
		Key:           &s3Path,
	}
	s3output, err := s3c.PutObject(input)
	if err != nil {
		return "", err
	}
	//fmt.Print(s3output)
	err = os.Remove(localpath)
	if err != nil {
		log.Fatal(err)
	}
	return *s3output.ETag, err
}
