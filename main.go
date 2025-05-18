package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-co-op/gocron/v2"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func (app *application) scheduler() error {
	s, err := gocron.NewScheduler(
		gocron.WithGlobalJobOptions(
			gocron.WithSingletonMode(
				gocron.LimitModeReschedule)),
	)
	if err != nil {
		return err
	}

	app.jobScheduler = s // Store the scheduler in the application struct

	// Get the interval from config as a string (e.g., "10s" or "15m")
	intervalStr := viper.GetString("interval")
	if intervalStr == "" {
		intervalStr = "15m" // Default to 15 minutes if not set
	}

	// Parse the duration string
	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		slog.Error("invalid duration format in config", "error", err, "value", intervalStr)
		interval = 15 * time.Minute // Fallback to 15 minutes if parsing fails
	}

	_, err = s.NewJob(
		gocron.DurationJob(interval),
		gocron.NewTask(func() {
			app.upload()
		}),
		gocron.WithStartAt(gocron.WithStartImmediately()),
	)
	if err != nil {
		return err
	}

	s.Start()
	return nil
}

type application struct {
	jobScheduler gocron.Scheduler // master scheduler
	client       *s3.Client
}

func newApplication() (*application, error) {

	// Parse command line flags
	pflag.String("sources", "", "Source directory to backup")
	pflag.String("bucket", "", "S3 bucket name")
	pflag.String("interval", "1d", "S3 bucket name")
	pflag.Parse()

	viper.SetEnvPrefix("S3BACKUP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.BindPFlag("sources", pflag.Lookup("sources"))
	viper.BindPFlag("bucket", pflag.Lookup("bucket"))

	cred := credentials.NewStaticCredentialsProvider(
		os.Getenv("EXO_ACCESS_KEY_ID"),
		os.Getenv("EXO_SECRET_ACCESS_KEY"),
		"",
	)

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(os.Getenv("EXO_REGION")),
		config.WithCredentialsProvider(cred),
	)
	if err != nil {
		return nil, err
	}

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String("https://sos-ch-dk-2.exo.io")
		o.Region = "ch-dk-2"
		o.UsePathStyle = false
	})
	return &application{
		client: s3Client,
	}, nil
}

func (app *application) upload() error {
	for _, sourceDir := range strings.Split(viper.GetString("sources"), ",") {
		err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip directories
			if info.IsDir() {
				return nil
			}

			// Get the relative path from the source directory
			relPath, err := filepath.Rel(sourceDir, path)
			if err != nil {
				return fmt.Errorf("failed to get relative path for %s: %v", path, err)
			}

			// Convert path separators to forward slashes for S3
			s3Key := filepath.ToSlash(relPath)

			// Open the file
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open file %s: %v", path, err)
			}
			defer file.Close()

			// Upload to S3
			_, err = app.client.PutObject(context.TODO(), &s3.PutObjectInput{
				Bucket: aws.String(viper.GetString("bucket")),
				Key:    aws.String(fmt.Sprintf("%s/%s", filepath.Base(sourceDir), s3Key)),
				Body:   file,
			})
			if err != nil {
				return fmt.Errorf("failed to upload %s: %v", path, err)
			}

			fmt.Printf("Uploaded: %s\n", s3Key)
			return nil
		})

		if err != nil {
			log.Fatalf("Error during backup: %v", err)
		}
	}
	fmt.Println("Backup completed successfully!")

	return nil
}

func main() {

	app, _ := newApplication()

	app.scheduler()

	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan

}
