package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

var (
	addr     string
	dbName   string
	username string
	password string
	stdin    *os.File
)

func main() {
	rootCmd := configureRootCommand()
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err.Error())
	}
}

func configureRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "handler-influx-db",
		Short: "an influx-db handler built for use with sensu",
		RunE:  run,
	}

	cmd.Flags().StringVarP(&addr,
		"addr",
		"a",
		"",
		"The address of the influx-db server")

	cmd.Flags().StringVarP(&dbName,
		"db-name",
		"d",
		"",
		"The influx-db to send metrics to")

	cmd.Flags().StringVarP(&username,
		"username",
		"u",
		"",
		"The username for the given db")

	cmd.Flags().StringVarP(&password,
		"password",
		"p",
		"",
		"The password for the given db")

	_ = cmd.MarkFlagRequired("addr")
	_ = cmd.MarkFlagRequired("db-name")
	_ = cmd.MarkFlagRequired("username")
	_ = cmd.MarkFlagRequired("password")

	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		_ = cmd.Help()
		return errors.New("invalid argument(s) received")
	}

	if stdin == nil {
		stdin = os.Stdin
	}

	eventJSON, err := ioutil.ReadAll(stdin)
	if err != nil {
		return fmt.Errorf("failed to read stdin: %s", err.Error())
	}

	event := &types.Event{}
	err = json.Unmarshal(eventJSON, event)
	if err != nil {
		return fmt.Errorf("failed to unmarshal stdin data: %s", err.Error())
	}

	if err = event.Validate(); err != nil {
		return fmt.Errorf("failed to validate event: %s", err.Error())
	}

	if event.Metrics == nil {
		return fmt.Errorf("event does not contain metrics")
	}

	return sendMetrics(event)
}

func sendMetrics(event *types.Event) error {
	var pt *client.Point
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     addr,
		Username: username,
		Password: password,
	})
	if err != nil {
		return err
	}
	defer c.Close()

	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  dbName,
		Precision: "s",
	})
	if err != nil {
		return err
	}

	for _, point := range event.Metrics.Points {
		var tagKey string
		nameField := strings.Split(point.Name, ".")
		name := nameField[0]
		if len(nameField) > 1 {
			tagKey = strings.Join(nameField[1:], ".")
		} else {
			tagKey = "value"
		}
		fields := map[string]interface{}{tagKey: point.Value}
		timestamp := time.Unix(0, point.Timestamp)
		tags := make(map[string]string)
		for _, tag := range point.Tags {
			tags[tag.Name] = tag.Value
		}

		pt, err = client.NewPoint(name, tags, fields, timestamp)
		if err != nil {
			return err
		}
		bp.AddPoint(pt)
	}

	if err = c.Write(bp); err != nil {
		return err
	}

	return c.Close()
}
