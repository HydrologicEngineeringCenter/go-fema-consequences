package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	fema_compute "github.com/HydrologicEngineeringCenter/go-fema-consequences/compute"
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
		log.Println(err1)
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
			currentlist := make(map[string]struct{})
			log.Println("Polling for .tif files on " + cfg.AWSS3Bucket)
			i, s := listS3TifObjects(cfg, s3c, cfg.AWSS3Prefix)
			if i != http.StatusOK {
				panic("Status Was NOT ok!")
			}
			events := strings.Split(s, "\n")
			for _, e := range events {
				if e == "" {
					break
				}
				currentlist[e] = struct{}{} //store current config files to make sure old ones are discarded after they are deleted
				_, ok := observer.eventlist[e]
				if !ok {
					computeFromTif(e, cfg, s3c)
					log.Printf("computed %s\n", e)
					observer.eventlist[e] = struct{}{}
				}
			}
			observer.eventlist = currentlist
		}
	}(observer, cfg, s3c)

	// end polling station //
	// Public Routes
	public := e.Group("")
	// Public Routes
	// NOTE: ALL GET REQUESTS ARE ALLOWED WITHOUT AUTHENTICATION USING JWTConfig Skipper. See appconfig/jwt.go
	public.GET("fema-consequences", func(c echo.Context) error {
		return c.String(http.StatusOK, "fema-consequences-api v0.0.1") //should probably have this pick up from an env variable for version info.
	})
	public.GET("fema-consequences/events", func(c echo.Context) error {
		i, s := listS3TifObjects(cfg, s3c, cfg.AWSS3Prefix) //200 is status ok.
		return c.String(i, s)
	})
	log.Print("starting fema-consequences server")
	log.Fatal(http.ListenAndServe(":"+port, e))
}
func computeFromTif(fp string, cfg AWSConfig, s3c *s3.S3) (int, string) {
	compute, err := fema_compute.Init(fp, "results")
	if err != nil {
		//write the results to fp
		if fp != "" {
			//this is a key to a tif file on an s3 bucket
			writeErrors(fp, cfg, s3c, err, "computeHASERRORS")
		}
		return http.StatusBadRequest, err.Error()
	}
	//prepare for move from temp to s3.
	parts := strings.Split(compute.TempFileOutput, "/")
	fname := parts[len(parts)-1]
	//check if it has been computed before hand.
	skipCompute := false
	if compute.OutputFolderPath == "" {
		if exists(cfg, s3c, cfg.AWSS3Prefix+"/"+fname) {
			//bad news bears... skipperooo
			skipCompute = true
		}
	} else {
		if exists(cfg, s3c, cfg.AWSS3Prefix+"/"+compute.OutputFolderPath+"/"+fname) {
			skipCompute = true
		}
	}
	if skipCompute {
		writeErrors(fp, cfg, s3c, errors.New("Previous Output Detected, Skipping Compute"), "PREVIOUSLYComputed")
		return http.StatusConflict, "Previous Output Detected In Directory, Skipping Compute"
	}
	compute.Compute() //compute and write to temp directory
	/*
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
	*/
	if compute.OutputFolderPath == "" {
		writeToS3(compute.TempFileOutput, cfg.AWSS3Prefix+"/"+fname, cfg, s3c)
	} else {
		writeToS3(compute.TempFileOutput, cfg.AWSS3Prefix+"/"+compute.OutputFolderPath+"/"+fname, cfg, s3c)
	}

	//}

	return http.StatusOK, "Compute Complete"
}
func writeToS3(localpath string, s3Path string, cfg AWSConfig, s3c *s3.S3) (string, error) {
	//read in the output file.
	if localpath == "" {
		return "", errors.New("Local path was blank!")
	}
	if localpath == "/app/" {
		return "", errors.New("Local path was /app/!")
	}
	if !exists(cfg, s3c, s3Path) {
		log.Println("Writing " + localpath + " to s3 at " + s3Path)
		b, err := ioutil.ReadFile(localpath)
		if err != nil {
			return "", err
		}
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
	} else {
		log.Println("File already exists, cleaning up loacal path")
		err := os.Remove(localpath)
		if err != nil {
			log.Fatal(err)
		}
		return "", errors.New("File already exists")
	}
}
func listS3TifObjects(cfg AWSConfig, s3c *s3.S3, prefix string) (int, string) {
	resp, err := s3c.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: &cfg.AWSS3Bucket, Prefix: &prefix})
	var list string
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				log.Println(s3.ErrCodeNoSuchBucket, aerr.Error())
			default:
				log.Println(aerr.Error())
			}
		} else {
			log.Println(err.Error())
		}
		return http.StatusBadRequest, "something bad happened."
	}
	for _, item := range resp.Contents {
		path := *item.Key
		if len(path) > 11 {
			if path[len(path)-4:] == ".tif" {
				log.Println(path)
				list += path + "\n"
			}
		}
	}
	return http.StatusOK, list
}
func exists(cfg AWSConfig, s3c *s3.S3, key string) bool {
	_, err := s3c.HeadObject(&s3.HeadObjectInput{Bucket: &cfg.AWSS3Bucket, Key: &key})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				log.Println(s3.ErrCodeNoSuchBucket, aerr.Error())
			default:
				log.Println(aerr.Error())
			}
		} else {
			log.Println(err.Error())
		}
		return false
	}
	return true
}
func writeErrors(fp string, cfg AWSConfig, s3c *s3.S3, err error, extension string) {
	parts := strings.Split(fp, ".")
	fp = strings.Replace(fp, parts[len(parts)-1], extension, -1)
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
