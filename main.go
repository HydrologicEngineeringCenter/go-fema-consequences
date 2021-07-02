package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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
type EventConfigStateObserver struct {
	eventlist map[string]struct{}
}

func main() {
	log.Println("Launching fema_consequences")
	var cfg AWSConfig
	if err := envconfig.Process("fema_consequences", &cfg); err != nil {
		log.Fatal(err.Error())
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	polldur := os.Getenv("POLLDURATION")
	if polldur == "" {
		polldur = "100"
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
	// polling station//
	eventlist := make(map[string]struct{})
	observer := EventConfigStateObserver{eventlist: eventlist}
	pd, err := strconv.Atoi(polldur)
	if err != nil {
		log.Fatal(err)
	}
	go func(o EventConfigStateObserver, cfg AWSConfig, s3c *s3.S3) {
		for {
			time.Sleep(time.Duration(pd) * time.Second)
			log.Println("Polling for .eventConfigs on " + cfg.AWSS3Bucket)
			i, s := listS3Objects(cfg, s3c)
			if i != http.StatusOK {
				panic("Status Was NOT ok!")
			}
			events := strings.Split(s, "\n")
			for _, e := range events {
				//fmt.Println("Looking for " + e)
				if e == "" {
					break
				}
				_, ok := observer.eventlist[e]
				if !ok {

					c, err := readFromS3(e, cfg, s3c)
					if err != nil {
						log.Println(err)
						observer.eventlist[e] = struct{}{}
					}
					//check if the config file contains a result that already exists?
					//outputfilepath :=
					dostuff(c, e, cfg, s3c)
					log.Printf("computed %s\n", e)
					observer.eventlist[e] = struct{}{}
				}
			}
		}
	}(observer, cfg, s3c)

	// end polling station //
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
		i, s := listS3Objects(cfg, s3c) //200 is status ok.
		return c.String(i, s)
	})
	public.POST("fema-consequences/compute", func(c echo.Context) error {
		var i config.Config
		if err := c.Bind(&i); err != nil {
			return c.String(http.StatusBadRequest, "Invalid Input")
		}
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		c.Response().WriteHeader(http.StatusOK)
		i2, s := dostuff(i, "", cfg, s3c)
		return c.String(i2, s)
	})

	log.Print("starting fema-consequences server")
	log.Fatal(http.ListenAndServe(":"+port, e))
}
func dostuff(i config.Config, fp string, cfg AWSConfig, s3c *s3.S3) (int, string) {
	compute, err := fema_compute.Init(i)
	if err != nil {
		//write the results to fp
		if fp != "" {
			//this is a key to an s3 bucket
			parts := strings.Split(fp, ".")
			fp = strings.Replace(fp, parts[len(parts)-1], "configHASERRORS", -1)
			//write to a temp directory.
			ofp := "/app/working/" + filepath.Base(fp)
			f, ferr := os.Create(ofp)
			if ferr != nil {
				err = errors.New(err.Error() + "\n" + ferr.Error())
			} else {
				f.WriteString(err.Error())
				f.Close()
				writeToS3(ofp, fp, cfg, s3c)
			}
		}
		return http.StatusBadRequest, err.Error()
	}
	compute.Compute() //compute and write to temp directory
	//move from temp to s3.
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
			if compute.OutputFolderPath == "" {
				writeToS3(tmp, cfg.AWSS3Prefix+"/"+fname, cfg, s3c)
			} else {
				writeToS3(tmp, cfg.AWSS3Prefix+"/"+compute.OutputFolderPath+"/"+fname, cfg, s3c)
			}
		}
	} else {
		if compute.OutputFolderPath == "" {
			writeToS3(compute.TempFileOutput, cfg.AWSS3Prefix+"/"+fname, cfg, s3c)
		} else {
			writeToS3(compute.TempFileOutput, cfg.AWSS3Prefix+"/"+compute.OutputFolderPath+"/"+fname, cfg, s3c)
		}

	}

	return http.StatusOK, "Compute Complete"
}
func writeToS3(localpath string, s3Path string, cfg AWSConfig, s3c *s3.S3) (string, error) {
	//read in the output file.
	log.Println("Writing " + localpath + " to s3 at " + s3Path)
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
func readFromS3(key string, cfg AWSConfig, s3c *s3.S3) (config.Config, error) {
	//fmt.Println("tryina read " + key)
	n, err := s3c.GetObject(&s3.GetObjectInput{
		Bucket: &cfg.AWSS3Bucket,
		Key:    &key,
	})
	if err != nil {
		log.Fatal(err)
	}
	b, err := ioutil.ReadAll(n.Body)
	if err != nil {
		log.Fatal(err)
	}
	c := config.Config{}
	json.Unmarshal(b, &c)
	return c, nil
}
func listS3Objects(cfg AWSConfig, s3c *s3.S3) (int, string) {
	resp, err := s3c.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: &cfg.AWSS3Bucket, Prefix: &cfg.AWSS3Prefix})
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
		return http.StatusBadRequest, "something bad happened."
	}
	for _, item := range resp.Contents {
		path := *item.Key
		//fmt.Println(path)
		if len(path) > 11 {
			if path[len(path)-11:] == "eventconfig" {
				fmt.Println(path)
				list += path + "\n"
			}
		}
	}
	return http.StatusOK, list
}
